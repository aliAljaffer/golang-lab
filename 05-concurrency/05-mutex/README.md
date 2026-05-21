# 05 — Mutex

When channels aren't the natural fit, `sync.Mutex` protects shared memory the boring way.

## When to use a mutex over channels

The Go proverb is "share memory by communicating, not communicate by sharing memory." But mutexes are still the right answer when:

- You're guarding a small piece of in-memory state (a counter, a cache, a map).
- The state is read/written from many goroutines and you have no natural producer/consumer split.
- Using a channel would require a "manager goroutine" that just serializes access — same effect, more code.

If you're tempted to write a manager goroutine that holds a map and answers requests over channels, just use a mutex.

## Things to notice

- **`defer mu.Unlock()`** on the line after `Lock()`. An early return or panic that skips Unlock leaves the mutex permanently locked = deadlock.
- A mutex is **not** reentrant. The same goroutine cannot `Lock` twice — it deadlocks.
- Zero value works: `var mu sync.Mutex` is usable immediately. No `New`.
- Embed it as the **first field** in your struct and document the invariants the lock protects. The goal: a reader can see at a glance which fields are guarded.

## RWMutex

`sync.RWMutex` allows multiple concurrent readers OR one writer. Use it when reads vastly outnumber writes (config caches, route tables). Don't reach for it by default — the bookkeeping makes it slower than `Mutex` under low contention.

## The race detector

```bash
go run -race .
go test -race ./...
```

`-race` instruments memory accesses and prints a report if two goroutines touch the same address without synchronization (and at least one is a write). It is the single most useful Go flag you can run. Use it in CI.

## Comparison

| Concept        | Go                       | Python                | TS / Node                | Java                         |
| -------------- | ------------------------ | --------------------- | ------------------------ | ---------------------------- |
| Mutex          | `sync.Mutex`             | `threading.Lock`      | (rare — single-threaded) | `ReentrantLock` (reentrant!) |
| RWMutex        | `sync.RWMutex`           | `threading.RLock`-ish | n/a                      | `ReentrantReadWriteLock`     |
| Concurrent map | `sync.Map` (specialized) | `dict` + `Lock`       | `Map` (single-threaded)  | `ConcurrentHashMap`          |

## Run

```bash
go run .
go run -race .
```
