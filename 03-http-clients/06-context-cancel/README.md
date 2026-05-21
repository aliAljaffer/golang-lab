# 06 — Context cancellation

`context.Context` is Go's universal "cancel signal." Every blocking call in the standard library accepts one. For HTTP clients this is how you wire request cancellation through transports, retries, and goroutines.

## Pattern

```go
ctx, cancel := context.WithTimeout(parentCtx, 2*time.Second)
defer cancel()

req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
if err != nil { return err }
resp, err := client.Do(req)
```

- `parentCtx` is usually `r.Context()` in a handler, or `context.Background()` at program top.
- `defer cancel()` is mandatory — it releases the timer resource and avoids leaks even if the request completes early.
- When `ctx` fires, `client.Do` unblocks immediately and returns an error wrapping `context.DeadlineExceeded` (or `context.Canceled` if you called `cancel()` manually).

## When to prefer this over `client.Timeout`

- You're a library and don't own the `http.Client`.
- You want per-call timeouts that vary.
- You want cancellation to be triggered by something other than time — e.g. the user closing the connection in a server handler.

In practice you often use both: `client.Timeout` as a hard ceiling, plus per-call `context.WithTimeout` for the operation's natural budget.

## Things to notice

- `errors.Is(err, context.DeadlineExceeded)` — this is how you detect "timed out" robustly. Don't string-match.
- `cancel()` is cheap to call multiple times. Always `defer cancel()`.

## Comparison

| Concept          | Go                    | Python                         | TS                                |
| ---------------- | --------------------- | ------------------------------ | --------------------------------- |
| Cancel signal    | `context.Context`     | n/a (lib-specific)             | `AbortController` / `AbortSignal` |
| Timeout per call | `context.WithTimeout` | `requests.get(..., timeout=2)` | `AbortSignal.timeout(2000)`       |

## Run

```bash
go run .
```
