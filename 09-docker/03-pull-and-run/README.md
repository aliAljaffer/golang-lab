# 03 — pull and run

`docker run --rm alpine:3 echo hello`, decomposed into the six SDK calls that
underpin it. Every higher-level container orchestration tool you'll write
ends up wiring this sequence somewhere.

## The six-step sequence

| # | Call | Equivalent CLI |
|---|---|---|
| 1 | `cli.ImagePull(ctx, ref, ...)` | `docker pull alpine:3` |
| 2 | `cli.ContainerCreate(ctx, cfg, ...)` | `docker create alpine:3 echo ...` |
| 3 | `cli.ContainerStart(ctx, id, ...)` | `docker start <id>` |
| 4 | `cli.ContainerWait(ctx, id, cond)` | `docker wait <id>` |
| 5 | `cli.ContainerLogs(ctx, id, ...)` | `docker logs <id>` |
| 6 | `cli.ContainerRemove(ctx, id, ...)` | `docker rm <id>` |

`docker run` is just a CLI helper that does all six.

## The "drain the pull reader" footgun

```go
rc, err := cli.ImagePull(ctx, "alpine:3", image.PullOptions{})
// ❌ if you forget the next line, the pull never completes:
_, _ = io.Copy(io.Discard, rc)
rc.Close()
```

The daemon streams JSON progress events back over the response body. As far
as the daemon is concerned, the pull is "in progress" until you read EOF.
Skipping the drain looks like it works on a cached image (already-present
layers fire near-zero events) but breaks the moment the image is new.

If you DO want to show progress, the lines are JSON objects with `status`,
`progressDetail`, etc. The CLI parses them and renders the progress bars.

## `ContainerWait`'s two-channel return

```go
statusCh, errCh := cli.ContainerWait(ctx, id, container.WaitConditionNotRunning)
select {
case err := <-errCh:
    // network or daemon error
case status := <-statusCh:
    // status.StatusCode is the exit code
}
```

This is the SDK's standard pattern for "long-poll until a condition." Don't
just `<-statusCh` — if the daemon goes away mid-wait, the error arm is the
only signal you'd get.

Conditions:

- `WaitConditionNotRunning` — container exited (most common).
- `WaitConditionNextExit` — container exited *after* this call. Use when you
  want to wait for the next exit specifically (e.g. `--restart` containers).
- `WaitConditionRemoved` — container has been removed.

## Why `stdcopy.StdCopy`?

Docker multiplexes stdout + stderr over a single connection with an 8-byte
header per chunk identifying the stream. `io.Copy(os.Stdout, logs)` would
mangle stderr into stdout AND interleave the header bytes — see
[`04-logs-stream`](../04-logs-stream/) for the full story.

## TODO

1. Uncomment all six steps.
2. Run; confirm `hello from a goroutine` lands on stdout.
3. Change `Cmd` to `[]string{"sh", "-c", "echo OUT; echo ERR >&2"}` and
   confirm `StdCopy` routes them to your stdout and stderr correctly.
4. Add a `HostConfig.AutoRemove: true` in `ContainerCreate` — then you can
   skip step 6. (But step 4's Wait now needs `WaitConditionRemoved`, since
   `NotRunning` would race with the auto-remove.)
