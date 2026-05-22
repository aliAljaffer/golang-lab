# 02 — Files & OS

> Status: ☑ scaffolded — examples + mini-project + exercises have failing tests. See [`PLAN.md`](./PLAN.md).

Ops scripts are 90% "do something to files, run other commands, react to signals." This section is the toolkit for replacing bash one-liners with robust, testable Go: file I/O, directory walks, subprocesses, signals, and tarballs.

The single most useful idea here is the `io.Reader` / `io.Writer` interface pair — they're how files, network sockets, gzip streams, tar entries, and stdout/stderr all compose. Once you internalize it, every other API in Go feels cheap.

---

## What you'll learn

- Reading and writing files with `os`, `io`, `bufio` — and knowing when to use which
- The `io.Reader` / `io.Writer` interfaces — Go's most reusable abstraction
- Walking directories with `path/filepath.WalkDir` (cross-platform; the newer, generally-preferred form over `filepath.Walk`)
- Running subprocesses with `os/exec` — capturing stdout/stderr, piping, exit codes
- Handling signals (`SIGINT`/`SIGTERM`) for graceful shutdown
- Tarball creation/extraction with `archive/tar` + `compress/gzip`
- The atomic-write pattern: temp file + `os.Rename` (the only crash-safe way to update a config file)

---

## Mental model from other languages

| Concept              | Go                            | Python                | Bash                   |
| -------------------- | ----------------------------- | --------------------- | ---------------------- |
| Walk directory       | `filepath.WalkDir`            | `os.walk`             | `find`                 |
| Read line-by-line    | `bufio.Scanner`               | `for line in file:`   | `while read line`      |
| Subprocess           | `os/exec.Command`             | `subprocess.run`      | backticks / `$()`      |
| Signal handling      | `signal.Notify` + channel     | `signal.signal`       | `trap`                 |
| Atomic file write    | `os.CreateTemp` + `os.Rename` | same pattern          | `mv` (atomic on same fs) |
| Path joining         | `filepath.Join`               | `os.path.join` / `pathlib.Path` | `"$dir/$file"`  |
| Compressed archive   | `archive/tar` + `compress/gzip` | `tarfile`           | `tar czf`              |

**The cultural difference:** there is no `pathlib`-style fluent API. Paths are strings; manipulate them with `filepath.Join` / `filepath.Dir` / `filepath.Ext`. **Don't use the `path` package for filesystem paths** — that one is for forward-slash-only URL/import paths. Use `path/filepath` for anything that touches the disk; it handles Windows `\` correctly.

---

## The DevOps angle

A surprising amount of production ops code is shaped like: read a file → transform → write somewhere else, atomically, while reacting to SIGTERM. Log shippers, config reloaders, backup scripts, log rotators, packaging tools — they all reduce to this section. Doing it in Go means you can ship one static binary that runs identically on the dev's Mac and the production Linux container; no `bash` / `sed` / `awk` version drift to debug.

**Atomic writes (`07-atomic-write`) is the load-bearing example.** If your config-reloader writes by truncating and rewriting in place, a crash mid-write leaves a corrupted file the service can't read on restart. The temp-file-plus-rename pattern is what tools like `etcd`, `consul`, `terraform`, and `helm` all use under the hood.

---

## Walkthrough

Read these in order. Each is a runnable example demonstrating one specific concept.

1. [`01-read-write/`](./01-read-write/) — `os.ReadFile` / `os.WriteFile` for whole-file I/O, `os.Open` / `os.Create` when you need a stream. Why the whole-file functions exist (most config files fit comfortably in RAM).
2. [`02-line-scanner/`](./02-line-scanner/) — `bufio.Scanner` reading a log file line by line. The most common pattern for log processing; mind the default 64 KB line cap (`scanner.Buffer(...)` to raise it).
3. [`03-walk/`](./03-walk/) — `filepath.WalkDir` to find files matching a pattern. The visitor function returns `filepath.SkipDir` to prune the descent; `nil` to keep going; an error to abort.
4. [`04-exec/`](./04-exec/) — `os/exec.Command` running `ls`, capturing stdout + stderr separately. The pitfalls: `Run()` returns `*ExitError` you cast for exit codes; subprocesses inherit your environment unless you set `cmd.Env`.
5. [`05-signals/`](./05-signals/) — trap `SIGINT`/`SIGTERM` via `signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)`, clean up, exit. Critical for any long-running tool: containers get SIGTERM 30 seconds before SIGKILL, and that window is your only chance to flush state.
6. [`06-tar-gz/`](./06-tar-gz/) — create and extract `.tar.gz`. Shows the `io.Writer` composition that makes Go's archive packages compose: `gzip.NewWriter(f)` returns an `io.Writer` you pass to `tar.NewWriter(...)`.
7. [`07-atomic-write/`](./07-atomic-write/) — safe file write that survives a crash mid-write. The recipe: `os.CreateTemp(dir, "prefix-*")` → write → `f.Sync()` → `f.Close()` → `os.Rename(tmp, final)`. The rename is atomic on POSIX filesystems as long as the temp file lives on the same filesystem as the destination.

---

## Mini-project: [`logrotate`](./mini-project/)

Rotate a log file: rename the current file to `.1`, gzip the previous `.1` to `.2.gz`, delete files older than N days. Configurable via flags.

The point: real log-management code combines almost everything in this section — file I/O, atomic operations, gzip, filesystem walks, and time-based deletion logic. Tests verify the rotation produces correctly named files, the gzip is valid (decompressible), and old files are cleaned up per `--keep-days`.

Spec and starter in [`mini-project/`](./mini-project/).

---

## Exercises

See [`exercises/`](./exercises/):

1. **[`01-dirdiff`](./exercises/01-dirdiff/)** — given two directory paths, print files that differ by content hash. Practices `WalkDir` + `crypto/sha256` streaming + set diff.
2. **[`02-tail-f`](./exercises/02-tail-f/)** — implement `tail -f`. Poll-based is fine for the exercise; `fsnotify` is the production-grade stretch.
3. **[`03-pipe-cmd`](./exercises/03-pipe-cmd/)** — run `cmd1 | cmd2` from Go using `os/exec`. The trick: wire `cmd2.Stdin = cmd1.StdoutPipe()`, `Start()` both, `Wait()` both in order.

---

## Further reading

- [`os` package docs](https://pkg.go.dev/os) — the system-call surface
- [`io` package docs](https://pkg.go.dev/io) — the `Reader`/`Writer` interface universe
- [`path/filepath` docs](https://pkg.go.dev/path/filepath) — cross-platform path manipulation
- [`os/exec` docs](https://pkg.go.dev/os/exec) — subprocesses, with the well-known footguns documented
- [Effective Go: errors](https://go.dev/doc/effective_go#errors) — relevant because `os` errors lean heavily on `errors.Is(err, os.ErrNotExist)` etc.
