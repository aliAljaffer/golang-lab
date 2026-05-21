# Exercise 02 — Broadcast

A pub/sub broker that fans out each `Publish` to every current subscriber.

## What to implement

Inside `broker.go`:

- `New() *Broker`
- `(*Broker).Subscribe() <-chan string`
- `(*Broker).Unsubscribe(ch <-chan string)` — also closes the channel
- `(*Broker).Publish(msg string)` — non-blocking per-subscriber sends
- `(*Broker).Close()` — closes every subscriber channel; further publishes are no-ops

## Design

- **Per-subscriber buffered channel** (size 8). Each Publish iterates subscribers and sends. Buffering decouples a slow consumer from the publisher.
- **`sync.RWMutex`** protecting the subscribers map. Publish takes `RLock`; mutating ops (Subscribe / Unsubscribe / Close) take `Lock`.
- **Non-blocking sends** in Publish: `select { case ch <- msg: default: }`. A slow subscriber drops messages rather than stalling the publisher. Tests stay within buffer so they don't observe drops.
- **Idempotent Close**. Double-Close (or Unsubscribe-then-Close) must not double-close a channel, which panics.

## Why two channel-direction types matter here

The map stores `chan string` (bidirectional — broker needs to send and close). Subscribers receive `<-chan string` (receive-only — they shouldn't be able to send into their own subscription or close it). The `Unsubscribe(<-chan string)` argument matches the type the caller has; internally you compare by underlying channel identity (`(<-chan string)(k) == ch`).

## Run the tests

```bash
go test -tags=exercise ./05-concurrency/exercises/02-broadcast/...
```

## Stretch

- Add `SubscribeTopic(topic string)` so only matching publishes deliver.
- Replace drop-on-full with `lossless: true` per subscriber: in that mode, Publish blocks (or spawns a forwarder goroutine).
- Add metrics: how many messages were dropped per subscriber.
- Make `Publish` cancellable: `Publish(ctx, msg) error`.
