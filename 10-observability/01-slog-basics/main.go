// 01-slog-basics — text vs JSON handlers, levels, attributes, AddSource.
//
// What this example proves:
//   - `slog.NewTextHandler(w, opts)` / `slog.NewJSONHandler(w, opts)` are the
//     two stdlib handlers. Text is dev-friendly; JSON is what ships to a
//     log aggregator (Loki, ELK, Datadog).
//   - Logger levels are set on the *handler*, not the logger. Same logger
//     can be cheap-cloned to point at a different-level handler.
//   - `slog.HandlerOptions{AddSource: true}` adds `source.file` and
//     `source.line` to every record — invaluable for grepping in prod, but
//     not free (it calls `runtime.Caller`).
//   - `slog.Group("user", slog.String("id", "u1"), slog.Int("age", 30))`
//     produces `user.id=u1 user.age=30` in text, nested JSON in JSON.
//
// Why slog killed third-party loggers (zap, logrus, zerolog):
//
//	stdlib in 1.21, structured-first, handler-pluggable, fast enough that
//	zap's perf edge no longer matters for ~99% of code. New Go code should
//	default to slog unless you already have a zap-shaped codebase.
//
// Run:
//
//	go run .
package main

import (
	"log/slog"
	"os"
)

func main() {
	// TODO: textLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	// TODO:     Level:     slog.LevelDebug,
	// TODO:     AddSource: true,
	// TODO: }))
	// TODO: textLogger.Info("boot", slog.String("env", "dev"), slog.Int("pid", os.Getpid()))
	// TODO: textLogger.Debug("debug visible because Level=Debug")
	// TODO: textLogger.Warn("low disk", slog.Int("free_gb", 3))

	// TODO: jsonLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	// TODO:     Level: slog.LevelInfo,
	// TODO: }))
	// TODO: jsonLogger.Info("request",
	// TODO:     slog.String("method", "GET"),
	// TODO:     slog.String("path", "/healthz"),
	// TODO:     slog.Group("user", slog.String("id", "u1"), slog.Int("age", 30)),
	// TODO: )

	// TODO: // slog.Default() is the global. SetDefault changes what `slog.Info(...)` uses.
	// TODO: slog.SetDefault(jsonLogger)
	// TODO: slog.Info("now the package-level helpers use JSON too")

	_ = slog.LevelDebug
	_ = os.Stdout
}
