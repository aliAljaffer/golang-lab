# Exercise 01 — Table tests

Convert five separate tests into a single table-driven test.

## What's here

- `classify.go` — `Classify(score int) string` returns a letter grade.
- `classify_separate_test.go` — the "before" state: five separate `Test*` functions, all passing. This is what you're replacing.
- `classify_table_test.go` (gated by `//go:build exercise`) — the "after" state: a single `TestClassify_Table` with an empty case slice. Fill it in.

## Your job

1. Run `go test -tags=exercise ./06-testing/exercises/01-table-tests/...`. It fails: `TODO: add cases`.
2. Fill in `tests` with enough cases to cover EVERY branch of `Classify`, including boundaries (89/90, 79/80, etc.).
3. Run again until green.
4. Delete `classify_separate_test.go` — the table version makes it redundant.

## What to think about

- **Boundary coverage matters.** `Classify(90)` and `Classify(89)` exercise *different* branches. A test of only `Classify(95)` and `Classify(85)` would pass even if the code said `case score >= 91` (a bug). Always include the boundary values.
- **Name the cases for readable failures.** A case named `"90 is A"` says more in the failure output than `"case 3"`.
- **One table = one assertion shape.** If you find yourself wanting `if/else` inside the loop, you've outgrown a table and should split the test.

## Verify

```
go test -tags=exercise -v ./06-testing/exercises/01-table-tests/...
```

You should see `TestClassify_Table` with one PASS line per subtest, AND no remaining `TestClassify_A`/`B`/`C`/`F`/`OutOfRange` lines (because you deleted `classify_separate_test.go`).
