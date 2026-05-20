# 00 — Setup & Toolchain

> Status: ☑ done

The foundation. Before writing real Go code, get clear on: how to install it, the CLI commands you'll use daily, and the files every Go project ships with.

If you skip this section and dive into code, you'll spend the next month vaguely confused about `go.sum`, `GOPATH`, and why `go run .` works in one folder and not another. Half an hour here saves a week of friction.

---

## What you'll learn

- Installing Go and what `go version` actually means
- The Go CLI: `run`, `build`, `install`, `fmt`, `vet`, `test`, `mod`, `get`, `doc`, `env`
- Anatomy of `go.mod` and `go.sum` (and why both are committed)
- The structure of every Go file
- Project layout conventions
- Editor setup

---

## Mental model from other languages

| Concept | Go | Python | TypeScript / Node | Java | Bash |
|---|---|---|---|---|---|
| Dependency manifest | `go.mod` | `pyproject.toml` | `package.json` | `pom.xml` | — |
| Lock file | `go.sum` | `poetry.lock` / `uv.lock` | `package-lock.json` | — | — |
| Init project | `go mod init <path>` | `poetry init` | `npm init` | `mvn archetype:generate` | — |
| Install deps | `go mod tidy` | `poetry install` | `npm install` | `mvn install` | — |
| Add a dep | `go get <pkg>@<v>` | `poetry add <pkg>` | `npm install <pkg>` | edit pom.xml | — |
| Run a script | `go run <pkg>` | `python script.py` | `npx tsx script.ts` | `java -jar` | `bash script.sh` |
| Build artifact | `go build` → single binary | `pyinstaller` | `pkg` / `esbuild --bundle` | `mvn package` → `.jar` | — |
| Install binary globally | `go install <pkg>@<v>` | `pipx install <pkg>` | `npm install -g` | — | — |
| Format code | `go fmt` (mandatory) | `black` (optional) | `prettier` (optional) | `google-java-format` | — |
| Static checks | `go vet` (built-in) | `mypy` / `ruff` | `tsc --noEmit` / `eslint` | `spotbugs` | `shellcheck` |
| Run tests | `go test ./...` | `pytest` | `npm test` | `mvn test` | — |

**The key cultural difference:** Go bakes opinions into the toolchain. Formatting isn't a discussion (`go fmt` is one way). Tests don't need a framework. The single static binary output is *the* reason `kubectl`, `terraform`, `docker`, `helm`, `gh`, and `prometheus` are all in Go — DevOps tooling needs to deploy as one file with no runtime.

---

## The DevOps angle

The pitch: **one binary, no runtime, fast startup, cross-compiles trivially.**

```bash
# Build a Linux binary on a Mac, in one command:
GOOS=linux GOARCH=amd64 go build -o myapp ./cmd/myapp

# Build a binary for ARM Linux (e.g., a Raspberry Pi):
GOOS=linux GOARCH=arm64 go build -o myapp ./cmd/myapp
```

No Docker required, no cross-compile toolchain to install, no `manylinux` wheels to fight. This is why ops engineers reach for Go when bash + jq stop scaling.

---

## Installing Go

### macOS
```bash
brew install go
```

### Linux
```bash
# Download the latest from https://go.dev/dl/
curl -LO https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### Multiple versions (recommended for serious dev)
- [`asdf`](https://asdf-vm.com/) — `asdf install golang 1.22.0`
- [`gvm`](https://github.com/moovweb/gvm)

### Verify
```bash
go version
# go version go1.22.0 darwin/arm64
```

---

## The Go CLI — what you'll type every day

```bash
go run <pkg>              # compile + execute in one step (like `python script.py`)
go build                  # produce a binary (single static file, no execution)
go install <pkg>@<v>      # build + install to $GOBIN (like `npm i -g`)
go fmt ./...              # auto-format (mandatory in Go culture)
go vet ./...              # static analysis for common bugs
go test ./...             # run all tests
go test -race ./...       # run with the race detector enabled
go mod init <module-path> # create a new module (run once per project)
go mod tidy               # sync go.mod with actual imports; update go.sum
go get <pkg>@<version>    # add or update a dependency
go doc <pkg>              # read package docs from CLI
go env                    # dump Go environment variables
gofmt -w .                # alternate to `go fmt`, writes files in place
```

### `./...` — what does it mean?
A pattern meaning "this directory and all subdirectories recursively". So `go test ./...` runs every test in your repo. Common.

---

## The files every Go project has

| File | Purpose | Analog |
|---|---|---|
| `go.mod` | Module path, Go version, direct dependencies | `package.json`, `pyproject.toml`, `pom.xml` |
| `go.sum` | Cryptographic checksums for every dep (direct + transitive). **Commit this.** | `package-lock.json`, `poetry.lock` |
| `main.go` | Entry point for executables. Must declare `package main` and define `func main()`. | Python's `if __name__ == "__main__":`; Java's `public static void main` |
| `*_test.go` | Test files. Discovered automatically — no config. | `test_*.py` for pytest |
| `vendor/` *(optional)* | Locally vendored deps. Most repos skip this. | `node_modules/` (but optional in Go) |

### Anatomy of `go.mod`

```
module github.com/alialjaffer/golang-learning   // the import path of this module

go 1.22                                          // minimum Go version

require (
    github.com/spf13/cobra v1.8.0                // direct dependency, semver pinned
    github.com/stretchr/testify v1.8.4
)

require (
    github.com/inconshreveable/mousetrap v1.1.0 // indirect; pulled in by cobra
)
```

### Anatomy of `go.sum`

Two lines per dep version: the module zip hash and the `go.mod` hash. Generated and maintained by `go mod tidy` — never edit by hand.

**Always commit `go.sum`.** Without it, builds aren't reproducible. (Same reason `poetry.lock` and `package-lock.json` must be committed.)

---

## Structure of every Go file

```go
package main                  // every file declares a package

import (                      // imports go in a single block
    "fmt"
    "os"
)

func main() {                 // entry point for `package main`
    fmt.Println("hello", os.Args)
}
```

Key rules:
- One `package` declaration per file. All files in a directory must declare the same package.
- A package called `main` is an executable; everything else is a library.
- Unused imports are **compile errors**. (Yes, really. `gopls` removes them on save.)
- Identifiers starting with an uppercase letter are exported (public); lowercase are package-private.

---

## Project layout

The community-maintained [`golang-standards/project-layout`](https://github.com/golang-standards/project-layout) describes patterns:

| Folder | When to use |
|---|---|
| `cmd/<binary-name>/` | When your repo produces multiple binaries; one `main.go` per subfolder |
| `internal/` | Code you don't want imported by other modules (enforced by the compiler) |
| `pkg/` | Code intended to be importable by others (controversial — many devs skip this) |

**For small projects (like this repo): don't overthink it.** Files at the root, or organized by topic. We use numbered folders for learning order, not for production-style layout.

---

## Editor setup

- **VS Code**: install the official Go extension. It uses `gopls` (the Go LSP) under the hood.
- **vim/neovim**: any LSP client + `gopls`
- **Auto-format on save** is universal Go practice — turn it on.

---

## Walkthrough

Read these in order. Each is a runnable example demonstrating one specific concept.

1. [`01-hello-world/`](./01-hello-world/) — the smallest Go program; introduces `package main` and `func main()`
2. [`02-go-run-vs-build/`](./02-go-run-vs-build/) — `go run` vs `go build`; inspect the resulting binary
3. [`03-modules-and-deps/`](./03-modules-and-deps/) — CLI walkthrough: init a module, add a dep, observe `go.mod` / `go.sum` changes
4. [`04-go-env-tour/`](./04-go-env-tour/) — tiny program that prints and explains Go environment variables

---

## Mini-project: [`gostat`](./mini-project/)

A CLI that prints info about the current Go installation and module: version, module path, dep count, GOOS/GOARCH. Reads `go.mod` directly.

Spec and starter in [`mini-project/`](./mini-project/). Tests verify the `--json` output schema.

---

## Exercises

See [`exercises/`](./exercises/) for the prompts:

1. **`01-tidy-experiment`** — observe how `go mod tidy` modifies `go.mod` and `go.sum`
2. **`02-static-binary`** — build a hello-world, verify it's statically linked
3. **`03-env-explorer`** — code exercise: implement a `GoEnv` lookup with tests

---

## Further reading

- [Effective Go](https://go.dev/doc/effective_go) — short, dense, official
- [`go help`](https://pkg.go.dev/cmd/go) — the Go command reference
- [Go Modules Reference](https://go.dev/ref/mod) — the deep dive on `go.mod`
- ["The Go Programming Language" by Donovan & Kernighan](https://www.gopl.io/) — the canonical book
