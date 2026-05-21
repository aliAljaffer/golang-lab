# 08 — Fuzzing

Built into Go's `testing` package since 1.18. Generates inputs, looks for crashes + invariant violations.

## Things to notice

- **`func FuzzXxx(f *testing.F)`** — naming convention, like Test/Benchmark.
- **Seed corpus** — `f.Add(input)` registers known-good inputs. The fuzzer uses these as starting points for mutation.
- **`f.Fuzz(func(t *testing.T, ...) { ... })`** — the target. The args after `t` are the typed inputs the fuzzer will generate. Supported types: `string`, `[]byte`, all int/uint, float32/64, bool, rune.
- **Invariant, not assertion.** A fuzz target doesn't check a fixed expected output (the input is random — you don't know what to expect). Instead, you check a property: "if accepted, the answer agrees with stdlib", "round-tripping is lossless", "no panic regardless of input".
- **Two modes:**
  - `go test ./...` (no `-fuzz`) — runs seed corpus as ordinary test cases. Fast. Deterministic. Runs in CI.
  - `go test -fuzz=FuzzParseInt -fuzztime=30s ./08-fuzz/...` — interactive fuzzing. Mutates inputs for up to 30s, reports first failure.
- **Failure persistence.** When the fuzzer finds a crash, it writes the input to `testdata/fuzz/<FuzzName>/<hash>`. From then on, plain `go test` replays that input as a regression case. So the fuzzer doubles as a way to grow the test suite.

## The bug in this example

`ParseInt("-")` returns `0, nil` instead of an error. Look at `parse.go` — the loop body never runs when the input is just `"-"`. The fuzzer will find this within seconds: it generates `"-"`, `ParseInt` returns `0, nil`, `strconv.Atoi("-")` returns an error, the invariant fails.

## Why it took a fuzzer to catch this

Most humans writing the table test would seed it with `"0"`, `"123"`, `"-42"`, `""`, `"abc"` — typical inputs. They wouldn't think to add `"-"` as an edge case. The fuzzer thinks of all the edge cases by _not thinking_ — it just mutates and tries.

## Run

```bash
# normal mode — seed corpus only
go test ./06-testing/08-fuzz/...

# fuzzing mode — generates random inputs
go test -fuzz=FuzzParseInt -fuzztime=10s ./06-testing/08-fuzz/...
```

After the fuzzer crashes, look in `testdata/fuzz/FuzzParseInt/`. The file is the persisted failing input. Fix the bug, the new test (which is just `go test`) replays that input — green from then on.

## Comparison

| Concept        | Go           | Python                                              | TS / Node                          | Java                  |
| -------------- | ------------ | --------------------------------------------------- | ---------------------------------- | --------------------- |
| Builtin fuzzer | yes (1.18+)  | no — use `hypothesis` (property tests) or `atheris` | no — `fast-check` (property tests) | `jqf` / `JQF`         |
| Seed corpus    | `f.Add(...)` | `@given(...)` strategies                            | `fc.assert(fc.property(...))`      | regression test class |
