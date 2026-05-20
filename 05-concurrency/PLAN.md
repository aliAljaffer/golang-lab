# Plan: 05-concurrency

## Concepts to cover

- [ ] Goroutines: `go f()`, the runtime scheduler, stack growth
- [ ] Channels: unbuffered (synchronous handoff) vs buffered (queue)
- [ ] Channel direction (`chan<-`, `<-chan`) in function signatures
- [ ] `close(ch)` and the comma-ok receive
- [ ] `select` statement: multiplexing, default case, timeout pattern
- [ ] `sync.WaitGroup` — wait for N goroutines
- [ ] `sync.Mutex` and `sync.RWMutex` — when channels aren't enough
- [ ] `sync.Once` — initialize-exactly-once
- [ ] `context.Context` — propagating cancellation across goroutines
- [ ] Worker pool pattern: jobs channel + N workers + results channel
- [ ] The race detector: `go test -race`, `go run -race`
- [ ] Common bugs: goroutine leaks, send-on-closed-channel panic, deadlocks

## Examples to build

| Folder | Demonstrates |
|---|---|
| `01-goroutine-basic/` | `go f()` and why a missing WaitGroup loses output |
| `02-channels/` | Unbuffered vs buffered, send/recv pairs |
| `03-select/` | Multiplexing two channels + timeout |
| `04-waitgroup/` | Wait for N goroutines to finish |
| `05-mutex/` | Protecting a shared map (and showing the race without it) |
| `06-context-cancel/` | Goroutine that stops when context is canceled |
| `07-worker-pool/` | Bounded worker pool pattern |
| `08-race-detector/` | A program that races; show `-race` flag catches it |

## Mini-project

**`fanout-ping`** — concurrently checks the HTTP health of N URLs with a configurable parallelism limit and per-request timeout. Outputs results as they arrive (streaming). Cancels remaining work on Ctrl-C.

Tests verify:
- Respects `--concurrency` limit (uses a semaphore channel)
- Per-request timeout works
- Context cancellation propagates (verify via test that interrupts mid-run)

## Exercises

1. **`01-rate-limiter`** — implement a token-bucket rate limiter using a channel
2. **`02-broadcast`** — fan-out a single value to N subscribers (channels)
3. **`03-pipeline`** — build a 3-stage pipeline (source → transform → sink) using channels

## Status

- [ ] Concepts in README walkthrough
- [ ] Examples 01-08 built
- [ ] Mini-project `fanout-ping` built + tested
- [ ] Exercises scaffolded
