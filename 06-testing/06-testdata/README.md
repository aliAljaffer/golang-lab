# 06 — `testdata/` fixtures + helpers

Go convention for keeping fixture files alongside your tests.

## Things to notice

- **`testdata/` is magic.** The Go build tool skips it when building packages, so non-`.go` files (CSVs, JSON, YAML, binaries) don't cause compile errors. It IS still committed to git — `.gitignore` should NOT exclude it.
- **`t.Helper()`.** Mark assertion helpers with `t.Helper()` so failures point at the caller's line, not the helper's. Without it, `t.Fatal` in `fixture()` blames line 12 of `csv_test.go` — useless. With it, it blames whichever test called `fixture(t, ...)`.
- **`t.Cleanup(fn)`.** Runs `fn` after the test finishes, even if it failed. Replaces the old `defer` dance in helper functions (a `defer` in a helper fires when the helper returns, not when the test ends).
- **Working directory.** `go test` runs with the package directory as cwd. `filepath.Join("testdata", "foo")` always works — no need to compute an absolute path.

## When to use a fixture file vs a string literal

- **Fixture file:** the input is large, hand-edited often, or comes from a real-world example.
- **String literal in test:** the input is short and the test reads better with the data right next to the assertion.

Don't overuse fixtures. A 3-line CSV is clearer inline.

## Comparison

| Concept       | Go                                 | pytest                              | jest                              |
| ------------- | ---------------------------------- | ----------------------------------- | --------------------------------- |
| Fixture dir   | `testdata/` (magic, build-skipped) | `tests/fixtures/` (convention)      | `__fixtures__/` (Jest convention) |
| Helper marker | `t.Helper()`                       | (no equivalent — traceback is full) | (no equivalent)                   |
| Teardown      | `t.Cleanup(fn)`                    | yield in fixture / `addfinalizer`   | `afterEach`                       |

## Run

```bash
go test -v ./06-testing/06-testdata/...
```
