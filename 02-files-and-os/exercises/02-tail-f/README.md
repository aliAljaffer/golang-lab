# Exercise 02 — `tail -f`

The classic "watch this log file" tool, in library form.

## What to build

In `tailf.go`, implement:

```go
func ReadAppend(f *os.File, lastSize int64) (newBytes []byte, newSize int64, err error)
```

The function should:

1. Seek to `lastSize`.
2. Read everything available from there to EOF.
3. Return those bytes plus the new total file size (so the caller can pass it back on the next call).

This is the **kernel** of `tail -f` — a real implementation would call this in a loop with `time.Sleep` between calls (polling) or with `fsnotify` watching for `IN_MODIFY` events.

## Why this shape

Splitting "what's new since I last looked" from the polling loop makes the logic testable without `time.Sleep`. The tests can `ReadAppend`, append to the file, `ReadAppend` again, and assert the deltas — instantly, no flaky waits.

## Behaviour

- If the file shrank (truncated/rotated), return an error of your choice; the test only checks `err != nil`.
- An empty read (file didn't grow) returns `(nil, lastSize, nil)`.
- Don't `os.Open` inside `ReadAppend` — the caller owns the `*os.File`.

## Run

```
go test -tags=exercise ./02-files-and-os/exercises/02-tail-f/...
```

## Stretch

- Wrap `ReadAppend` in a `Follow(ctx context.Context, path string, out io.Writer) error` that polls with a ticker until ctx cancels.
- Detect log rotation (file is recreated; inode changed) by stat-ing periodically.
- Replace polling with `fsnotify` (already in `go.sum` as a viper transitive).
