# 05 — Retries with backoff + jitter

Networks are flaky and servers occasionally return 503/504. A retry loop with backoff is table-stakes for any client that talks to a real API.

## The rules

1. **Retry on:** transport errors (DNS/connection reset), 5xx responses, 429 (rate-limited — though that one should honor `Retry-After`).
2. **Don't retry on:** 2xx (success), 3xx (the client follows redirects automatically), 4xx (won't change).
3. **Exponential backoff:** sleep `base * 2^attempt`, so 200ms → 400ms → 800ms → 1.6s …
4. **Jitter:** add a random fraction so a herd of clients doesn't all retry on the same tick. Without jitter, you can DDoS a server during its own recovery.
5. **Always close the response Body on the doomed attempt** — otherwise you leak file descriptors.

## What you don't have to write yourself

For real production code, `github.com/cenkalti/backoff/v4` is the de-facto choice. This exercise builds the loop from scratch so you understand what it does.

## Things to notice

- The retry loop **must close the previous response Body** before the next attempt, even on retryable failures. Missing this is a slow-burn fd leak that only shows up in a long-running service.
- Use `math/rand` (not `crypto/rand`) for jitter — it's cheaper and the randomness quality doesn't matter here.
- Many APIs return a `Retry-After` header on 429/503. A nicer retry loop honors it.

## Comparison

| Concept | Go | Python | TS |
|---|---|---|---|
| Retry library | `cenkalti/backoff` | `tenacity` | `axios-retry` |
| Jitter | manual | `tenacity.wait_random_exponential` | manual |

## Run

```
go run .
```
