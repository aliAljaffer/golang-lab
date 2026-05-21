# 04 — `os/exec`

How Go runs subprocesses. Replaces shell pipelines from Bash and `subprocess` from Python.

## The two-and-a-half APIs

```go
// Easiest: just want the combined output.
out, err := exec.Command("ls", "-la").CombinedOutput()

// Common: separate stdout and stderr.
cmd := exec.Command("ls", "-la")
var stdout, stderr bytes.Buffer
cmd.Stdout = &stdout
cmd.Stderr = &stderr
err := cmd.Run()

// Streaming: attach pipes and read while the process runs.
cmd := exec.Command("tail", "-f", "log.txt")
stdout, _ := cmd.StdoutPipe()
cmd.Start()
// read from stdout...
cmd.Wait()
```

## Things to learn

- `exec.Command` does **not** invoke a shell — `exec.Command("ls *.go")` looks for a binary literally named `ls *.go`. To get shell semantics, pass `"sh", "-c", "ls *.go"` explicitly. (Usually you don't want shell semantics — it's a frequent source of injection bugs.)
- `cmd.Run()` returns an `*exec.ExitError` when the child exits non-zero. Type-assert to inspect the exit code: `if ee, ok := err.(*exec.ExitError); ok { ee.ExitCode() }`.
- `cmd.Env = []string{"KEY=VAL"}` replaces the env; to **add** to the parent env, copy `os.Environ()` first.
- For cancellation, use `exec.CommandContext(ctx, ...)` and cancel the ctx.

## Comparison

| Language | Run + capture |
|---|---|
| Go | `exec.Command("ls").Output()` |
| Python | `subprocess.run(["ls"], capture_output=True)` |
| Bash | `out=$(ls)` |
