// 02-slog-context — logger-in-context pattern.
//
// The problem: an HTTP request enters a handler, a logger is built with
// per-request attributes (request_id, user_id), then four function calls
// deeper a helper wants to log too — without taking a `*slog.Logger`
// parameter everywhere.
//
// The pattern (canon in every slog-shaped codebase):
//
//	type ctxKey struct{}
//
//	func WithLogger(ctx, l) context.Context  // stash
//	func FromContext(ctx) *slog.Logger        // retrieve, fallback to default
//
// Then middleware sets it once at the boundary; everything reads it from ctx.
//
// What this example proves:
//   - `logger.With(slog.String("k", "v"))` returns a NEW logger with that
//     attribute pre-bound. Doesn't mutate the original.
//   - The "logger in context" pattern is just a typed key + helper pair —
//     no library needed.
//   - `slog.Default()` is the safe fallback so callers don't have to handle
//     a nil-logger case.
//
// Run:
//
//	go run .
package main

import (
	"context"
	"log/slog"
	"os"
)

// ctxKey is unexported so callers must use WithLogger / FromContext.
// Using a typed struct{} (not a string) prevents collisions with other
// packages using the same key name.
type ctxKey struct{}

// WithLogger returns ctx with l attached. Caller-side helper.
func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

// FromContext retrieves the logger attached to ctx, or slog.Default() if
// none is set. The fallback is what makes this pattern safe to call from
// anywhere — your library code doesn't have to know whether a logger was
// stashed.
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

// doWork is the "four calls deep" helper that needs to log without being
// passed the logger explicitly.
func doWork(ctx context.Context) {
	// TODO: FromContext(ctx).Info("doing work", slog.String("step", "validate"))
	_ = ctx
}

func main() {
	base := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Simulate request boundary: build a per-request logger with request_id
	// pre-bound, stash it on ctx.
	// TODO: reqLogger := base.With(slog.String("request_id", "req-abc123"))
	// TODO: ctx := WithLogger(context.Background(), reqLogger)

	// Now any code anywhere down the stack can pull a logger that
	// already carries request_id.
	// TODO: doWork(ctx)
	// TODO: doWork(ctx)

	// Compare: without ctx, you get the default logger (no request_id).
	// TODO: doWork(context.Background())

	_ = base
	_ = WithLogger
	_ = FromContext
	_ = doWork
	_ = context.Background
}
