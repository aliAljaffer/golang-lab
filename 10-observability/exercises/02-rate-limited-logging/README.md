# Exercise 02 — rate-limited logging

A worker hits "connection refused" 10,000 times a second when the downstream
is down. You want to know about it — but one log line per second per error
key is enough. Implement a per-key rate limiter for log calls.

## What you implement

```go
type Limiter struct {
    Window time.Duration
    Now    func() time.Time   // injectable clock
    // (unexported state)
}
func New(window time.Duration) *Limiter
func (l *Limiter) Allow(key string) bool
```

## Behavior pinned by tests

- First call for a key returns true.
- Subsequent call within `Window` returns false.
- After `Window` expires, returns true again (and stamps the new time).
- Keys are independent — `Allow("a")` doesn't affect `Allow("b")`.
- Concurrency safe (the tests don't pound on it, but the mutex is part of
  the design).

## Idiomatic usage

```go
limiter := ratelog.New(1 * time.Second)
for err := range errors {
    if limiter.Allow("conn-refused") {
        logger.Error("downstream connection refused", slog.Any("err", err))
    }
    // ... handle err either way ...
}
```

## Wrap as a slog handler (optional follow-up)

The real win is wrapping this as a `slog.Handler` so callers don't have to
do the `if Allow { log }` dance:

```go
type Handler struct {
    base    slog.Handler
    limiter *Limiter
    keyFunc func(slog.Record) string   // e.g., extract a "key" attr
}

func (h Handler) Handle(ctx context.Context, r slog.Record) error {
    if !h.limiter.Allow(h.keyFunc(r)) { return nil }
    return h.base.Handle(ctx, r)
}
// + Enabled / WithAttrs / WithGroup forwarders
```

This isn't in the tests — but it's the natural next step. The
`slog.Handler` interface is exactly four methods; wrapping is mechanical.

## The injected clock

The Now field defaults to `time.Now` in `New`. Tests overwrite it:

```go
now := time.Unix(0, 0)
l.Now = func() time.Time { return now }
// advance time without sleeping
now = now.Add(2 * time.Second)
```

Same pattern as `06-testing/02-fake-clock`. Real-clock tests are flaky;
injected clocks are deterministic.

## Run the failing suite

```bash
go test -tags=exercise ./10-observability/exercises/02-rate-limited-logging/
```
