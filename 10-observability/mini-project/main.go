// webhook-runner-instrumented — the webhook-runner from 04-http-servers
// with logs, metrics, and traces wired in.
//
// What changes compared to the original:
//
//   - Every request gets a request_id (set by middleware, stashed on ctx,
//     pulled into every slog line via [[slog-context pattern from 02]]).
//   - `/metrics` exposes:
//       http_requests_total{method,path,status}
//       http_request_seconds{method,path}  (histogram)
//       http_in_flight                      (gauge)
//       webhook_jobs_total{job,result}      (counter, result=ok|fail|unknown)
//   - OTel spans: server span (root) → verify-hmac → run-job. Spans share
//     a trace_id; that trace_id is added to every slog line within the request.
//
// Testable surface (top of file). Wiring is at the bottom.
//
// Run:
//
//	WEBHOOK_SECRET=swordfish go run ./10-observability/mini-project
//	# in another terminal:
//	curl -s localhost:8080/metrics | grep webhook_
package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

const (
	defaultMaxBody   = 1 << 20
	defaultMaxOutput = 4 << 10
)

// Job is a single command to run.
type Job struct {
	Command []string
}

// WebhookRequest is the JSON body of POST /webhook.
type WebhookRequest struct {
	Job string `json:"job"`
}

// WebhookResponse is the success reply.
type WebhookResponse struct {
	Job      string `json:"job"`
	ExitCode int    `json:"exit_code"`
	Output   string `json:"output"`
}

// Metrics bundles all the Prom metrics this server exposes. Holding them on
// a struct (rather than package-level promauto) makes the server testable —
// each test can pass a fresh `*prometheus.Registry` and inspect it in isolation.
type Metrics struct {
	Requests *prometheus.CounterVec
	Duration *prometheus.HistogramVec
	InFlight prometheus.Gauge
	Jobs     *prometheus.CounterVec

	reg *prometheus.Registry
}

// NewMetrics constructs a fresh registry, registers all server metrics against
// it, and returns the bundle. The bundle holds a reference to the registry so
// newServer can plumb it to promhttp.HandlerFor.
func NewMetrics() *Metrics {
	reg := prometheus.NewRegistry()
	m := &Metrics{
		Requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "HTTP requests handled, by method/path/status.",
		}, []string{"method", "path", "status"}),
		Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_seconds",
			Help:    "HTTP request latency in seconds.",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 14),
		}, []string{"method", "path"}),
		InFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "http_in_flight",
			Help: "Number of HTTP requests currently being served.",
		}),
		Jobs: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "webhook_jobs_total",
			Help: "Webhook jobs executed, by job/result (ok|fail|unknown).",
		}, []string{"job", "result"}),
		reg: reg,
	}
	reg.MustRegister(m.Requests, m.Duration, m.InFlight, m.Jobs)
	return m
}

// ServerOpts is what the constructor needs. All fields are injectable so tests
// can swap loggers, registries, and tracers.
type ServerOpts struct {
	Secret    []byte
	Jobs      map[string]Job
	MaxOutput int
	Logger    *slog.Logger
	Metrics   *Metrics
	Tracer    trace.Tracer
}

// ctxKey is the type-safe key for stashing a per-request logger on context.
type ctxKey struct{}

func withLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

func loggerFrom(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

// statusRecorder wraps a ResponseWriter to remember the status code that was
// written. Without this, you can't tell what status the handler wrote from
// inside middleware (ResponseWriter has Write+WriteHeader but no getter).
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// observability is the middleware that does all the cross-cutting work:
// generate request_id, build a logger with request_id + trace_id, increment
// in-flight, time the request, increment counters.
func observability(opts ServerOpts) func(http.Handler) http.Handler {
	var reqCounter atomic.Uint64
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := strconv.FormatUint(reqCounter.Add(1), 10)
			start := time.Now()

			// Start the server span. ctx now carries the span.
			ctx, span := opts.Tracer.Start(r.Context(), r.Method+" "+r.URL.Path)
			defer span.End()
			span.SetAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.path", r.URL.Path),
				attribute.String("request_id", reqID),
			)

			// Per-request logger: request_id + trace_id (if available).
			l := opts.Logger.With(slog.String("request_id", reqID))
			if sc := span.SpanContext(); sc.HasTraceID() {
				l = l.With(slog.String("trace_id", sc.TraceID().String()))
			}
			ctx = withLogger(ctx, l)
			r = r.WithContext(ctx)

			// In-flight gauge.
			opts.Metrics.InFlight.Inc()
			defer opts.Metrics.InFlight.Dec()

			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			l.Info("request.start", slog.String("method", r.Method), slog.String("path", r.URL.Path))

			next.ServeHTTP(rec, r)

			dur := time.Since(start).Seconds()
			opts.Metrics.Duration.WithLabelValues(r.Method, r.URL.Path).Observe(dur)
			opts.Metrics.Requests.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(rec.status)).Inc()
			span.SetAttributes(attribute.Int("http.status_code", rec.status))

			l.Info("request.end",
				slog.Int("status", rec.status),
				slog.Float64("duration_s", dur),
			)
		})
	}
}

// VerifyHMAC validates a GitHub-style "sha256=<hex>" signature header.
// Creates a span so a trace shows time spent on signature verification.
func VerifyHMAC(ctx context.Context, tracer trace.Tracer, secret, body []byte, header string) bool {
	// TODO: _, span := tracer.Start(ctx, "verify-hmac")
	// TODO: defer span.End()
	// TODO: const prefix = "sha256="
	// TODO: if !strings.HasPrefix(header, prefix) { span.SetAttributes(attribute.Bool("valid", false)); return false }
	// TODO: want, err := hex.DecodeString(header[len(prefix):])
	// TODO: if err != nil { return false }
	// TODO: mac := hmac.New(sha256.New, secret); mac.Write(body); got := mac.Sum(nil)
	// TODO: ok := hmac.Equal(got, want)
	// TODO: span.SetAttributes(attribute.Bool("valid", ok))
	// TODO: return ok

	_ = ctx
	_ = tracer
	_ = strings.HasPrefix
	_ = hex.DecodeString
	_ = hmac.New
	_ = sha256.New
	_ = attribute.Bool
	return false
}

// RunJob runs j.Command. Creates a "run-job" span and records exit code as an
// attribute. Returns exit code (0 on success, >0 on process failure) and
// truncated combined stdout+stderr. err is non-nil only on spawn-time failures.
func RunJob(ctx context.Context, tracer trace.Tracer, j Job, maxOutput int) (exitCode int, output string, err error) {
	// TODO: ctx, span := tracer.Start(ctx, "run-job")
	// TODO: defer span.End()
	// TODO: span.SetAttributes(attribute.String("job.argv0", j.Command[0]))
	// TODO: if len(j.Command) == 0 { return 0, "", errors.New("empty command") }
	// TODO: cmd := exec.CommandContext(ctx, j.Command[0], j.Command[1:]...)
	// TODO: var buf bytes.Buffer; cmd.Stdout = &buf; cmd.Stderr = &buf
	// TODO: runErr := cmd.Run()
	// TODO: out := buf.String(); if len(out) > maxOutput { out = out[:maxOutput] }
	// TODO: switch e := runErr.(type) {
	// TODO: case nil:                span.SetAttributes(attribute.Int("exit.code", 0)); return 0, out, nil
	// TODO: case *exec.ExitError:    span.SetAttributes(attribute.Int("exit.code", e.ExitCode())); return e.ExitCode(), out, nil
	// TODO: default:                 return 0, out, runErr
	// TODO: }

	_ = tracer
	_ = bytes.NewBuffer
	_ = exec.CommandContext
	_ = errors.New
	return 0, "", fmt.Errorf("RunJob: not implemented")
}

// newServer wires up the full mux: instrumented /webhook + /metrics.
func newServer(opts ServerOpts) http.Handler {
	if opts.Tracer == nil {
		opts.Tracer = noop.NewTracerProvider().Tracer("noop")
	}
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}
	if opts.MaxOutput == 0 {
		opts.MaxOutput = defaultMaxOutput
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /webhook", func(w http.ResponseWriter, r *http.Request) {
		l := loggerFrom(r.Context())

		r.Body = http.MaxBytesReader(w, r.Body, defaultMaxBody)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if !VerifyHMAC(r.Context(), opts.Tracer, opts.Secret, body, r.Header.Get("X-Hub-Signature-256")) {
			l.Warn("hmac.reject")
			http.Error(w, "bad signature", http.StatusUnauthorized)
			return
		}

		var req WebhookRequest
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		job, ok := opts.Jobs[req.Job]
		if !ok {
			opts.Metrics.Jobs.WithLabelValues(req.Job, "unknown").Inc()
			l.Warn("job.unknown", slog.String("job", req.Job))
			http.Error(w, "unknown job", http.StatusNotFound)
			return
		}

		exit, out, err := RunJob(r.Context(), opts.Tracer, job, opts.MaxOutput)
		if err != nil {
			opts.Metrics.Jobs.WithLabelValues(req.Job, "fail").Inc()
			l.Error("job.spawn_error", slog.String("err", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		result := "ok"
		if exit != 0 {
			result = "fail"
		}
		opts.Metrics.Jobs.WithLabelValues(req.Job, result).Inc()
		l.Info("job.done", slog.String("job", req.Job), slog.Int("exit", exit))

		writeJSON(w, http.StatusOK, WebhookResponse{Job: req.Job, ExitCode: exit, Output: out})
	})

	// Wrap webhook routes with the observability middleware.
	wrapped := observability(opts)(mux)

	// /metrics is served without observability wrapping — we don't want to
	// count scrapes against ourselves. Add a second mux at the outer layer.
	outer := http.NewServeMux()
	outer.Handle("/metrics", promhttp.HandlerFor(opts.Metrics.registry(), promhttp.HandlerOpts{}))
	outer.Handle("/", wrapped)
	return outer
}

// registry returns the registry holding the metrics bundle. We use an
// explicit registry rather than prometheus.DefaultRegisterer so multiple
// newServer instances in tests don't collide on metric registration.
func (m *Metrics) registry() prometheus.Gatherer {
	return m.reg
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func main() {
	secret := []byte(os.Getenv("WEBHOOK_SECRET"))
	if len(secret) == 0 {
		fmt.Fprintln(os.Stderr, "WEBHOOK_SECRET env var is required")
		os.Exit(1)
	}

	m := NewMetrics()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Hard-coded job map for the demo — the YAML loading lives in 04's
	// version; observability is the lesson here.
	jobs := map[string]Job{
		"echo":      {Command: []string{"echo", "hello"}},
		"date":      {Command: []string{"date"}},
		"sleep-5ms": {Command: []string{"sleep", "0.005"}},
	}

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           newServer(ServerOpts{Secret: secret, Jobs: jobs, Logger: logger, Metrics: m}),
		ReadHeaderTimeout: 5 * time.Second,
	}
	logger.Info("listening", slog.String("addr", srv.Addr))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
