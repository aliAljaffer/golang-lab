# Exercise 03 — `pipe-cmd`

The bash `cmd1 | cmd2 | cmd3` of Go's `os/exec`. The point is to understand
how `StdoutPipe` connects to the next `Stdin`, and that **all** commands run
concurrently — the OS handles the buffering.

## What to build

In `pipe.go`, implement:

```go
func Pipe(input io.Reader, cmds ...[]string) (stdout []byte, err error)
```

Where each `cmds[i]` is `[binary, arg1, arg2, ...]`. The function should run
all commands concurrently and pipe them:

```bash
input -> cmds[0] -> cmds[1] -> ... -> stdout
```

If any command exits non-zero, return a non-nil error.

## Behaviour

- For each command `c[i]`: `cmd[i].Stdin = cmd[i-1].StdoutPipe()` (special-case `i==0` to use the `input` reader).
- The final command's stdout goes into a `bytes.Buffer` you return.
- Call `Start()` on each in order, then `Wait()` on each (also in order).
- If you `Run()` the first one before starting the second, the OS pipe buffer (typically 64 KB) will block the first when it fills. This is the bug the exercise teaches.

## Run

```bash
go test -tags=exercise ./02-files-and-os/exercises/03-pipe-cmd/...
```

Tests use `cat`, `tr`, `sort`, `wc` — all stdlib Unix tools. Tests are gated
to non-Windows builds; on macOS / Linux they should pass once implemented.

## Stretch

- Capture each command's stderr separately and return them.
- Support cancellation via `context.Context` (`exec.CommandContext`).
- Add a timeout per command, killing stuck pipelines.
