# 05 — Signals & graceful shutdown

The DevOps essential. A long-running process needs to clean up when the
orchestrator sends `SIGTERM` (Kubernetes does this 30 seconds before a
`SIGKILL` during pod termination).

## The pattern

```go
ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
defer stop()

// do work, periodically checking ctx.Done()
select {
case <-ctx.Done():
    // graceful shutdown: flush buffers, close connections, etc.
}
```

`signal.NotifyContext` (Go 1.16+) is the modern API — it returns a context that
cancels when any of the listed signals arrives. Before that, the idiom was a
channel:

```go
sig := make(chan os.Signal, 1)
signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
<-sig
```

Either works; `NotifyContext` composes better with `http.Server.Shutdown(ctx)`,
`exec.CommandContext`, etc.

## Things to learn

- The channel **must** be buffered (capacity 1). Signal delivery is non-blocking; a full unbuffered channel drops the signal.
- `SIGKILL` (signal 9) and `SIGSTOP` cannot be caught. By definition.
- On Windows, only `os.Interrupt` (Ctrl-C) is reliable. `SIGTERM` doesn't exist there.

## Comparison

| Language | Idiom |
|---|---|
| Go | `signal.NotifyContext(ctx, SIGINT, SIGTERM)` |
| Python | `signal.signal(signal.SIGTERM, handler)` |
| Bash | `trap 'cleanup' INT TERM` |
