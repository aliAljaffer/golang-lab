# Exercise 01 — `url-health-check`

Given a list of URLs, return each one's HTTP status, response time, and any transport error.

## What to build

In `healthcheck.go`, implement:

```go
func CheckAll(client *http.Client, urls []string) []Result
```

Result fields:

| Field | Meaning |
|---|---|
| `URL`      | The input URL |
| `Status`   | HTTP status code; 0 if the request failed before getting a response |
| `OK`       | `Status >= 200 && Status < 300` |
| `Duration` | Wall-clock time elapsed for this request |
| `Err`      | Transport error (DNS, refused, timeout); nil on any HTTP response |

## Behaviour

- One transport error does **not** abort the run — every URL gets a `Result`.
- The result order **must match** the input order.
- Use `client.Get`; do not pass nil — the caller decides the timeout via `client.Timeout`.

## Run

```
go test -tags=exercise ./03-http-clients/exercises/01-url-health-check/...
```

## Stretch

- Run requests concurrently with a worker pool, preserving order on output.
- Add a `--concurrency N` knob in a small CLI wrapper.
- Honor a `context.Context` so the whole run can be cancelled.
