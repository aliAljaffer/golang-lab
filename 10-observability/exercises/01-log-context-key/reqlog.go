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
	// TODO: attach the id under keyRequestID.
	return ctx
}

// RequestIDFromContext returns the request ID attached to ctx, or "" if none.
func RequestIDFromContext(ctx context.Context) string {
	// TODO: pull the value back out as a string. Missing or wrong type
	//   must return ""; the ok-form of the type assertion is how you avoid
	//   panicking when the middleware didn't run.
	return ""
}

// WithLogger returns ctx with l attached.
func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	// TODO: attach the logger under keyLogger.
	return ctx
}

// LoggerFromContext returns the logger attached to ctx, or slog.Default()
// if none. Never returns nil — this is what makes ctx-bound loggers safe
// to call from anywhere in the codebase.
func LoggerFromContext(ctx context.Context) *slog.Logger {
	// TODO: pull the logger back out; the fallback to slog.Default() is the
	//   load-bearing detail — never return nil, or every caller has to
	//   nil-check before logging.
	return slog.Default()
}

// Middleware returns an http middleware that wires in request_id +
// ctx-bound logger. idGen is called to mint a new ID when the incoming
// X-Request-ID header is absent or empty; injecting it (rather than
// hard-coding uuid.New) keeps tests deterministic.
func Middleware(base *slog.Logger, idGen func() string) func(http.Handler) http.Handler {
	// TODO: extract-or-mint the ID, echo it on the response header, attach
	//   BOTH the ID and a `base.With(request_id=...)` logger to the request
	//   context, then call next with the new context. The "with"-binding is
	//   what makes every downstream log line carry request_id automatically;
	//   without it, each handler has to remember to attach it.
	_ = base
	_ = idGen
	return func(next http.Handler) http.Handler { return next }
}
