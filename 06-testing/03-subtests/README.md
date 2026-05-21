# 03 — Subtests with `t.Run`

The table from example 02, but each case gets its own subtest name. Three things this buys you.

## Things to notice

- **Named cases in output.** `go test -v` prints `TestRepeat/zero`, `TestRepeat/single`, etc., one PASS/FAIL line per case.
- **Filtering.** `go test -run TestRepeat/zero` runs ONLY that case. The `-run` flag matches against `parent/child`. Pattern is a regex: `-run TestRepeat/single|twice` runs both.
- **Per-case parallelism.** Adding `t.Parallel()` as the first line of a subtest tells the runner this case can run concurrently with sibling subtests. The parent test waits for them all before completing.

## Subtle gotcha

Pre-Go-1.22: when using `t.Parallel()` inside a `for _, tc := range ...` loop, you had to do `tc := tc` to rebind the loop variable per iteration. Otherwise all subtests would share the same `tc`. Go 1.22+ fixes this with per-iteration loop scoping, but you'll see `tc := tc` in older code — that's why.

## When to use subtests vs a plain loop

- **Subtests:** when you want case-by-case names, filtering, or parallelism. Default to this — the cost is one extra line per case.
- **Plain loop:** quick smoke checks where the case index in the failure message is sufficient.

## Run

```
go test -v ./06-testing/03-subtests/...
go test -v -run TestRepeat/single ./06-testing/03-subtests/...
```
