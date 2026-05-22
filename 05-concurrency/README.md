# 05 — Concurrency

> Status: ☑ scaffolded — examples + mini-project + exercises ready to implement. See [`PLAN.md`](./PLAN.md).

The part of Go that has no direct analog in your other languages. Goroutines are *cheap* — kilobytes of stack, not megabytes — so spawning ten thousand of them is normal. Channels are first-class: typed, blocking, with `select` for multiplexing. And the language-level race detector catches the most common concurrency bugs before they hit production.

The cultural slogan is **"don't communicate by sharing memory; share memory by communicating."** Reach for channels first, mutexes only when channels would be awkward (caches, large shared maps, sync.Once-style init).

---

## What you'll learn

- Goroutines: `go f()`, the runtime scheduler, why a missed `Wait` swallows output
- Channels: unbuffered (synchronous handoff) vs buffered (queue), channel direction (`chan<-`, `<-chan`) in signatures
- `close(ch)` and the comma-ok receive (`v, ok := <-ch`)
- `select` statement: multiplexing, the `default` case for non-blocking sends/receives, the timeout pattern via `time.After`
- `sync.WaitGroup` — wait for N goroutines to finish
- `sync.Mutex` / `sync.RWMutex` — when channels aren't the right tool
- `sync.Once` — initialize-exactly-once
- `context.Context` — propagating cancellation across goroutine trees
- The worker pool pattern: jobs channel + N workers + results channel
- The race detector: `go test -race`, `go run -race`
- The common bugs: goroutine leaks, send-on-closed-channel panics, deadlocks

---

## Mental model from other languages

| Concept                 | Go                          | Python                                            | TS / Node                             |
| ----------------------- | --------------------------- | ------------------------------------------------- | ------------------------------------- |
| Lightweight concurrency | goroutine (`go f()`)        | `asyncio` task / thread                           | `async function` / Promise            |
| Channels                | `chan T`                    | `asyncio.Queue` / `queue.Queue`                   | (no native; use libraries)            |
| Wait for N tasks        | `sync.WaitGroup`            | `asyncio.gather`                                  | `Promise.all`                         |
| Cancellation            | `context.Context`           | cancellation tokens (manual) / `asyncio.CancelledError` | `AbortController`              |
| Locks                   | `sync.Mutex`                | `threading.Lock` / `asyncio.Lock`                 | (rare — single-threaded JS)           |
| Race detection          | `go test -race` (built-in)  | `pytest-xdist` (no equiv) / TSan via C extensions | none                                  |
| One-time init           | `sync.Once`                 | module-level lazy init                            | module-level lazy init                |

**Go's twist:** goroutines are scheduled by the Go runtime (M:N onto OS threads), not by you. You don't `await` them — they just run; you communicate via channels or wait on them via `WaitGroup`. And **the race detector is in the toolchain** — every `go test` you write should at some point pass under `-race`.

---

## The DevOps angle

Parallelizing infra work is where this section earns its keep: ping 100 servers in 10 seconds instead of 100 seconds, watch 10 Kubernetes informers concurrently from one binary, fan-out a deploy command to 50 hosts with a concurrency cap. The mini-project (`fanout-ping`) is exactly this pattern.

The non-obvious production details:

- **Always wire `context.Context` down through every blocking call.** Without it, `Ctrl-C` on a hung tool only aborts the goroutine reading stdin — the 50 worker goroutines keep running until the process exits, sometimes long after the user has walked away.
- **Use a bounded worker pool, not "one goroutine per task".** Unbounded goroutines hammering an external API will trip its rate limiter (and yours).
- **Run `-race` in CI.** Real concurrency bugs are non-deterministic; the race detector flips them deterministic. The cost is ~2× slower test runs, which is a bargain.

---

## Walkthrough

Read these in order. Each is a runnable example demonstrating one specific concept.

1. [`01-goroutine-basic/`](./01-goroutine-basic/) — `go f()` and why a missing `WaitGroup` loses the output. The smallest demonstration of "goroutines don't block `main`."
2. [`02-channels/`](./02-channels/) — unbuffered vs buffered, send/receive pairing. The unbuffered channel is a *synchronization point*, not just a queue.
3. [`03-select/`](./03-select/) — multiplexing two channels + a timeout via `<-time.After(d)`. The `default` clause for non-blocking sends/receives.
4. [`04-waitgroup/`](./04-waitgroup/) — wait for N goroutines to finish. The `defer wg.Done()` discipline and the `wg.Add(1)` before `go f()` rule.
5. [`05-mutex/`](./05-mutex/) — protect a shared map; the same code without the mutex visibly races (run it with `-race`).
6. [`06-context-cancel/`](./06-context-cancel/) — goroutine that stops when its context is cancelled. The `select { case <-ctx.Done(): return }` idiom is everywhere in real Go code.
7. [`07-worker-pool/`](./07-worker-pool/) — bounded worker pool: a jobs channel, N worker goroutines, a results channel, and a `WaitGroup` that closes the results channel when all workers finish. Memorize this shape.
8. [`08-race-detector/`](./08-race-detector/) — a program that races; `go run -race ./...` catches it and prints the offending stack traces. The most underused tool in the Go toolchain.

---

## Mini-project: [`fanout-ping`](./mini-project/)

Concurrently check the HTTP health of N URLs with a configurable parallelism limit and per-request timeout. Output results as they arrive (streaming). Cancel remaining work on `Ctrl-C`.

The point: real fan-out tools combine the worker-pool pattern, per-request timeouts, context propagation for cancellation, and streamed output (you don't wait for the slowest URL before printing the first result). Tests verify the `--concurrency` cap is respected (semaphore channel pattern), per-request timeouts fire, and context cancellation propagates mid-run.

Spec and starter in [`mini-project/`](./mini-project/).

---

## Exercises

See [`exercises/`](./exercises/):

1. **[`01-rate-limiter`](./exercises/01-rate-limiter/)** — implement a token-bucket rate limiter using a channel + a refill goroutine. The pattern that every API client should bring to a rate-limited downstream.
2. **[`02-broadcast`](./exercises/02-broadcast/)** — fan-out a single value to N subscribers (channels). The "pub-sub" primitive at its smallest.
3. **[`03-pipeline`](./exercises/03-pipeline/)** — build a 3-stage pipeline (source → transform → sink) using channels. Each stage is a goroutine; channels close down the chain on completion.

---

## Further reading

- [Go Memory Model](https://go.dev/ref/mem) — dense but authoritative on what `-race` is actually checking
- [Rob Pike: "Go Concurrency Patterns"](https://www.youtube.com/watch?v=f6kdp27TYZs) — the canonical talk, holds up
- [Sameer Ajmani: "Advanced Go Concurrency Patterns"](https://www.youtube.com/watch?v=QDDwwePbDtw) — the worker pool + cancellation patterns from the source
- [`context` docs](https://pkg.go.dev/context) — the cancellation API; example 06 is the smallest possible introduction
- [`sync` docs](https://pkg.go.dev/sync) — `Mutex`, `RWMutex`, `WaitGroup`, `Once`, `Map`, `Pool`
