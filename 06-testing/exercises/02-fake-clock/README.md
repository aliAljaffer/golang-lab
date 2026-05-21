# Exercise 02 — Fake clock

Refactor time-dependent code so tests can pin "now" instead of sleeping.

## What's here

- `scheduler.go` — a `Scheduler` with `ShouldFire(name, cooldown)` + `Fire(name)`. The `Now func() time.Time` seam is already declared on the struct and initialized in `New()`, but the two methods still call `time.Now()` directly. That's the bug.
- `scheduler_test.go` (`//go:build exercise`) — five tests that use a fake clock. `TestShouldFire_DefaultsToRealClock` passes already; the other four fail because `ShouldFire`/`Fire` ignore the fake `Now`.

## Your job

Replace both `time.Now()` calls inside `ShouldFire`/`Fire` with `s.Now()`, then re-run:

```bash
go test -tags=exercise ./06-testing/exercises/02-fake-clock/...
```

That's literally the whole exercise — a 2-character change. The point is the _thinking_ around it, not the keystrokes: why a function-value field beats an interface here, how it lets tests pin time without sleeping, and why every code path that uses `time.Now`/`time.Sleep`/`rand.Read`/`os.Getenv` should go through an injectable seam.

## Why `func() time.Time` instead of a `Clock` interface

A bigger codebase that needs to fake `Sleep`, `After`, `Tick`, and `Now` will want a `Clock` interface (e.g. [`github.com/benbjohnson/clock`](https://github.com/benbjohnson/clock)). For "just fake Now," a struct field of type `func() time.Time` is lighter — no interface, no second type, no constructor ceremony. Pick the smallest seam that solves your problem.

## What you're learning

- **Time is a dependency.** Code that talks to `time.Now()` (or `time.Sleep`, or `rand.Read`, or `os.Getenv`) is depending on global state. Inject it the same way you inject everything else.
- **The bug it prevents:** without a fake clock, the cooldown test would have to `time.Sleep(time.Hour)` to verify "after cooldown" — obviously infeasible. The test would either get watered down ("Sleep(50ms), assert false") or get skipped entirely.

## Verify

```bash
go test -tags=exercise -v ./06-testing/exercises/02-fake-clock/...
```

All five tests should pass. The whole suite should take milliseconds — no real time-based waiting anywhere.
