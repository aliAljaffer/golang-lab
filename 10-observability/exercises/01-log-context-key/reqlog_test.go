//go:build exercise

package reqlog

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWithRequestID_RoundTrips(t *testing.T) {
	ctx := WithRequestID(context.Background(), "req-42")
	if got := RequestIDFromContext(ctx); got != "req-42" {
		t.Errorf("RequestIDFromContext = %q, want %q", got, "req-42")
	}
}

func TestRequestIDFromContext_MissingReturnsEmpty(t *testing.T) {
	if got := RequestIDFromContext(context.Background()); got != "" {
		t.Errorf("missing = %q, want \"\"", got)
	}
}

func TestMiddleware_HonoursIncomingHeader(t *testing.T) {
	var seen string
	h := Middleware(slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil)), func() string { return "GENERATED" })(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			seen = RequestIDFromContext(r.Context())
		}),
	)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "incoming-abc")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if seen != "incoming-abc" {
		t.Errorf("ctx request_id = %q, want incoming-abc (should not have regenerated)", seen)
	}
	if got := rr.Header().Get("X-Request-ID"); got != "incoming-abc" {
		t.Errorf("response X-Request-ID = %q, want incoming-abc (must echo)", got)
	}
}

func TestMiddleware_GeneratesIDWhenHeaderMissing(t *testing.T) {
	var seen string
	h := Middleware(slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil)), func() string { return "GENERATED" })(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			seen = RequestIDFromContext(r.Context())
		}),
	)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if seen != "GENERATED" {
		t.Errorf("ctx request_id = %q, want GENERATED (no incoming header)", seen)
	}
	if got := rr.Header().Get("X-Request-ID"); got != "GENERATED" {
		t.Errorf("response X-Request-ID = %q, want GENERATED", got)
	}
}

func TestMiddleware_LoggerInContextHasRequestIDPreBound(t *testing.T) {
	var buf bytes.Buffer
	base := slog.New(slog.NewJSONHandler(&buf, nil))

	h := Middleware(base, func() string { return "id-from-gen" })(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// The handler pulls the logger from ctx — it shouldn't need to
			// know the request ID exists.
			LoggerFromContext(r.Context()).Info("hello")
		}),
	)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	out := buf.String()
	if !strings.Contains(out, `"request_id":"id-from-gen"`) {
		t.Errorf("log output missing request_id; got: %s", out)
	}
	if !strings.Contains(out, `"msg":"hello"`) {
		t.Errorf("log output missing msg; got: %s", out)
	}
}
