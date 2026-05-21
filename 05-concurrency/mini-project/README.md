# Mini-project — `fanout-ping`

Concurrent URL health checker with bounded parallelism, per-request timeouts, and clean cancellation on Ctrl-C.

## Spec

```
fanout-ping --concurrency 4 --timeout 2s \
  https://example.com \
  https://example.org \
  https://example.net
```

Sample output:

```
OK    https://example.org  status=200  (52ms)
OK    https://example.net  status=200  (61ms)
BAD   https://example.com  status=503  (88ms)
```

Lines stream as each check finishes — not held until the end.

## How it's split for testability

| Function | Job |
|---|---|
| `Check(ctx, client, url)` | One HTTP GET. Transport errors → `Err` set, `Status=0`. 4xx/5xx → `Status` set, `Err=nil`. |
| `Run(ctx, client, urls, concurrency)` | Fan out `Check`. Semaphore channel bounds in-flight requests. Returns a streaming `<-chan Result`. |
| `newRootCmd()` | cobra wiring + `signal.NotifyContext` for graceful cancel. |

Splitting `Run` from cobra lets the tests drive it with `httptest.NewServer` and assert behavior without parsing flags.

## Key patterns

- **Semaphore channel**: `make(chan struct{}, N)` — acquire by sending, release by receiving. Bounded parallelism in 3 lines.
- **`signal.NotifyContext`**: Go 1.16+ shortcut for "cancel this context on SIGINT/SIGTERM." Replaces the old `signal.Notify` + goroutine + `cancel()` boilerplate.
- **Streaming results**: each completed `Check` sends on the results channel immediately. Consumer can react as soon as the slowest in-flight finishes, not after all of them.
- **Context flows through**: the same `ctx` is given to `Run`, threaded into every `Check`, and passed to `http.NewRequestWithContext`. Ctrl-C cancels everything.

## Run the tests

```
go test -tags=exercise ./05-concurrency/mini-project/...
```

All tests fail (or panic on `0 results, want N`) until you implement the TODOs in `Check` and `Run`. The concurrency-peak test is the most diagnostic: it records the max number of simultaneous in-flight handlers and asserts it equals your `--concurrency` setting.

## Stretch ideas

- Read URLs from stdin if no positional args were given.
- Sort the final summary by status code (group OK, BAD, FAIL).
- Print a JSON-lines mode behind `--format json`.
- Add retries with exponential backoff (steal from `03-http-clients/05-retry-backoff/`).
- Replace the semaphore-and-WaitGroup with `golang.org/x/sync/errgroup`'s `WithContext` + `SetLimit(N)` — same semantics, less code.
