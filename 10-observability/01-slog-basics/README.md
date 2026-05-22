# 01 — slog basics

`log/slog` is the structured logger that landed in Go 1.21. It's stdlib, so
no module to add. New Go code should default to it.

## Three things to internalize

1. **Handler vs logger.** The logger is a thin wrapper; the handler decides
   format + destination + level filtering. To change format, swap handlers.
2. **Levels live on the handler.** `slog.HandlerOptions{Level: slog.LevelDebug}`.
   Cheap to clone a logger with a different level — `logger.WithGroup`,
   `logger.With(...)`.
3. **`slog.String`, `slog.Int`, `slog.Group`** are the attribute constructors.
   Always prefer them over `slog.Any` for known types — they're zero-alloc
   in the fast path.

## Two handlers, one API

| Handler | Output | When |
|---|---|---|
| `slog.NewTextHandler` | `time=... level=INFO msg=boot env=dev pid=42` | dev terminal, local tail |
| `slog.NewJSONHandler` | `{"time":"...","level":"INFO","msg":"boot","env":"dev","pid":42}` | prod, shipping to aggregators |

Both honour the same `*HandlerOptions`. Both write to any `io.Writer`.

## `AddSource: true`

Adds `source={file:..., line:...}` (text) or `"source":{"function":...,"file":...,"line":...}` (JSON). Costs one `runtime.Caller`
per record. Worth it in production; turn off in tight benchmark loops.

## Groups

```go
logger.Info("request",
    slog.String("method", "GET"),
    slog.Group("user",
        slog.String("id", "u1"),
        slog.Int("age", 30),
    ),
)
```

Text: `method=GET user.id=u1 user.age=30`
JSON: `{"method":"GET","user":{"id":"u1","age":30}}`

Groups are the right tool for "namespace this set of fields" — a request's
HTTP info, a DB call's stats, etc.

## Compare to other ecosystems

|                 | Go (`log/slog`)              | Python (`structlog`)         | TS / Node (`pino`)        |
|-----------------|------------------------------|------------------------------|---------------------------|
| Library         | stdlib                       | `structlog` (third-party)    | `pino` (third-party)      |
| JSON output     | `slog.NewJSONHandler`        | `JSONRenderer`               | default                   |
| Level filter    | `slog.HandlerOptions.Level`  | `wrap_logger(level=...)`     | `pino({level: 'info'})`   |
| Field add       | `logger.With(slog.String...)`| `logger.bind(...)`           | `logger.child({...})`     |

## TODO

1. Uncomment the TODO block, `go run .`.
2. Observe the difference between the text and JSON outputs.
3. Drop the `AddSource: true` and re-run — `source=` disappears.
4. Add a `slog.LevelError` line that you don't want to leak to text-handler
   output, then change the text handler's `Level` to `slog.LevelError` and
   confirm only that one survives.
