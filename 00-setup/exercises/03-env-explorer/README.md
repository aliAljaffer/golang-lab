# Exercise 03 — Env explorer

A code exercise. Implement a function that looks up Go environment variables programmatically.

## Goal

Write `GoEnv(key string) (string, error)` that returns the value of a Go env variable (the same values `go env <KEY>` would print).

## Steps

1. Open `starter.go` — see the function signature and TODO.
2. Run the tests (they should fail):
   ```bash
   go test -tags=exercise ./00-setup/exercises/03-env-explorer/
   ```
3. Implement `GoEnv`. Hint: shell out to `go env <KEY>` using `os/exec.Command`.
4. Run the tests again — they should pass.

## Concepts

- Running subprocesses with `os/exec.Command`
- Capturing stdout
- Returning errors (Go's `error` type is just an interface; functions return `(value, error)`)
- Handling the case where `go` isn't on `$PATH`

## Hints

- `exec.Command("go", "env", key).Output()` returns `([]byte, error)`
- The output has a trailing newline — `strings.TrimSpace` it
- If the key doesn't exist, `go env` prints an empty string (not an error)
