# Plan: 00-setup

## Concepts to cover

- [ ] Installing Go (mac/Linux/version managers)
- [ ] `GOROOT`, `GOPATH`, `GOBIN`, `GOMODCACHE` — what each means, why modules made `GOPATH` mostly irrelevant
- [ ] The Go CLI daily commands: `run`, `build`, `install`, `fmt`, `vet`, `test`, `mod init`, `mod tidy`, `get`, `doc`, `env`
- [ ] Anatomy of `go.mod` and `go.sum` — and why both are committed
- [ ] What every Go file looks like: `package`, `import`, `func main()`
- [ ] Project layout philosophy: when to use `cmd/`, `internal/`, `pkg/` (and when not to)
- [ ] Editor setup: gopls, format-on-save
- [ ] Tooling beyond stdlib: `goimports`, `golangci-lint`

## Examples to build

| Folder | Demonstrates |
|---|---|
| `01-hello-world/` | Smallest possible Go program; introduces `package main` + `func main()` |
| `02-go-run-vs-build/` | Same code, two ways to execute; inspect the binary |
| `03-modules-and-deps/` | Init a module, add a dep, observe go.mod/go.sum changes |
| `04-go-env-tour/` | Tiny program that prints what `go env` shows; explains each var |

## Mini-project

**`gostat`** — a CLI that prints info about a Go installation and the current module (Go version, module path, dep count, GOOS/GOARCH). Reads `go.mod` directly.

Tests verify it outputs valid JSON with the expected keys when run with `--json`.

## Exercises

1. **`01-tidy-experiment`** — given a `main.go` that imports a package not in `go.mod`, write code/run commands until `go test` passes (test checks that `go.mod` contains the dep)
2. **`02-static-binary`** — write a hello program, build it, write a test that asserts `file <binary>` shows it's statically linked
3. **`03-env-explorer`** — implement a function that returns specific `go env` values; tests pin expected keys

## Status

- [x] Concepts documented in README.md
- [x] Examples 01-04 built (01-hello-world, 02-go-run-vs-build, 03-modules-and-deps, 04-go-env-tour)
- [x] Mini-project `gostat` scaffolded (starter + tests; reader implements)
- [x] Exercises scaffolded (01-tidy CLI walkthrough, 02-static-binary CLI walkthrough, 03-env-explorer code exercise)

## Connections (for the section README)

- `go.mod` + `go.sum` ≈ `package.json` + `package-lock.json` (TS) / `pyproject.toml` + `poetry.lock` (Python) / `pom.xml` (Java)
- `go run` ≈ `npx tsx` / `python -m`
- `go install` ≈ `npm install -g` / `pipx install`
- `go fmt` is non-negotiable — no Prettier/Black debate; the tool ships with the language
- Single static binary output ≈ `pyinstaller` / `pkg` but *native*. This is *the* reason Go won DevOps.

## Session Log

When a Claude session does work in this section, append an entry to the root [`SESSIONS.md`](../SESSIONS.md) before ending — do **not** log session history in this file. `PLAN.md` is the plan; `SESSIONS.md` is the history. Tick the Status boxes above as items complete.
