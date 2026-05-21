# 02 — Files & OS

> Status: ☑ scaffolded — examples + mini-project + exercises have failing tests. See [`PLAN.md`](./PLAN.md).

## What you'll learn

- Reading and writing files with `os`, `io`, `bufio`
- Walking directories with `path/filepath`
- Running subprocesses with `os/exec`
- Handling signals (graceful shutdown) with `os/signal`
- Working with tar/gzip archives

## Mental model from other languages

| Concept | Go | Python | Bash |
|---|---|---|---|
| Walk directory | `filepath.WalkDir` | `os.walk` | `find` |
| Read line-by-line | `bufio.Scanner` | `for line in file:` | `while read line` |
| Subprocess | `os/exec.Command` | `subprocess.run` | backticks / `$()` |
| Signal handling | `signal.Notify` + channel | `signal.signal` | `trap` |
| Atomic file write | `os.CreateTemp` + `os.Rename` | same pattern | `mv` (atomic on same fs) |

## The DevOps angle

Ops scripts are 90% "do something to files, run other commands, react to signals". This section is the toolkit for replacing bash one-liners with robust, testable Go.

See [`PLAN.md`](./PLAN.md).
