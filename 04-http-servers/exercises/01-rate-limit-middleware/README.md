# Exercise 01 — Rate-limit middleware

Build an IP-based rate limiter as net/http middleware.

## What to implement

Inside `ratelimit.go`:

- `(*Limiter).Allow(key string) bool` — fixed-window counter, concurrency-safe.
- `Middleware(l *Limiter) func(http.Handler) http.Handler` — call `l.Allow(r.RemoteAddr)`; if false, return 429 and skip the inner handler.

## Why split it this way

Splitting `Limiter` (pure logic, clock-injectable) from `Middleware` (HTTP wrapper) means you can test the window-reset behavior **deterministically** by overriding `l.Now` — no `time.Sleep` in tests.

Look at `TestAllow_ResetsAfterWindow` — the test moves a fake clock forward by 1.5s instantly. You'd hate writing this with real sleeps once you have a dozen of them.

## Run the tests

```bash
go test -tags=exercise ./04-http-servers/exercises/01-rate-limit-middleware/...
```

## Stretch

- Switch to a **token bucket**: tokens refill at a rate, requests consume one each. Smoother than fixed-window.
- Strip the port from `RemoteAddr` so the same client across requests keys to the same bucket.
- Honor `X-Forwarded-For` when the server is behind a proxy.
- Add a janitor goroutine that evicts buckets unused for > 10 windows so the map doesn't grow unbounded.
