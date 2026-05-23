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
	// TODO: pull the value at requestIDKey back out as a string. ctx.Value
	//   returns interface{}; the type assertion's ok-form is how you keep
	//   "missing" and "wrong type" from panicking.
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
			// TODO: thread an ID through this request:
			//   - reuse X-Request-ID if the caller supplied one (lets a load
			//     balancer or upstream tracer keep continuity); otherwise mint
			//     via idGen.
			//   - echo it on the response header, stash it in the context so
			//     inner handlers can pull it back with RequestIDFromContext.
			//   - wrap w in statusRecorder so the "end" log line knows what
			//     status code went out — http.ResponseWriter doesn't tell you.
			//   - logger == nil is a valid case (silent middleware). Don't
			//     branch around that with a panic guard; just nil-check.

			_ = time.Now
			next.ServeHTTP(w, r)
		})
	}
}
