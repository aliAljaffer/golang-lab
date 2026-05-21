# 07 — Benchmarks

`testing` includes a benchmark runner. No separate tool. No separate framework.

## Things to notice

- **Naming.** `func BenchmarkXxx(b *testing.B)`. The `Benchmark` prefix + capital next letter is the discovery rule (just like `Test`).
- **`b.N`.** The runner sets it. You loop `b.N` times. The runner re-runs the benchmark with growing `N` until the per-op time stabilizes (default: run for 1 second).
- **`b.ResetTimer()`.** Excludes one-shot setup (e.g. `makeInput(10_000)`) from the measurement. Without it, allocation cost from setup gets divided across all `b.N` iterations and pollutes the result.
- **`-benchmem`.** Adds `allocs/op` + `B/op` columns. Crucial for spotting unnecessary allocations.
- **`b.Run(name, fn)`.** Same idea as `t.Run` for tests — sub-benchmarks. Use it to parametrize over input sizes.
- **`benchstat`.** External tool (`go install golang.org/x/perf/cmd/benchstat@latest`). Compares two `-bench` runs and reports a t-test on the delta. The right way to claim "X% faster" — eyeballing one number is a recipe for fooling yourself.

## Don't be misled

- Benchmarks measure wall-clock time on the test machine. Numbers from your laptop ≠ numbers from CI ≠ numbers from prod. Compare deltas, not absolutes.
- Run benchmarks multiple times. `-count=10` gives benchstat enough samples to be useful.
- The compiler might inline / dead-code-eliminate `_ = SumLoop(ns)` if the result is truly unused. Assigning to a package-level `var sink int` and writing to it in the loop is the classic dodge — for this example the inputs/outputs are non-trivial enough not to matter.

## Run

```bash
go test -bench=. ./06-testing/07-benchmark/...
go test -bench=. -benchmem ./06-testing/07-benchmark/...
go test -bench=. -count=10 ./06-testing/07-benchmark/... | tee a.txt
# (edit code, re-run, save to b.txt)
benchstat a.txt b.txt
```

## Comparison

| Concept      | Go          | Python                                      | TS / Node                               | Java         |
| ------------ | ----------- | ------------------------------------------- | --------------------------------------- | ------------ |
| Builtin      | `testing.B` | (none — use `pytest-benchmark` or `timeit`) | (none — use `tinybench`/`benchmark.js`) | JMH          |
| Stat compare | `benchstat` | `pytest-benchmark compare`                  | manual                                  | JMH baseline |
