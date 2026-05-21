# Exercise 01 — Token-bucket rate limiter

Implement a token-bucket rate limiter using a single buffered channel and a refill goroutine.

## What to implement

Inside `bucket.go`:

- `New(capacity, refillEvery)` — pre-fills the bucket to `capacity` and starts a refiller goroutine
- `(*Bucket).Allow() bool` — non-blocking try-take
- `(*Bucket).Wait(ctx) error` — blocking take; respects ctx cancel and `Stop`
- `(*Bucket).Stop()` — halts the refiller

## Why a channel?

The buffered-channel trick is the most idiomatic Go implementation:

- **Bucket size = channel capacity.** No counter, no mutex.
- **Add a token = send.** Drop if full via `select { case ch <- t: default: }`.
- **Consume a token = receive.**
- **Non-blocking try = `select` with `default`.**
- **Blocking wait = `select` with `<-ctx.Done()` and `<-stop` arms.**

Mutex-based implementations work too but they're 2× the code and easier to deadlock.

## Run the tests

```
go test -tags=exercise ./05-concurrency/exercises/01-rate-limiter/...
```

## Stretch

- Add a `WaitN(ctx, n)` that consumes N tokens atomically (a single Wait should not return holding only 3 out of 5).
- Allow a burst arg distinct from capacity (`capacity` = max stored, `burst` = max consumed in one Wait).
- Replace the ticker with a calculated next-token timestamp so the bucket is correct even if the goroutine is paused (no token "loss" on a sluggish scheduler).
- Try a mutex+counter version and benchmark both. Channels are idiomatic, but a contended channel can be slower than a tight mutex.
