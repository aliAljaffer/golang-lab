# 04 — Mocking via interfaces

Go has no mocking framework in its stdlib, and the community has not converged on one. The reason: interfaces + structural typing make manual fakes trivial.

## The pattern

1. The consumer declares a tiny interface with only the methods IT calls (`Notifier` here — one method).
2. Production code passes a real impl (an SMTP client, an SNS publisher).
3. Tests pass a fake — usually a struct with a recording slice + an optional `failOn` switch.

The fake is co-located with the test that uses it. Do not export it. Do not put it in a shared "test helpers" package — that just creates coupling.

## Why "interface at the consumer"

A common mistake (especially for Java/C# folks) is to define a giant `INotifier` interface in the package that *implements* it, then have the consumer depend on it. That makes the interface a contract the producer has to maintain, even if a consumer only needs one method.

The Go convention is the reverse: **define interfaces where they're used.** The producer just returns a concrete type. The consumer declares the small subset it needs and accepts anything that satisfies it. Structural typing does the rest — no `implements` keyword needed.

## When NOT to mock

- Pure functions (no side effects): just call them with input, assert output. No interface needed.
- Stdlib types with built-in test doubles: `bytes.Buffer` is an `io.Writer`. `strings.NewReader` is an `io.Reader`. `httptest.NewServer` is a real server. Don't mock what stdlib already gives you a fake for.

## Comparison

| Concept | Go | Python | TS/Node | Java |
|---|---|---|---|---|
| Fake | struct with recording slice | `unittest.mock.MagicMock` | `jest.fn()` | Mockito `mock(X.class)` |
| Verify call | inspect the slice | `m.assert_called_with(...)` | `expect(fn).toHaveBeenCalledWith(...)` | `verify(m).method(...)` |
| Lines of code | ~10 | ~1 | ~1 | ~1 |

Yes, Go's version is more lines. The payoff is that the fake is type-checked, refactor-safe, and obvious — there's no magic argument matching or call-record DSL to learn.

## Run

```
go test ./06-testing/04-mock-interface/...
```
