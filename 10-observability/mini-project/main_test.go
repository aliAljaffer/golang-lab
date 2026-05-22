//go:build exercise

package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// ---- helpers ---------------------------------------------------------------

func sign(secret, body []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// testServer builds a server with sane defaults plus an in-memory span
// recorder so tests can inspect spans.
func testServer(t *testing.T) (http.Handler, *Metrics, *bytes.Buffer, *tracetest.InMemoryExporter) {
	t.Helper()
	m := NewMetrics()
	var logbuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logbuf, nil))

	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	tracer := tp.Tracer("test")
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	jobs := map[string]Job{
		"echo": {Command: []string{"echo", "hello"}},
		"fail": {Command: []string{"sh", "-c", "exit 7"}},
	}

	h := newServer(ServerOpts{
		Secret:    []byte("swordfish"),
		Jobs:      jobs,
		Logger:    logger,
		Metrics:   m,
		Tracer:    tracer,
		MaxOutput: defaultMaxOutput,
	})
	return h, m, &logbuf, exp
}

// counter looks up a counter by labels and returns its current value.
func counter(t *testing.T, vec *prometheus.CounterVec, labels ...string) float64 {
	t.Helper()
	c, err := vec.GetMetricWithLabelValues(labels...)
	if err != nil {
		t.Fatalf("GetMetricWithLabelValues: %v", err)
	}
	var m dto.Metric
	if err := c.Write(&m); err != nil {
		t.Fatalf("Write: %v", err)
	}
	return m.GetCounter().GetValue()
}

// gauge reads the current value of a Gauge.
func gauge(t *testing.T, g prometheus.Gauge) float64 {
	t.Helper()
	var m dto.Metric
	if err := g.Write(&m); err != nil {
		t.Fatalf("Write: %v", err)
	}
	return m.GetGauge().GetValue()
}

// post sends a POST to /webhook with the given body, signed with secret.
func post(h http.Handler, secret, body []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set("X-Hub-Signature-256", sign(secret, body))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

// ---- VerifyHMAC ------------------------------------------------------------

func TestVerifyHMAC_ValidSignature(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	secret := []byte("s3cret")
	body := []byte(`{"job":"echo"}`)
	if !VerifyHMAC(context.Background(), tracer, secret, body, sign(secret, body)) {
		t.Error("VerifyHMAC valid = false, want true")
	}
}

func TestVerifyHMAC_BadSignature(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	if VerifyHMAC(context.Background(), tracer, []byte("s"), []byte("x"), "sha256=deadbeef") {
		t.Error("VerifyHMAC bad = true, want false")
	}
}

func TestVerifyHMAC_MissingPrefix(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	if VerifyHMAC(context.Background(), tracer, []byte("s"), []byte("x"), "deadbeef") {
		t.Error("VerifyHMAC without 'sha256=' prefix = true, want false")
	}
}

// ---- RunJob ----------------------------------------------------------------

func TestRunJob_Success(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	exit, out, err := RunJob(context.Background(), tracer, Job{Command: []string{"echo", "hello"}}, 100)
	if err != nil {
		t.Fatalf("RunJob: %v", err)
	}
	if exit != 0 {
		t.Errorf("exit = %d, want 0", exit)
	}
	if !strings.Contains(out, "hello") {
		t.Errorf("output = %q, want contain 'hello'", out)
	}
}

func TestRunJob_NonZeroExit(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	exit, _, err := RunJob(context.Background(), tracer, Job{Command: []string{"sh", "-c", "exit 7"}}, 100)
	if err != nil {
		t.Fatalf("RunJob: %v", err)
	}
	if exit != 7 {
		t.Errorf("exit = %d, want 7", exit)
	}
}

func TestRunJob_OutputTruncated(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	// Produce 1000 bytes; cap at 10.
	_, out, err := RunJob(context.Background(), tracer,
		Job{Command: []string{"sh", "-c", "yes a | head -c 1000"}}, 10)
	if err != nil {
		t.Fatalf("RunJob: %v", err)
	}
	if len(out) != 10 {
		t.Errorf("len(out) = %d, want 10 (truncation)", len(out))
	}
}

// ---- Server: end-to-end ----------------------------------------------------

func TestServer_HappyPath(t *testing.T) {
	h, m, _, _ := testServer(t)
	rr := post(h, []byte("swordfish"), []byte(`{"job":"echo"}`))
	if rr.Code != 200 {
		t.Fatalf("status = %d, want 200; body=%q", rr.Code, rr.Body.String())
	}

	var resp WebhookResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v (body=%q)", err, rr.Body.String())
	}
	if resp.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", resp.ExitCode)
	}

	if got := counter(t, m.Requests, "POST", "/webhook", "200"); got != 1 {
		t.Errorf("http_requests_total{status=200} = %f, want 1", got)
	}
	if got := counter(t, m.Jobs, "echo", "ok"); got != 1 {
		t.Errorf("webhook_jobs_total{job=echo,result=ok} = %f, want 1", got)
	}
}

func TestServer_BadSignatureReturns401AndCounts(t *testing.T) {
	h, m, _, _ := testServer(t)
	body := []byte(`{"job":"echo"}`)
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set("X-Hub-Signature-256", "sha256=deadbeef")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != 401 {
		t.Fatalf("status = %d, want 401", rr.Code)
	}
	if got := counter(t, m.Requests, "POST", "/webhook", "401"); got != 1 {
		t.Errorf("http_requests_total{status=401} = %f, want 1", got)
	}
}

func TestServer_UnknownJobReturns404AndCountsUnknown(t *testing.T) {
	h, m, _, _ := testServer(t)
	rr := post(h, []byte("swordfish"), []byte(`{"job":"nope"}`))
	if rr.Code != 404 {
		t.Fatalf("status = %d, want 404", rr.Code)
	}
	if got := counter(t, m.Jobs, "nope", "unknown"); got != 1 {
		t.Errorf("webhook_jobs_total{job=nope,result=unknown} = %f, want 1", got)
	}
}

func TestServer_FailingJobCountsFail(t *testing.T) {
	h, m, _, _ := testServer(t)
	rr := post(h, []byte("swordfish"), []byte(`{"job":"fail"}`))
	if rr.Code != 200 {
		t.Fatalf("status = %d, want 200 (exit != 0 is still a successful run); body=%q", rr.Code, rr.Body.String())
	}
	if got := counter(t, m.Jobs, "fail", "fail"); got != 1 {
		t.Errorf("webhook_jobs_total{job=fail,result=fail} = %f, want 1", got)
	}
}

// ---- /metrics endpoint -----------------------------------------------------

func TestServer_MetricsEndpointExposesSeries(t *testing.T) {
	h, _, _, _ := testServer(t)
	// Drive one request so non-zero series exist.
	_ = post(h, []byte("swordfish"), []byte(`{"job":"echo"}`))

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("/metrics status = %d, want 200", rr.Code)
	}
	body := rr.Body.String()
	for _, want := range []string{
		"http_requests_total",
		"http_request_seconds",
		"http_in_flight",
		"webhook_jobs_total",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("/metrics body missing %q", want)
		}
	}
}

// ---- Logging ---------------------------------------------------------------

func TestServer_RequestLogIncludesRequestID(t *testing.T) {
	h, _, logbuf, _ := testServer(t)
	_ = post(h, []byte("swordfish"), []byte(`{"job":"echo"}`))

	out := logbuf.String()
	if !strings.Contains(out, `"request_id"`) {
		t.Errorf("log output missing request_id; got: %s", out)
	}
	if !strings.Contains(out, `"trace_id"`) {
		t.Errorf("log output missing trace_id; got: %s", out)
	}
	if !strings.Contains(out, `"request.start"`) || !strings.Contains(out, `"request.end"`) {
		t.Errorf("log output missing request.start / request.end events; got: %s", out)
	}
}

// ---- Tracing ---------------------------------------------------------------

func TestServer_SpansCreated(t *testing.T) {
	h, _, _, exp := testServer(t)
	_ = post(h, []byte("swordfish"), []byte(`{"job":"echo"}`))

	spans := exp.GetSpans()
	names := map[string]trace.SpanContext{}
	for _, s := range spans {
		names[s.Name] = s.SpanContext
	}

	for _, want := range []string{"POST /webhook", "verify-hmac", "run-job"} {
		if _, ok := names[want]; !ok {
			t.Errorf("missing span %q; got: %v", want, spanNames(spans))
		}
	}
}

func spanNames(spans tracetest.SpanStubs) []string {
	out := make([]string, 0, len(spans))
	for _, s := range spans {
		out = append(out, s.Name)
	}
	return out
}

// ---- In-flight gauge -------------------------------------------------------

// Confirm the in-flight gauge returns to 0 after the request completes. (We
// can't easily catch the intra-request peak from a synchronous test without
// blocking the handler; verifying the deferred Dec() runs is the next-best
// check.)
func TestServer_InFlightReturnsToZero(t *testing.T) {
	h, m, _, _ := testServer(t)
	if got := gauge(t, m.InFlight); got != 0 {
		t.Fatalf("InFlight before request = %f, want 0", got)
	}
	_ = post(h, []byte("swordfish"), []byte(`{"job":"echo"}`))
	if got := gauge(t, m.InFlight); got != 0 {
		t.Errorf("InFlight after request = %f, want 0", got)
	}
}

// Compile-time guard: make sure we don't accidentally drop the io import we
// rely on elsewhere in the package.
var _ = io.Discard
