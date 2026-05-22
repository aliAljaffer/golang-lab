# 02 — slog in context

How do you get a per-request logger to code that's four calls deep without
threading `*slog.Logger` through every signature? You stash it on `context.Context`.

## The pattern

```go
type ctxKey struct{}

func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
    return context.WithValue(ctx, ctxKey{}, l)
}

func FromContext(ctx context.Context) *slog.Logger {
    if l, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok {
        return l
    }
    return slog.Default()
}
```

Three rules:

1. **Key is an unexported typed struct** (`type ctxKey struct{}`), not a
   string. Strings collide; typed keys can't.
2. **`FromContext` falls back to `slog.Default()`.** Never returns nil; never
   forces the caller to nil-check. This is what makes the pattern safe to
   call from anywhere in the codebase.
3. **`logger.With(...)` is a builder** — it returns a new logger with the
   given attributes pre-bound. Doesn't mutate.

## The middleware shape

In an HTTP server, the boundary handler builds the per-request logger and
stashes it once:

```go
func RequestLogger(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        l := slog.Default().With(
            slog.String("request_id", uuid.New().String()),
            slog.String("method", r.Method),
            slog.String("path", r.URL.Path),
        )
        ctx := WithLogger(r.Context(), l)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

Every handler downstream just calls `FromContext(r.Context())` and gets a
logger that already knows the request ID. No threading.

## When NOT to use ctx-bound loggers

- **Initialization code** that runs once at startup. Use a package-level
  logger or pass it explicitly.
- **Library packages** that don't otherwise need a `context.Context`. Don't
  introduce ctx just for logging — pass `*slog.Logger` as a parameter.

The pattern is for *request-scoped* attributes (request_id, user_id, trace_id).
Service-scoped attributes (service name, version) belong on the base logger.

## Compare to other ecosystems

|                 | Go                       | Python                          | TS / Node                |
|-----------------|--------------------------|---------------------------------|--------------------------|
| Per-request    | stash on ctx              | `structlog.contextvars.bind_contextvars` | `pino` child loggers / AsyncLocalStorage |
| Retrieve in helper | `FromContext(ctx)`     | implicit (contextvars)          | implicit (AsyncLocalStorage) |

Go is more explicit because it doesn't have implicit thread-local storage —
which is annoying boilerplate at first but means you can grep for *every*
logger plumbing point.

## TODO

1. Uncomment the TODO block. Run `go run .`.
2. Confirm the first two log lines have `request_id=req-abc123`, the third
   does not.
3. Try passing a different logger via `WithLogger` and observe `doWork`
   pick it up automatically.
4. Add a second value (e.g. `user_id`) by calling `.With(...)` again on the
   ctx-bound logger before stashing.
