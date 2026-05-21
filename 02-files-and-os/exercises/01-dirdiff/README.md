# Exercise 01 — `dirdiff`

Compare two directory trees and report which files differ.

## What to build

In `dirdiff.go`, implement:

```go
func Diff(left, right string) ([]Entry, error)
```

Where `Entry.Kind` is one of:

| Kind        | Meaning                                           |
| ----------- | ------------------------------------------------- |
| `OnlyLeft`  | File exists in `left` but not in `right`          |
| `OnlyRight` | File exists in `right` but not in `left`          |
| `Modified`  | File exists in both, but contents (sha256) differ |

Files that exist in both with identical sha256 are **not** in the result.

## Behaviour

- Walk both trees with `filepath.WalkDir`.
- Compare files by their relative path (relative to `left` / `right`).
- Compute sha256 via `crypto/sha256` — `sha256.New()` is an `io.Writer`, so `io.Copy(h, f)` is the idiom.
- Ignore directories themselves; only emit entries for regular files.
- Order doesn't matter (the tests sort before comparing).

## Run

```bash
go test -tags=exercise ./02-files-and-os/exercises/01-dirdiff/...
```

## Stretch

- Add an `IgnorePatterns []string` option (matched with `filepath.Match`).
- Parallelize the hashing (worker pool — useful prep for section 05).
- Emit a unified-diff for `Modified` text files.
