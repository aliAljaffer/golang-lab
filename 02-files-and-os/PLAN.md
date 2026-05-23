# Plan: 02-files-and-os

## Concepts to cover

- [ ] `os.Open`, `os.Create`, `os.ReadFile`, `os.WriteFile` — when to use which
- [ ] `bufio.Scanner` for line reading; `bufio.Writer` for buffered writes
- [ ] `io.Reader` / `io.Writer` interfaces — Go's most reusable abstraction
- [ ] `path/filepath` — cross-platform paths (don't use `path` for filesystem paths!)
- [ ] `filepath.WalkDir` (preferred over older `filepath.Walk`)
- [ ] `os/exec.Command` — running subprocesses, capturing output, piping
- [ ] `os/signal` — graceful shutdown with `SIGINT`/`SIGTERM`
- [ ] `archive/tar`, `compress/gzip` — creating and reading tarballs
- [ ] Atomic file writes (temp file + rename pattern)

## Examples to build

| Folder | Demonstrates |
|---|---|
| `01-read-write/` | Basic file I/O — ReadFile / WriteFile |
| `02-line-scanner/` | bufio.Scanner reading a log file line-by-line |
| `03-walk/` | filepath.WalkDir to find files matching a pattern |
| `04-exec/` | os/exec running `ls`, capturing stdout + stderr |
| `05-signals/` | Trap Ctrl-C, clean up, exit |
| `06-tar-gz/` | Create and extract a tarball |
| `07-atomic-write/` | Safe file write that survives crashes mid-write |

## Mini-project

**`logrotate`** — rotates a log file: rename current to `.1`, gzip the old `.1` to `.2.gz`, delete files older than N days. Configurable via flags.

Tests verify:
- Creates correctly named rotated files
- Gzip is valid (decompressible)
- Old files cleaned up per `--keep-days` flag

## Exercises

1. **`01-dirdiff`** — given two directory paths, print files that differ (by hash)
2. **`02-tail-f`** — implement `tail -f` (poll-based is fine for learning; fsnotify is a stretch)
3. **`03-pipe-cmd`** — run `cmd1 | cmd2` from Go using `os/exec`

## Status

- [x] Concepts in README walkthrough
- [x] Examples 01-07 scaffolded (TODO-style — fill in `main.go` of each)
- [x] Mini-project `logrotate` scaffolded with failing tests
- [x] Exercises scaffolded with failing tests

## Session Log

When a Claude session does work in this section, append an entry to the root [`SESSIONS.md`](../SESSIONS.md) before ending — do **not** log session history in this file. `PLAN.md` is the plan; `SESSIONS.md` is the history. Tick the Status boxes above as items complete.
