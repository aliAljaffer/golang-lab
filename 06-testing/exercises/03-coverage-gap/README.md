# Exercise 03 — Coverage gap

Write tests until `go test -cover` reports 100%.

## What's here

- `validate.go` — `Validate(c Config) error` with several branches.
- `validate_test.go` (`//go:build exercise`) — 2 starter tests + a TODO list of branches you still need to cover.

## Your job

1. Run baseline coverage:
   ```
   go test -tags=exercise -cover ./06-testing/exercises/03-coverage-gap/...
   ```
   Note the percentage.
2. Generate the HTML report to see which lines aren't hit:
   ```
   go test -tags=exercise -coverprofile=/tmp/c.out ./06-testing/exercises/03-coverage-gap/...
   go tool cover -html=/tmp/c.out
   ```
   Red lines are uncovered.
3. Add tests, one per uncovered branch, until you see `coverage: 100.0% of statements`.

## Things to notice

- **Coverage tells you "did this line execute," not "did the test assert on it."** A test that runs `Validate(c)` and ignores the error still bumps coverage. Always pair coverage with a real assertion — otherwise you have busywork lines.
- **One branch may need multiple tests.** The `port < 1024 && env == "prod"` rule is a single line of code but needs different inputs from "env == prod with port 8080" to actually hit it. Coverage tools highlight the line as red until BOTH conditions of the AND happen to be true.
- **100% is a checkpoint, not a goal.** Hitting 100% on `validate.go` says nothing about whether the rules themselves are correct. After you hit 100%, look at the rules again — are any redundant? Missing? Off by one?

## Verify

```
go test -tags=exercise -cover ./06-testing/exercises/03-coverage-gap/...
```

Goal: `coverage: 100.0% of statements`.

## Stretch

- Convert your individual tests into a single table-driven test with subtests, the same way you did in exercise 01. Compare line count.
- Add a `FuzzValidate(f *testing.F)` that generates random Config values and asserts `Validate` never panics, regardless of input.
