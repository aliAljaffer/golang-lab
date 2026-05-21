# `dirsize` — mini-project

A CLI that walks a directory and prints sizes per immediate subdirectory. Like `du -sh *` but sortable and pipeable into `jq`.

## Spec

```
dirsize PATH                # human-readable, sorted descending by size
dirsize PATH --top 3        # only the 3 largest
dirsize PATH --json         # JSON: [{"path":"...","bytes":1234}, ...]
```

| Exit | When |
|---|---|
| `0` | Success |
| `1` | `PATH` is missing or not a directory |
| `2` | Usage error (handled by cobra) |

## How to implement

The scaffold splits the work into four pure-ish functions plus a cobra command:

| Function | Job |
|---|---|
| `scan(root)` | Walk every immediate child dir; recursively sum regular-file sizes. |
| `sortAndTrim(entries, top)` | Sort descending by `Bytes`, then truncate to `top` if > 0. |
| `renderText(entries)` | Build the human-readable output string (humanize bytes). |
| `renderJSON(entries)` | `json.Marshal` the slice. |
| `newRootCmd()` | Wire flags, validation, and call the four above. |

This shape is deliberately testable — each function gets unit tests in `main_test.go`. The cobra plumbing is the thinnest possible.

## Run the tests

```
go test ./01-cli-tools/mini-project/...
```

All tests fail until you implement the TODOs.

## Stretch ideas (after the tests pass)

- `--sort name` / `--sort size` toggle
- Show file count alongside bytes
- Concurrent walk with a worker pool (great prep for section 05)
