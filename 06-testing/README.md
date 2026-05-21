# 06 — Testing

> Status: ✅ scaffolded — examples + exercises + mini-project ready. Walkthrough doc still TODO.

## What you'll learn

- The stdlib `testing` package (no separate framework needed)
- **Table-driven tests** — the idiomatic Go testing style
- Subtests with `t.Run` for organization + selective running
- Test fixtures with the `testdata/` directory
- Mocking via interfaces (no mocking framework needed)
- `httptest` for HTTP testing without spinning up real servers
- Benchmarks (`go test -bench`)
- Fuzzing (`go test -fuzz` — built into Go 1.18+)

## Mental model from other languages

| Concept | Go | Python | TS / Node | Java |
|---|---|---|---|---|
| Test runner | `go test` (built-in) | pytest | jest / vitest | JUnit |
| Parametrized | Table-driven (slice of structs) | `@pytest.mark.parametrize` | `describe.each` | `@ParameterizedTest` |
| Mocking | Interfaces, hand-rolled | `unittest.mock` | jest mocks | Mockito |
| HTTP testing | `httptest.NewServer` | `requests-mock` | nock / msw | MockMvc |

**Go's twist:** no test framework — `testing` is enough. No mocking framework — interfaces + structural typing make manual mocks trivial.

## The DevOps angle

Testing infra code is a force multiplier. Mock the AWS SDK with an interface, use `httptest` to test webhook receivers, write table tests for your config parser. This section is where everything earlier becomes *reliable*.

See [`PLAN.md`](./PLAN.md).
