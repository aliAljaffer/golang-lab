# Exercise 03 — Channel pipeline

Build a 3-stage pipeline of channel-connected goroutines.

## What to implement

Inside `pipeline.go`:

- `Source(ctx, nums) <-chan int` — emits each int, then closes
- `Square(ctx, in) <-chan int` — squares each input, then closes when input closes
- `Sum(ctx, in) (int, error)` — accumulates until input closes or ctx cancels

## Two rules every stage follows

1. **Return the output channel synchronously; do the work on a goroutine.** Callers don't want to block on construction.
2. **Always close the output channel exactly once.** `defer close(out)` at the top of the goroutine is the safe move. Forgetting it leaks the downstream `range`.

## The two-way `select`

`Square` is the most subtle stage: it must `select` on **both** "input has a value" and "ctx cancelled" while also being ready to bail out during a send. Look at the TODO carefully — nested selects are normal here.

```go
select {
case n, ok := <-in:
    if !ok { return }
    select {
    case out <- n * n:
    case <-ctx.Done():
        return
    }
case <-ctx.Done():
    return
}
```

## Why this matters

Pipelines compose: `Sum(ctx, Square(ctx, Source(ctx, nums)))` runs the three stages concurrently, streaming values through. This pattern is the backbone of log processors, ETL stages, image-processing chains.

## Run the tests

```
go test -tags=exercise ./05-concurrency/exercises/03-pipeline/...
```

## Stretch

- Add `FanOut(ctx, in <-chan int, n int) []<-chan int` — split a stream across n consumers.
- Add `Merge(ctx, ins ...<-chan int) <-chan int` — fan-in N streams into one.
- Make the pipeline carry an error type: `<-chan Result` instead of `<-chan int`. Each stage stops on the first Err.
- Benchmark unbuffered vs `make(chan int, 16)` on a real workload — buffering can amortize scheduler wake-ups significantly.
