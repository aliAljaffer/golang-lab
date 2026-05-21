# 05 — Concurrency

> Status: ☑ scaffolded — examples + mini-project + exercises ready to implement. See [`PLAN.md`](./PLAN.md).

## What you'll learn

- Goroutines: cheap, preemptive, thousands-at-a-time threading
- Channels: typed pipes for communication (the Go idiom)
- `select` for multiplexing
- `sync.WaitGroup`, `sync.Mutex`, `sync.Once`
- `context.Context` for cancellation and deadlines
- Worker pool patterns
- The race detector (`go test -race`)

## Mental model from other languages

| Concept | Go | Python | TS / Node |
|---|---|---|---|
| Lightweight concurrency | goroutine (`go f()`) | `asyncio` task / thread | `async function` / Promise |
| Channels | `chan T` | `asyncio.Queue` / `queue.Queue` | (no native; use libraries) |
| Wait for N tasks | `sync.WaitGroup` | `asyncio.gather` | `Promise.all` |
| Cancellation | `context.Context` | cancellation tokens (manual) / `asyncio.CancelledError` | `AbortController` |
| Locks | `sync.Mutex` | `threading.Lock` / `asyncio.Lock` | (rare — single-threaded JS) |

**Go's twist:** goroutines are *cheap* (KB stack, not MB) — you can spawn thousands. And "don't communicate by sharing memory; share memory by communicating" — channels first, locks only when necessary. This is the part of Go that has no direct analog in your other languages.

## The DevOps angle

Parallelizing infra tasks: ping 100 servers, scan a CIDR range, kick off 50 deploys, watch 10 k8s informers concurrently. Goroutines + channels make this trivial — and the race detector keeps you honest.

See [`PLAN.md`](./PLAN.md).
