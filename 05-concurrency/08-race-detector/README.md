# 08 — Race detector

`-race` is the most important Go flag you'll ever use. Run it. Always.

## What a race actually is

Two goroutines access the same memory location, **at least one of them writes**, and there is no synchronization (channel send/recv, mutex, atomic) ordering them. The result is undefined: torn reads, lost writes, "impossible" branches. Programs _may appear_ to work and then fail in production at the worst time.

## Try it

```bash
go run .              # output is undefined; may look fine
go run -race .        # WARNING: DATA RACE with two stack traces
```

The detector instruments memory accesses at compile time, tracks the happens-before relation at runtime, and reports any pair that races. The cost is real (~5–10× slowdown, ~2× memory) — that's why it's not on by default. But for `go test`, the cost is fine and the payoff is enormous.

## Where to use it

| Where       | How                                                    |
| ----------- | ------------------------------------------------------ |
| Local dev   | `go run -race ./cmd/foo`                               |
| Tests       | `go test -race ./...`                                  |
| CI          | always — make it required on PRs                       |
| Prod binary | **no** — too slow; ship binaries built without `-race` |

## Common races the detector catches

- Concurrent writes to a map (also panics: `concurrent map writes`)
- Concurrent read+write of any field without a lock
- `wg.Add` from inside the goroutine (Add can happen after Wait)
- Closing a channel while another goroutine sends
- Writing to a slice header from multiple goroutines (append re-slicing!)

## Common races the detector does _not_ catch

- Logic races that don't touch the same address (e.g. "I expected X to finish before Y")
- Races that simply never trigger in your test run — the detector reports observed races, not theoretical ones. Higher coverage = more confidence.

## Run

```bash
go run .
go run -race .
```
