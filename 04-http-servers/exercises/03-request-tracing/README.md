# Exercise 03 — Request tracing middleware

A request-ID middleware with context propagation and lifecycle logging.

## What to implement

In `tracing.go`:

- `WithRequestID(idGen, logger)` — middleware that:
  - Reuses incoming `X-Request-ID` if set, else mints one via `idGen()`.
  - Stashes the ID on `r.Context()` so inner handlers can read it.
  - Echoes it on the response header.
  - Logs a start line and an end line (with status + duration) via `logger`.
- `RequestIDFromContext(ctx)` — accessor for handlers.

## Why this pattern

The request ID is the **trace ID**. Every log line your handler emits should include it. When something explodes, you grep one ID across N services and reconstruct the full request path. Without it, you have N piles of logs you cannot correlate.

The `idGen` parameter is **injection for testability**: in tests you return a fixed string; in main you'd use crypto-random hex bytes. The middleware doesn't care.

## On context keys

```go
type ctxKey int
const requestIDKey ctxKey = 0
```

Using an unexported `ctxKey` _type_ (not a `string`) prevents accidental collisions with other packages stuffing things in the same context — the type is package-private so no one outside can construct a matching key. Stdlib uses the same pattern.

## Run the tests

```bash
go test -tags=exercise ./04-http-servers/exercises/03-request-tracing/...
```

## Stretch

- Add a `X-Trace-ID` second header for compatibility with OpenTelemetry-style propagation.
- Inject `slog.Logger` instead of `log.Logger` and attach the request ID as a `slog.Attr` so every nested log line in the handler carries it automatically.
- Wire the middleware into example 03's chain and confirm the request ID shows up in panic recovery logs.
