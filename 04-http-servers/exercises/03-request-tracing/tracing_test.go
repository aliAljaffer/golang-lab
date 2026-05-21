//go:build exercise

package tracing

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newRecorder() (*log.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	return log.New(&buf, "", 0), &buf
}

func TestWithRequestID_SetsResponseHeader(t *testing.T) {
	gen := func() string { return "fixed-id-123" }
	h := WithRequestID(gen, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))

	if got := rec.Header().Get("X-Request-ID"); got != "fixed-id-123" {
		t.Errorf("X-Request-ID = %q, want %q", got, "fixed-id-123")
	}
}

func TestWithRequestID_HonorsIncomingHeader(t *testing.T) {
	gen := func() string {
		t.Error("idGen called when X-Request-ID was already set on the request")
		return "should-not-be-used"
	}
	h := WithRequestID(gen, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-ID", "from-client")
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-ID"); got != "from-client" {
		t.Errorf("X-Request-ID = %q, want %q (should reuse incoming)", got, "from-client")
	}
}

func TestWithRequestID_AvailableInContext(t *testing.T) {
	gen := func() string { return "ctx-id" }
	var seen string
	h := WithRequestID(gen, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = RequestIDFromContext(r.Context())
	}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

	if seen != "ctx-id" {
		t.Errorf("RequestIDFromContext = %q, want %q", seen, "ctx-id")
	}
}

func TestWithRequestID_LogsStartAndEnd(t *testing.T) {
	logger, buf := newRecorder()
	gen := func() string { return "log-id" }
	h := WithRequestID(gen, logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/widgets", nil))

	out := buf.String()
	if !strings.Contains(out, "start") || !strings.Contains(out, "log-id") {
		t.Errorf("logs missing start line with id:\n%s", out)
	}
	if !strings.Contains(out, "end") {
		t.Errorf("logs missing end line:\n%s", out)
	}
	if !strings.Contains(out, "418") {
		t.Errorf("end log should include status 418:\n%s", out)
	}
}

func TestRequestIDFromContext_NoIDReturnsEmpty(t *testing.T) {
	if got := RequestIDFromContext(httptest.NewRequest("GET", "/", nil).Context()); got != "" {
		t.Errorf("RequestIDFromContext on bare ctx = %q, want \"\"", got)
	}
}
