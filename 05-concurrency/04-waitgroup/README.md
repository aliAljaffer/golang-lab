# 04 — WaitGroup

`sync.WaitGroup` is a counter that goroutines decrement, and one (usually main) blocks on.

## Things to notice

- Call `wg.Add(N)` *before* the `go` statement, not inside it. Otherwise `wg.Wait()` may run before `Add` and return immediately — a classic race.
- Use `defer wg.Done()` as the first line in the goroutine so an early return or panic still decrements the counter.
- A WaitGroup is *not* for collecting results. If you need return values, use a channel (or a results slice with a mutex). WaitGroup just answers "are they all done?"
- Calling `wg.Done()` more times than `wg.Add(N)` accumulated **panics** (negative counter).

## When to use what

| Need | Use |
|---|---|
| "Wait for N goroutines to all finish" | `sync.WaitGroup` |
| "Wait for the *first* of N to finish" | `select` over N result channels |
| "Wait for N with results" | results channel + WaitGroup, or `errgroup.Group` |
| "Wait for cancellation" | `<-ctx.Done()` |

## errgroup preview

`golang.org/x/sync/errgroup` is the production upgrade: it's a WaitGroup that also collects the first non-nil error and cancels a shared context. You'll meet it again in real code.

## Comparison

| Concept | Go | Python | TS / Node | Java |
|---|---|---|---|---|
| Wait for N | `wg.Wait()` | `await asyncio.gather(*tasks)` | `await Promise.all([...])` | `CompletableFuture.allOf(...).join()` |

## Run

```
go run .
```
