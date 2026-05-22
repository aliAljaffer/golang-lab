// Package reqlog implements an HTTP middleware that:
//
//  1. Extracts a request ID from the incoming `X-Request-ID` header, or
//     generates one if missing.
//  2. Stashes that ID on the request's `context.Context`.
//  3. Builds a `*slog.Logger` with `request_id` pre-bound and stashes that
//     on the context too.
//  4. Echoes the ID back on the response as `X-Request-ID`.
//
// Downstream handlers retrieve the logger via `LoggerFromContext(r.Context())`
// and the ID via `RequestIDFromContext(r.Context())`. Every log line they
// emit on that logger carries `request_id="..."` automatically.
//
// Why request IDs matter:
//
//	A single user-facing request can fan out across 10 services. The
//	request ID is the thread that lets you grep ALL logs for that one
//	request, in any service. Doing this consistently is half the value
//	of an observability stack.
package reqlog

import (
	"context"
	"log/slog"
	"net/http"
)

// Header is the incoming/outgoing header name. Production services often
// also honour `X-Correlation-ID` or `Request-Id` — start with one canonical.
const Header = "X-Request-ID"

// ctxKey is the type-safe key for stashing values on context.Context.
// Two separate keys: one for the ID string, one for the logger.
type ctxKey int

const (
	keyRequestID ctxKey = iota
	keyLogger
)

// WithRequestID returns ctx with id attached.
func WithRequestID(ctx context.Context, id string) context.Context {
	// TODO: return context.WithValue(ctx, keyRequestID, id)
	return ctx
}

// RequestIDFromContext returns the request ID attached to ctx, or "" if none.
func RequestIDFromContext(ctx context.Context) string {
	// TODO: if v, ok := ctx.Value(keyRequestID).(string); ok { return v }
	// TODO: return ""
	return ""
}

// WithLogger returns ctx with l attached.
func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	// TODO: return context.WithValue(ctx, keyLogger, l)
	return ctx
}

// LoggerFromContext returns the logger attached to ctx, or slog.Default()
// if none. Never returns nil — this is what makes ctx-bound loggers safe
// to call from anywhere in the codebase.
func LoggerFromContext(ctx context.Context) *slog.Logger {
	// TODO: if l, ok := ctx.Value(keyLogger).(*slog.Logger); ok { return l }
	// TODO: return slog.Default()
	return slog.Default()
}

// Middleware returns an http middleware that wires in request_id +
// ctx-bound logger. idGen is called to mint a new ID when the incoming
// X-Request-ID header is absent or empty; injecting it (rather than
// hard-coding uuid.New) keeps tests deterministic.
func Middleware(base *slog.Logger, idGen func() string) func(http.Handler) http.Handler {
	// TODO: return func(next http.Handler) http.Handler {
	// TODO:     return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// TODO:         id := r.Header.Get(Header)
	// TODO:         if id == "" { id = idGen() }
	// TODO:         w.Header().Set(Header, id)
	// TODO:         ctx := WithRequestID(r.Context(), id)
	// TODO:         l := base.With(slog.String("request_id", id))
	// TODO:         ctx = WithLogger(ctx, l)
	// TODO:         next.ServeHTTP(w, r.WithContext(ctx))
	// TODO:     })
	// TODO: }
	_ = base
	_ = idGen
	return func(next http.Handler) http.Handler { return next }
}
