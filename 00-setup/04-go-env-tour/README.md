# 04 — Go Environment Tour

## Run

```bash
go run ./00-setup/04-go-env-tour
```

## What each variable means

These are the values you'll most often care about. Inspect all of them with `go env`.

| Variable | What it is |
|---|---|
| `GOOS` | Target OS for the current binary (`darwin`, `linux`, `windows`, ...) |
| `GOARCH` | Target CPU architecture (`amd64`, `arm64`, ...) |
| `GOROOT` | Where Go itself is installed (the standard library lives here) |
| `GOPATH` | Legacy workspace dir. With modules, mostly used for `$GOPATH/bin` (where `go install` puts binaries) and `$GOPATH/pkg/mod` (the download cache) |
| `GOBIN` | Where `go install` writes executables. Defaults to `$GOPATH/bin`. **Add this to your PATH.** |
| `GOMODCACHE` | Where downloaded module versions are cached (under `$GOPATH/pkg/mod` by default) |
| `GOCACHE` | Where the *build* cache lives (speeds up incremental builds) |
| `GOPROXY` | URL of the module proxy. Default `https://proxy.golang.org,direct`. In corporate envs, often pointed at a private Athens / artifactory proxy. |

## Cross-compilation in action

```bash
GOOS=linux GOARCH=amd64 go run ./00-setup/04-go-env-tour
```

Wait — `go run` cross-compiles? Yes; it still emits a temp binary first. But you can't *execute* a linux binary on macOS, so this will likely fail with `exec format error`. Use `go build` instead when cross-compiling:

```bash
GOOS=linux GOARCH=amd64 go build -o tour-linux ./00-setup/04-go-env-tour
file tour-linux
```

## Try this

1. Run `go env GOPATH GOBIN` — note the values. Is `GOBIN` on your `$PATH`?
2. Run `go env GOPROXY` — what's it set to? Why might a corporate environment override this?
3. Run `du -sh $(go env GOMODCACHE)` — how much disk is your module cache using?
