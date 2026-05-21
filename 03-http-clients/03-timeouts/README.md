# 03 — Timeouts

> **Read this twice.** Missing timeouts is the single most common bug in production Go networking code.

## The bug

```go
resp, err := http.Get("https://slow.example.com/")
```

`http.Get` uses `http.DefaultClient`, and `http.DefaultClient.Timeout == 0`, which means **no timeout**. If the server hangs, your goroutine hangs forever. In a server, you accumulate hanging goroutines until you OOM.

## The fix

```go
client := &http.Client{Timeout: 10 * time.Second}
resp, err := client.Get(url)
```

The `Timeout` field is a wall-clock budget for the _entire_ request: DNS + dial + TLS handshake + sending the request + reading the response. It is the simplest correct knob.

## When you need more control

If you need to separate "time to first byte" from "time to read the whole body" (e.g., streaming), construct a custom `http.Transport` and use `context.WithTimeout` per request. See example `06-context-cancel`.

## Things to notice

- A timeout-blown request returns `err != nil` — you cannot inspect `resp` (it's nil).
- The error string contains `Client.Timeout exceeded while awaiting headers`. Most callers don't try to match on this; they just propagate the error.

## Comparison

| Concept             | Go                  | Python                          | TS                               |
| ------------------- | ------------------- | ------------------------------- | -------------------------------- |
| Per-request timeout | `client.Timeout`    | `requests.get(..., timeout=10)` | `AbortController` + `setTimeout` |
| Default             | **none** (infinity) | none (also a footgun)           | none                             |

## Run

```bash
go run .
```
