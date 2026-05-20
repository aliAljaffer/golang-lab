# 02 — `go run` vs `go build`

Two ways to execute a Go program. Understanding the difference is the gateway to understanding why Go owns DevOps tooling.

## `go run` — compile + execute, no artifact

```bash
go run ./00-setup/02-go-run-vs-build
```

The compiler emits a binary to a temp directory, runs it, deletes it. Convenient for development. Analog: `python script.py` or `npx tsx script.ts`.

## `go build` — produce a binary

```bash
cd 00-setup/02-go-run-vs-build
go build
ls -lh   # observe the binary that appeared (named after the folder)
./02-go-run-vs-build  # run it directly
```

Now you have a standalone binary. **No Go installation needed to run it.** Ship it to a fresh server and it just works.

## Inspect the binary

```bash
file ./02-go-run-vs-build
# 02-go-run-vs-build: Mach-O 64-bit executable arm64
# (on Linux you'd see "ELF 64-bit LSB executable, ..., statically linked")
```

**Statically linked** = no `.so` / `.dll` dependencies. The Go runtime is *inside* the binary. This is **the DevOps superpower**: a single file is the deploy artifact.

## Cross-compile to Linux from a Mac

```bash
GOOS=linux GOARCH=amd64 go build -o helloapp-linux
file helloapp-linux
# helloapp-linux: ELF 64-bit LSB executable, x86-64, ..., statically linked, ...
```

No Docker. No cross-compile toolchain to install. This is why `kubectl`, `terraform`, `gh`, and `docker` are Go binaries you can `curl` and run.

## Try this

1. Run `go build` and notice the binary appears in the *current* directory, not the source directory. (Compare to `go install`, which puts it in `$GOBIN`.)
2. Run `du -h <binary>` — note the size (usually 1-5 MB for tiny programs). Larger than C because the Go runtime is bundled, but trivially shippable.
3. Try `GOOS=windows GOARCH=amd64 go build -o app.exe` — observe that you just produced a Windows binary on macOS.

## Clean up

```bash
rm 02-go-run-vs-build helloapp-linux app.exe 2>/dev/null
```
