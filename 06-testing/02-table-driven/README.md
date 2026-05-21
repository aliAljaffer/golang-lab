# 02 — Table-driven tests

The idiomatic Go testing style. Once you internalize this, you won't write tests any other way.

## Things to notice

- The struct is declared inline — anonymous, scoped to the test. No need to name a `testCase` type unless multiple tests share it.
- `name` field gives each case a label that appears in error output. Pick names that read well in failure messages.
- The loop body is the assertion logic, written ONCE. Adding a 20th case is a one-line addition to the slice.
- A failing case names itself in the error (`%s: Reverse(%q) = ...`) so you can tell which row broke without re-reading code.
- This is still ONE `Test*` function — every case runs even if earlier ones fail (because we use `t.Errorf`, not `t.Fatalf`).

## When NOT to use a table

- The cases differ in setup/teardown (e.g. some need a tmpdir, others don't). Different shapes → different tests.
- One case is so subtle it deserves its own test name + comment. Pull it out.

## Comparison

| Concept         | Go (table)                      | pytest                     | jest                          |
| --------------- | ------------------------------- | -------------------------- | ----------------------------- |
| Idiom           | `[]struct{...}` + for loop      | `@pytest.mark.parametrize` | `describe.each` / `test.each` |
| Per-case naming | `name` field used in `t.Errorf` | id= or auto-generated      | template literal              |

## Next

The next example (`03-subtests`) shows the upgrade: wrap the loop body in `t.Run(tc.name, ...)` so each case becomes its own runnable + filterable subtest.

## Run

```bash
go test ./06-testing/02-table-driven/...
go test -v ./06-testing/02-table-driven/...
```
