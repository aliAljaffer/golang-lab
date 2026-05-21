// 03-request-tracing — request-ID middleware with context propagation.
//
// You implement:
//   - WithRequestID(idGen func() string, logger *log.Logger) func(http.Handler) http.Handler
//   - RequestIDFromContext(ctx) string
//
// Behavior:
//   - If the incoming request has an X-Request-ID header, reuse it.
//     Otherwise mint one via idGen().
//   - Stash the ID in the request context so inner handlers can read it.
//   - Echo it on the response header.
//   - Log "start" and "end" lines with the ID + status + duration via `logger`.
//
// This is the foundation for distributed tracing: every log line your handler
// emits should include the request ID so you can grep across services.
package tracing

import (
	"context"
	"log"
	"net/http"
	"time"
)

type ctxKey int

const requestIDKey ctxKey = 0

// RequestIDFromContext returns the request ID stored in ctx, or "" if none.
func RequestIDFromContext(ctx context.Context) string {
	// TODO: v, _ := ctx.Value(requestIDKey).(string); return v
	return ""
}

// statusRecorder records the status code so the logger can include it.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// WithRequestID returns middleware that ensures every request has an ID and logs lifecycle.
//
// idGen is called only when the incoming request has no X-Request-ID header.
// Tests pass a deterministic generator; in main code, use a random hex generator.
//
// logger is where "start" and "end" lines go. If nil, no logging is done
// (still sets the header + context).
func WithRequestID(idGen func() string, logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: id := r.Header.Get("X-Request-ID"); if id == "" { id = idGen() }
			// TODO: w.Header().Set("X-Request-ID", id)
			// TODO: ctx := context.WithValue(r.Context(), requestIDKey, id)
			// TODO: sr := &statusRecorder{ResponseWriter: w, status: 200}
			// TODO: started := time.Now()
			// TODO: if logger != nil { logger.Printf("start id=%s method=%s path=%s", id, r.Method, r.URL.Path) }
			// TODO: next.ServeHTTP(sr, r.WithContext(ctx))
			// TODO: if logger != nil { logger.Printf("end   id=%s status=%d duration=%s", id, sr.status, time.Since(started)) }

			_ = time.Now
			next.ServeHTTP(w, r)
		})
	}
}
