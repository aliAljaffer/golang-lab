# 07 — Worker pool

Bounded parallelism over a stream of jobs. One of the highest-mileage patterns in Go.

## The shape

```bash
producer ──> jobs ────┐
                      ├──> [worker 1] ┐
                      ├──> [worker 2] ├──> results ──> consumer
                      └──> [worker 3] ┘
```

- A `jobs` channel feeds N worker goroutines.
- Each worker pulls from `jobs` (`for j := range jobs`) until the channel is closed.
- Workers push onto a `results` channel.
- A separate goroutine `wg.Wait()`s then `close(results)` — never close from a worker, and never close before workers finish.

## Why this pattern wins

- **Backpressure**: an unbuffered `jobs` channel means the producer blocks when all workers are busy. No need to track in-flight work or write a semaphore.
- **Bounded concurrency**: exactly N workers means at most N concurrent operations. Trivial to reason about, trivial to tune.
- **Composes with `context.Context`**: pass `ctx` into `worker` and select on `ctx.Done()` for cancellation.

## The "who closes what" rule

| Channel   | Closed by                                        |
| --------- | ------------------------------------------------ |
| `jobs`    | producer, after the last send                    |
| `results` | a single goroutine that `wg.Wait()`s the workers |

If you remember nothing else: **the sender closes**. The receiver never closes.

## Semaphore variant

For ad-hoc "do these in parallel, up to N at a time," a semaphore channel is simpler than a full pool:

```go
sem := make(chan struct{}, N)
for _, item := range items {
    sem <- struct{}{}             // acquire (blocks if N in flight)
    go func(it Item) {
        defer func() { <-sem }()  // release
        work(it)
    }(item)
}
```

The mini-project (`fanout-ping`) uses this variant.

## Run

```bash
go run .
```
