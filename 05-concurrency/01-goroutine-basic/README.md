# 01 — Goroutine basics

The cheapest concurrent primitive Go gives you, and the gotcha that bites everyone the first time.

## Things to notice

- `go f()` runs `f` on a new goroutine and **returns immediately**. There is no handle, no Promise, no Future. If you want a result, you communicate via a channel.
- Goroutines are not OS threads. The runtime schedules thousands onto a small thread pool (`GOMAXPROCS` by default). Spawning 10k goroutines is normal Go; spawning 10k threads is a disaster in most languages.
- When `main` returns, the program exits — every goroutine dies mid-instruction. This is why naive loops that `go fmt.Println(...)` print nothing: main wins the race.
- Capture-by-reference gotcha (pre-Go 1.22): `for i := 0; i < 5; i++ { go func() { fmt.Println(i) }() }` prints `5 5 5 5 5` because the closure captures the *loop variable*. Pass `i` as an argument, or in Go 1.22+ rely on per-iteration scoping.

## Comparison

| Concept | Go | Python | TS / Node | Java |
|---|---|---|---|---|
| Spawn | `go f()` | `threading.Thread(target=f).start()` / `asyncio.create_task(f())` | `setImmediate(f)` / Promise | `new Thread(r).start()` |
| Wait | `sync.WaitGroup` | `t.join()` / `await gather(...)` | `await Promise.all(...)` | `t.join()` |
| Cost | ~2 KB stack, ms to spawn | MB stack | callback queue / event loop | MB stack |

## Run

```
go run .
```

## Stretch

- Spawn 100,000 goroutines that each sleep 1 second. Watch RSS — it should comfortably fit in a few hundred MB. Try the same with OS threads in any other language.
- Print `runtime.NumGoroutine()` before and after. Match expectation against reality.
