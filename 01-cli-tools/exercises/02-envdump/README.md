# Exercise 02 — `envdump`

Print env vars whose **keys** match a glob, with an `--unset` mode that clears matches.

## What to build

Two pure functions in `envdump.go`:

```go
func Match(env []string, pattern string) ([]Pair, error)
func UnsetMatching(env []string, pattern string, unsetter Unsetter) ([]string, error)
```

- `env` is in `KEY=VALUE` shape — same as `os.Environ()`.
- `pattern` uses `path/filepath.Match` (shell-style globs).
- Match against the **key only** (split on the first `=`).
- Preserve input order in the result.
- `UnsetMatching` takes an `Unsetter` interface, not `os.Unsetenv` directly — this is what makes the tests deterministic. (In a real CLI you'd inject `osUnsetter{}` from `main`.)

## Run

```bash
go test ./01-cli-tools/exercises/02-envdump/...
```

## Stretch

- Wrap with cobra: `envdump 'APP_*'` and `envdump --unset 'TMP_*'`.
- Add `--values-only` to print just the values (useful in shell expansion).
- Read env from a `.env` file instead of the live environment.
