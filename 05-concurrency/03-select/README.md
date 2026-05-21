# 03 — Select

`select` is `switch` for channel operations. It is the second piece of the goroutine puzzle (channels being the first).

## Things to notice

- Each `case` is a channel send or receive. The first case whose op can proceed runs.
- If **multiple** cases are ready, Go picks one at **random**. No round-robin, no priority. Do not rely on ordering.
- If **none** are ready and there's no `default`, `select` **blocks** until one becomes ready.
- `default` makes the select non-blocking: try every case, take `default` if none can proceed.
- `time.After(d)` returns a `<-chan time.Time` that fires once after `d`. The standard timeout idiom.
  - Heads-up: `time.After` allocates a timer that lives until it fires. In tight loops use `time.NewTimer(d)` so you can `Stop()` it and avoid leaks.

## The cancellation pattern (preview of `06-context-cancel`)

```go
for {
    select {
    case <-ctx.Done():
        return
    case job := <-jobs:
        process(job)
    }
}
```

A goroutine that processes work _and_ listens for cancellation. This is the workhorse pattern in production Go.

## Comparison

| Concept                 | Go                        | Python                                           | TS / Node                     |
| ----------------------- | ------------------------- | ------------------------------------------------ | ----------------------------- |
| Multiplex async sources | `select`                  | `asyncio.wait(..., return_when=FIRST_COMPLETED)` | `Promise.race(...)`           |
| Timeout                 | `case <-time.After(d):`   | `asyncio.wait_for(..., timeout=d)`               | `Promise.race([p, sleep(d)])` |
| Non-blocking try        | `select { default: ... }` | `q.get_nowait()`                                 | n/a                           |

## Run

```bash
go run .
```
