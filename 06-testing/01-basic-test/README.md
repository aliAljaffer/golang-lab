# 01 — Basic test

The smallest possible Go test, and the two failure verbs.

## Things to notice

- File name must end in `_test.go`. The compiler excludes these from regular builds.
- Function must be `func TestXxx(t *testing.T)`. The `Test` prefix + capital next letter is how `go test` discovers it.
- Same package as the code → tests can see unexported identifiers. Use a `_test` package suffix (`package calc_test`) for black-box tests when you want to assert against the public API only.
- `t.Errorf` — record a failure, keep running this test. Use when more assertions in the same test add diagnostic value.
- `t.Fatalf` — record a failure, abort this test. Use when continuing would panic (e.g. dereferencing a nil pointer after `New()` failed).
- Tests in the same file are independent — one `t.Fatalf` doesn't stop the others.

## Run

```
go test ./06-testing/01-basic-test/...
go test -v ./06-testing/01-basic-test/...   # show each test name + PASS/FAIL
go test -run TestAdd ./06-testing/01-basic-test/...  # filter by name
```

## Comparison

| Concept | Go | pytest | jest |
|---|---|---|---|
| Test file | `foo_test.go` | `test_foo.py` | `foo.test.ts` |
| Test func | `func TestX(t *testing.T)` | `def test_x():` | `test('x', () => {})` |
| Assert | `if got != want { t.Errorf(...) }` | `assert got == want` | `expect(got).toBe(want)` |
| Continue on fail | `t.Errorf` | (default) | (default) |
| Abort on fail | `t.Fatalf` | (default) | (use `expect().toThrow`) |

No assertion library is idiomatic — plain `if` + `t.Errorf` is the convention.
