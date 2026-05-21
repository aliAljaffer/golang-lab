# 06 — Context cancellation

`context.Context` is Go's standard way to say "stop what you're doing" across goroutine and API boundaries.

## Things to notice

- A `Context` carries a deadline + cancellation signal + request-scoped values down a call tree. Pass it as the **first argument** to any function that does IO or runs work.
- `ctx.Done()` returns a `<-chan struct{}` that's closed when the context is cancelled (manually, timeout, or deadline). Select on it.
- `ctx.Err()` tells you *why*: `context.Canceled`, `context.DeadlineExceeded`, or nil if still live.
- `WithCancel`, `WithTimeout`, `WithDeadline` all return `(ctx, cancel)`. **Always `defer cancel()`** — even if the timeout already fired. Forgetting to cancel leaks the timer and the parent's child list.
- Cancellation is cooperative. The runtime does not interrupt your goroutine. *You* must select on `ctx.Done()` and return.

## The canonical worker loop

```go
for {
    select {
    case <-ctx.Done():
        return
    case job := <-jobs:
        // process job, ideally also passing ctx into anything blocking
    }
}
```

Everything in this section, the next, and most production Go is some variation on this.

## Values

`context.WithValue(ctx, key, val)` carries request-scoped data (trace IDs, auth subjects) through the call tree. Use sparingly — context is not a general-purpose map. Don't put config in there.

## Comparison

| Concept | Go | Python | TS / Node |
|---|---|---|---|
| Cancellation signal | `context.Context` | `asyncio.CancelledError` raised in awaited coro | `AbortSignal` |
| Cancel trigger | `cancel()` func | `task.cancel()` | `controller.abort()` |
| Deadline | `context.WithTimeout` | `asyncio.wait_for` | `AbortSignal.timeout()` |

## Run

```
go run .
```
