# 04 — logs stream

`docker logs -f <container>` in Go. The non-obvious part isn't the API call —
it's the multiplexed stream format you need `stdcopy.StdCopy` to demux.

## The multiplex frame format

Without a TTY, the daemon streams logs as a series of frames:

```
+-+-+-+-+-+-+-+-+
|0|0|0|0|0|0|0|0|     ← 8-byte header
+-+-+-+-+-+-+-+-+
 │       │
 │       └── 4 bytes big-endian: payload length N
 └── 1 byte: stream code (1=stdout, 2=stderr)
+-+-+-+-+...
| N bytes payload |
+-+-+-+-+...
```

`stdcopy.StdCopy(out, errOut, r)` reads frames in a loop, writes the payload
to `out` (when code is 1) or `errOut` (when code is 2), and discards the
header. Use it; don't write your own demuxer.

If you `io.Copy(os.Stdout, logs)` instead, your output looks like:

```
?  ?  line 0
?  ?  err 0
```

— those `?`s are the literal header bytes leaking through.

## The TTY exception

When the container was created with `Tty: true` (the `-t` flag), the daemon
*doesn't* multiplex — the stream is just raw bytes from the pseudo-TTY, and
`StdCopy` would misread them. In that case:

```go
io.Copy(os.Stdout, logs)
```

How do you know? Inspect first:

```go
info, _ := cli.ContainerInspect(ctx, id)
if info.Config.Tty { /* raw copy */ } else { /* stdcopy */ }
```

## LogsOptions you'll actually use

| Field | What |
|---|---|
| `ShowStdout` / `ShowStderr` | Both default false. Set both, or you get nothing. |
| `Follow` | Like `-f` — keep reading new bytes after the existing log is exhausted. |
| `Tail` | `"all"` or `"<n>"` — start from the last N lines instead of the beginning. |
| `Timestamps` | Prepend `2026-05-22T18:30:00Z ` to each line. |
| `Since` / `Until` | RFC3339 or Unix timestamps. `docker logs --since 5m` equivalent. |

## Stopping the follow

Two clean ways:

1. Cancel the `ctx` you passed to `ContainerLogs`. Underlying HTTP read
   returns; `StdCopy` returns.
2. Call `rc.Close()` on the reader.

Either is fine; the example uses `signal.NotifyContext` so SIGINT cancels.
Important: `StdCopy` returns an error in *both* cases, so the success check
is `if err != nil && ctx.Err() == nil` — i.e. only flag an error if the ctx
wasn't the cause.

## TODO

1. Start a chatty container: `docker run -d --name talker alpine sh -c 'i=0; while :; do echo line $i; echo err $i >&2; i=$((i+1)); sleep 1; done'`.
2. Uncomment the TODO block; `go run . talker`.
3. Confirm stdout and stderr land in the right places: `go run . talker 2>/dev/null` should only show "line N".
4. Add `Tail: "20"` and confirm you see backlog before the live tail starts.
5. Add `Timestamps: true`.
