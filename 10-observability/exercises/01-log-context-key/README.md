# Exercise 01 — request_id middleware

Implement an HTTP middleware that gives every request a `request_id` —
honoured from the incoming `X-Request-ID` header or generated fresh — and
makes that ID available to every log line within the request via a
ctx-bound `*slog.Logger`.

## What you implement

```go
const Header = "X-Request-ID"

func WithRequestID(ctx, id) context.Context
func RequestIDFromContext(ctx) string                  // "" if missing

func WithLogger(ctx, l) context.Context
func LoggerFromContext(ctx) *slog.Logger              // slog.Default() if missing

func Middleware(base *slog.Logger, idGen func() string) func(http.Handler) http.Handler
```

## Behavior pinned by tests

- `WithRequestID` round-trips.
- `RequestIDFromContext` returns `""` (not nil, not error) when absent.
- Middleware **honours incoming header** if set — never regenerates over it.
- Middleware **generates via `idGen`** when header absent or empty.
- Middleware **echoes the ID back** as `X-Request-ID` on the response.
- The logger on ctx has `request_id="..."` pre-bound — downstream `Info()`
  calls don't need to add it manually.

## Key idea — `idGen` is injected

```go
Middleware(logger, uuid.New().String)  // production
Middleware(logger, func() string { return "GENERATED" })  // tests
```

This is the same "inject the clock" pattern from `06-testing/02-fake-clock`.
The exercise tests rely on a deterministic ID generator — production should
use `uuid.New().String` or `ulid.Make().String()`.

## Run the failing suite

```bash
go test -tags=exercise ./10-observability/exercises/01-log-context-key/
```
