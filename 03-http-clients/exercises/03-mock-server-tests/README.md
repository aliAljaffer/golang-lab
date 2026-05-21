# Exercise 03 — `mock-server-tests`

Practice testing real HTTP behaviour using `httptest.NewServer` instead of mocking the `http.Client`.

## Why httptest.NewServer over a mock client

A common mistake is to mock `http.Client`'s `Do` method via an interface seam. That tests *your code's call into the abstraction*, not the actual networking. `httptest.NewServer` spins up a real localhost server in microseconds — you get to test:

- That you set the right headers.
- That `defer resp.Body.Close()` is in the right place (you'll find leaks with `-race` and goroutine counts).
- That the `http.Transport` connection pool actually reuses connections.
- That redirect / cookie / TLS handling is correct.

All in a hermetic, network-free test.

## What to build

In `retry.go`, implement:

```go
func DoWithRetry(client *http.Client, req *http.Request, maxAttempts int) (*http.Response, error)
```

Rules:

| Condition | Action |
|---|---|
| Transport error | retry |
| Status 429 | retry |
| Status 5xx | retry |
| Status 2xx or 4xx | return immediately |
| Out of attempts | return the last response/error |

**Don't leak bodies.** On every retried attempt, close the response Body before sleeping. The tests don't assert this directly, but `go test -race` and a follow-up `httptest.Server.CloseClientConnections` test will catch it.

## Run

```
go test -tags=exercise ./03-http-clients/exercises/03-mock-server-tests/...
```

## Stretch

- Honor `Retry-After` on 429/503 (parse seconds and HTTP-date).
- Add a context-aware variant: `DoWithRetry(ctx, ...)` that returns early on `ctx.Done()`.
- Cap total elapsed time instead of total attempts (`maxElapsed time.Duration`).
- Write a property-based test using `testing/quick` that asserts: "for any sequence of server responses, the number of calls equals min(maxAttempts, first-non-retryable-index+1)".
