# `logrotate` — mini-project

A log rotation tool: rename → gzip → prune. Configurable. Combines almost
every concept from this section: file I/O, walking, gzip, atomic rename,
flags.

## Spec

```
logrotate --file PATH                  # rotate PATH once
logrotate --file PATH --keep-days 7    # delete rotated files older than 7 days
```

Given `--file=/var/log/app.log`, one invocation does:

1. If `app.log.1` exists, gzip it to `app.log.2.gz` (and delete `app.log.1`).
2. Rename `app.log` → `app.log.1`.
3. Create a fresh empty `app.log` (so the writer's open fd isn't broken any more than rename already broke it).
4. Walk the log's directory; delete any `app.log.*.gz` whose mtime is older than `--keep-days` days.

Exit `0` on success, `1` on any I/O error, `2` on flag misuse (cobra handles).

## How to implement

The scaffold splits the work into testable functions:

| Function | Job |
|---|---|
| `rotateOnce(path)` | Steps 1–3 above. Pure file I/O, no flag parsing. |
| `gzipFile(src, dst)` | Read `src`, write gzip-compressed bytes to `dst`. |
| `pruneOld(dir, prefix, keepDays, now)` | Step 4. `now` is injected so tests can pin time. |
| `newRootCmd()` | cobra wiring. |

Pulling out `gzipFile` and `pruneOld` keeps tests fast — no flag parsing, no `time.Now()` calls inside the logic.

## Run the tests

```
go test -tags=exercise ./02-files-and-os/mini-project/...
```

All tests fail until you implement the TODOs.

## Stretch ideas

- Rotate by size instead of just on-demand (`--max-mb`).
- Keep N rotated files instead of N days (`--keep-files`).
- Make the gzip step concurrent with the next rename (channel + worker).
- Detect that the log file is still held open by another process (lsof-style) and warn — useful in real systems where the writer needs a `SIGHUP` to reopen.
