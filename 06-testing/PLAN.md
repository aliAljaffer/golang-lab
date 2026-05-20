# Plan: 06-testing

## Concepts to cover

- [ ] Test file naming: `_test.go` suffix, same package or `_test` external package
- [ ] `func TestXxx(t *testing.T)` — discoverable by convention
- [ ] `t.Errorf` vs `t.Fatalf` vs `t.Fatal` — when each is right
- [ ] Table-driven tests with a slice of struct cases + `t.Run` for each
- [ ] `t.Helper()` to make assertion helpers report the right line
- [ ] `testdata/` — Go's convention for fixture files (gitignored from build, not from VCS)
- [ ] Setup/teardown: `TestMain` for package-level, `t.Cleanup` for per-test
- [ ] Mocking via interfaces — define small interfaces at the consumption side
- [ ] `httptest.NewServer` for HTTP client tests
- [ ] Benchmarks: `func BenchmarkXxx(b *testing.B)` + `go test -bench=.`
- [ ] Fuzzing: `func FuzzXxx(f *testing.F)` + `go test -fuzz=Fuzz...`
- [ ] Coverage: `go test -cover`, `-coverprofile`, `go tool cover -html`

## Examples to build

| Folder | Demonstrates |
|---|---|
| `01-basic-test/` | One func, one test; happy path + error path |
| `02-table-driven/` | Same logic, parametrized via a `[]struct{}` |
| `03-subtests/` | `t.Run` with named cases |
| `04-mock-interface/` | A consumer that takes an interface; tests inject a fake |
| `05-httptest/` | Use `httptest.NewServer` to test a client that calls it |
| `06-testdata/` | Load fixture files from `testdata/` |
| `07-benchmark/` | A function with two implementations; benchmark both |
| `08-fuzz/` | Fuzz a parser to find unexpected panics |

## Mini-project

Add tests retroactively to two earlier mini-projects:
- `01-cli-tools/mini-project` (`dirsize`) — table tests for the size formatter, integration test for the walker
- `03-http-clients/mini-project` (`gh-repo-stats`) — httptest-based tests for the cache + retry logic

Verifies that the testing patterns work on real code, not just toy examples.

## Exercises

1. **`01-table-tests`** — convert a given function with 5 separate `Test*` functions into a single table test
2. **`02-fake-clock`** — given code that uses `time.Now()`, refactor it to take a `Clock` interface; test time-dependent behavior deterministically
3. **`03-coverage-gap`** — given a `.go` file with hidden uncovered branches, write tests until coverage is 100%

## Status

- [ ] Concepts in README walkthrough
- [ ] Examples 01-08 built
- [ ] Mini-project: tests added to dirsize + gh-repo-stats
- [ ] Exercises scaffolded
