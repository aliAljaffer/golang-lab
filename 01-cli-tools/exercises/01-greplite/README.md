# Exercise 01 — `greplite`

Implement a minimal grep as a **library function**, not a full CLI. Tests drive the contract.

## What to build

In `greplite.go`, fill in:

```go
func Grep(input io.Reader, pattern string, opts Options) ([]Match, error)
```

## Behaviour

- Scan `input` line by line with `bufio.Scanner`.
- Return one `Match` per line containing `pattern` as a substring.
- `Options.IgnoreCase`: compare case-insensitively (`strings.ToLower` on both sides, or `strings.EqualFold` — but you need _contains_, so think about it).
- `Options.LineNumbers`: populate `Match.LineNumber` (1-based). When false, leave it at `0`.
- Empty pattern matches every line.
- Strip the trailing newline from `Match.Text` (Scanner does this for you).

## Run

```bash
go test ./01-cli-tools/exercises/01-greplite/...
```

## Stretch

- Wrap it in a cobra CLI that accepts a file path or reads stdin.
- Add `-v` (invert match) and `-c` (count only).
- Replace substring match with `regexp.Compile`.
