# 06 — Exit codes, stderr, and `log.Fatal`

CLI hygiene that scripts depend on:

## Streams

| Stream | What goes here |
|---|---|
| stdout (`fmt.Println`) | The **data** the user / next pipeline stage consumes |
| stderr (`fmt.Fprintln(os.Stderr, ...)`) | Errors, progress, diagnostics |

If you `tool | jq`, jq only sees stdout. Progress bars and warnings must go to stderr or they corrupt the pipe.

## Exit code conventions

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | Generic runtime error |
| `2` | Usage error (bad flags, wrong arg count) |
| `126`, `127`, `130` | Reserved by the shell (not your code) |

`make`, CI systems, and your own scripts branch on these. Returning `0` from a failed run is one of the most insidious bugs in DevOps tooling — silent failures.

## The `log.Fatal` trap

```go
func main() {
    f, _ := os.Create("out.txt")
    defer f.Close()           // looks safe...
    log.Fatal("boom")          // calls os.Exit(1) — defer DOES NOT RUN
}
```

`log.Fatal` and `log.Fatalf` are fine for `main` when you have nothing to clean up. Inside libraries or anywhere with `defer`, return an `error` instead.
