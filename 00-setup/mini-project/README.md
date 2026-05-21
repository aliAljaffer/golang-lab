# Mini-project — `gostat`

A CLI that prints info about a Go module: module path, declared Go version, dependency count, runtime GOOS/GOARCH.

## Run (once implemented)

```bash
go run ./00-setup/mini-project /path/to/go.mod
go run ./00-setup/mini-project --json /path/to/go.mod
```

Example output:

```bash
module:   github.com/alialjaffer/golang-lab
go:       1.22
deps:     0
runtime:  darwin/arm64
```

JSON mode:

```json
{
  "module": "github.com/alialjaffer/golang-lab",
  "go": "1.22",
  "deps": 0,
  "goos": "darwin",
  "goarch": "arm64"
}
```

## Your task

Open `main.go`. Replace the `TODO` stubs.

You'll need to:

1. Implement `parseGoMod(content []byte) (info, error)` — extract `module` and `go` lines, count `require` blocks
2. Wire `main` to: parse flags (`--json`), read the file from argv[1], call `parseGoMod`, print

## Verify

Tests for this exercise live behind the `exercise` build tag so they don't break the default test run.

```bash
go test -tags=exercise ./00-setup/mini-project/
```

When all tests pass, you're done. Then optionally:

```bash
go run ./00-setup/mini-project ./go.mod   # try it on this repo's own go.mod
```

## Concepts you'll touch

- Reading files (`os.ReadFile`)
- Parsing line-by-line (`bufio.Scanner` or `strings.Split`)
- Defining a struct
- `flag` package for `--json`
- `encoding/json` for the JSON output
- `fmt.Printf` for the human output
