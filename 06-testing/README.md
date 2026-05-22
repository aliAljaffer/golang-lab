# 06 — Testing

> Status: ☑ scaffolded — examples + mini-project + exercises ready to implement. See [`PLAN.md`](./PLAN.md).

The section where everything earlier becomes *reliable*. Go ships with a complete testing toolkit in the stdlib — no separate framework to install, no mocking library required. The conventions are tight: files end in `_test.go`, functions start with `Test`, and the test runner discovers everything automatically.

The single most important pattern is the **table-driven test** — one function, a slice of `struct` cases, a `t.Run(name, func)` per case. Once it clicks, you'll write almost every test this way.

---

## What you'll learn

- Test file naming: `_test.go` suffix, same package or `_test` external package
- `func TestXxx(t *testing.T)` — discovered by convention; no registration ceremony
- `t.Errorf` vs `t.Fatalf` vs `t.Fatal` — when to fail-and-continue vs fail-and-stop
- Table-driven tests with a slice of struct cases + `t.Run` for each
- `t.Helper()` so assertion helpers report the *caller's* line, not the helper's
- `testdata/` — Go's convention for fixture files (the toolchain ignores it from build, not from VCS)
- Setup/teardown: `TestMain` for package-level, `t.Cleanup` for per-test (preferred over `defer`)
- Mocking via interfaces — define a small interface at the *consumption side* and the test injects a fake
- `httptest.NewServer` for testing HTTP clients without a real network
- Benchmarks: `func BenchmarkXxx(b *testing.B)` + `go test -bench=.`
- Fuzzing: `func FuzzXxx(f *testing.F)` + `go test -fuzz=Fuzz...` (built into Go 1.18+)
- Coverage: `go test -cover`, `-coverprofile`, `go tool cover -html`

---

## Mental model from other languages

| Concept            | Go                              | Python                              | TS / Node                  | Java                       |
| ------------------ | ------------------------------- | ----------------------------------- | -------------------------- | -------------------------- |
| Test runner        | `go test` (built-in)            | pytest                              | jest / vitest              | JUnit                      |
| Parametrized       | Table-driven (slice of structs) | `@pytest.mark.parametrize`          | `describe.each`            | `@ParameterizedTest`       |
| Mocking            | Interfaces, hand-rolled         | `unittest.mock`                     | jest mocks                 | Mockito                    |
| HTTP testing       | `httptest.NewServer`            | `responses` / `httpx_mock`          | `nock` / `msw`             | MockMvc                    |
| Coverage           | `go test -cover` (built-in)     | `coverage.py`                       | `--coverage` flag          | JaCoCo                     |
| Fuzz               | `go test -fuzz` (built-in 1.18+) | `hypothesis`                       | `fast-check`               | jqwik                      |

**Go's twist:** no test framework — `testing` is enough. No mocking framework — interfaces + structural typing make hand-rolled fakes a five-line affair. The community deliberately resists the dependency-injection-container, expect-library, BDD-runner sprawl that other ecosystems normalize. This is good: tests stay readable and don't break across framework upgrades.

---

## The DevOps angle

Testing infrastructure code is a force multiplier. Mock the AWS SDK behind an interface (section 07's example 07 demonstrates this pattern), use `httptest` to drive your webhook receiver through every signature-validation path, write table tests for your config parser. The patterns in this section recur in every later section's mini-project: every `S3API` / `DockerAPI` / `KubernetesAPI` / `NamespaceAPI` interface in this repo exists because of section 06.

**The pinned convention: small interfaces, defined at the consumption side, not the SDK side.** Don't try to mock `*s3.Client` — define `type S3API interface { ListObjectsV2(...) ... }` with exactly the three methods your code uses, then your code accepts `S3API` and the test passes a `fakeS3` struct. Same pattern in every section from 07 onward.

---

## Walkthrough

Read these in order. Each is a runnable example demonstrating one specific concept. **Note: unlike sections 01-05, the examples here come fully implemented** — the lesson *is* the testing pattern, so the working test code is the artefact.

1. [`01-basic-test/`](./01-basic-test/) — one function, one test; happy path + error path. The `if got != want { t.Errorf("...") }` shape every test starts from.
2. [`02-table-driven/`](./02-table-driven/) — same logic, parametrized via a `[]struct{ name, in, want }` slice + `for _, tc := range cases { t.Run(tc.name, ...) }`. Idiomatic Go.
3. [`03-subtests/`](./03-subtests/) — `t.Run("named-case", func(t *testing.T) {...})`. Why naming subtests matters: `go test -run TestX/specific-case` filters to one case.
4. [`04-mock-interface/`](./04-mock-interface/) — a consumer that takes an interface; tests inject a fake. The pattern every later section reuses.
5. [`05-httptest/`](./05-httptest/) — `httptest.NewServer(handler)` for testing HTTP clients without a real network. Pair with the retry/backoff client from section 03.
6. [`06-testdata/`](./06-testdata/) — load fixture files from `testdata/`. The directory is sacred — the Go toolchain ignores it when building but not when testing.
7. [`07-benchmark/`](./07-benchmark/) — a function with two implementations; benchmark both with `for i := 0; i < b.N; i++`. `go test -bench=.` runs them; `-benchmem` adds allocations.
8. [`08-fuzz/`](./08-fuzz/) — fuzz a parser to find unexpected panics. Seed the corpus with `f.Add(...)`, run `go test -fuzz=FuzzX -fuzztime=30s`, watch it find edge cases you didn't write.

---

## Mini-project: [`logstats`](./mini-project/)

A small log aggregator that intentionally has surface area for every testing pattern from examples 01-08: a file-or-URL source, a parser to test with fixtures, a formatter to table-test, an HTTP fetch path to drive with `httptest`, and a hotspot to benchmark/fuzz.

The point: every pattern in this section gets exercised on one realistic codebase. Spec and starter in [`mini-project/`](./mini-project/).

---

## Exercises

See [`exercises/`](./exercises/):

1. **[`01-table-tests`](./exercises/01-table-tests/)** — convert a function tested with 5 separate `Test*` functions into a single table test. The refactor that makes the rest of the section click.
2. **[`02-fake-clock`](./exercises/02-fake-clock/)** — given code that uses `time.Now()`, refactor it to take a `Clock` interface; test time-dependent behavior deterministically. The pattern reused by `08-kubernetes/mini-project` (the dedup `Now func() time.Time` injection) and `10-observability/exercises/02-rate-limited-logging`.
3. **[`03-coverage-gap`](./exercises/03-coverage-gap/)** — given a `.go` file with hidden uncovered branches, write tests until `go test -cover` reports 100%. Teaches `go tool cover -html=coverage.out` for visual gap-hunting.

---

## Further reading

- [`testing` package docs](https://pkg.go.dev/testing) — the complete API surface, including `T.Cleanup`, `T.Setenv`, `T.TempDir`
- [Go Wiki: TableDrivenTests](https://go.dev/wiki/TableDrivenTests) — the canonical pattern, explained by the language team
- [Dave Cheney: "Prefer table driven tests"](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests) — the rationale and idioms
- [Go Blog: Fuzzing in Go 1.18](https://go.dev/blog/fuzz-beta) — the introduction with worked examples
- [`httptest` docs](https://pkg.go.dev/net/http/httptest) — the stdlib HTTP testing helpers
