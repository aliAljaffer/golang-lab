# 05 — exec

`docker exec <container> <cmd>` in Go. The trick is the API is two phases:
**create** reserves the exec, **attach** actually starts it AND gives you the
streams. You then **inspect** afterwards for the exit code.

## The four calls

| # | Call | What it does |
|---|---|---|
| 1 | `ContainerExecCreate` | Reserves an exec ID. Nothing runs yet. |
| 2 | `ContainerExecAttach` | Starts the exec; returns a `HijackedResponse` with stdin (Conn), stdout+stderr (Reader, multiplexed). |
| 3 | `stdcopy.StdCopy` | Demux Reader into your own stdout/stderr writers — same format as `04-logs-stream`. |
| 4 | `ContainerExecInspect` | After the exec has exited, returns `ExitCode`. Call it AFTER step 3 returns. |

There's also `ContainerExecStart` (no attach) for fire-and-forget execs — e.g.
healthchecks where you only care about the exit code, not the output. Most
tools want Attach.

## Why no exit code in the attach response?

Same reason `docker exec` blocks until exit but you only know the code by
asking afterward — the streams close when the exec finishes, but the exit
code is a separate field on the daemon's exec record. You retrieve it by
ID with `ContainerExecInspect`.

If you call Inspect WHILE the exec is still running (e.g. timing race),
`Running` will be `true` and `ExitCode` is meaningless. Wait for `StdCopy`
to return first — that means EOF on the stream, which means the process is done.

## ExecOptions you'll actually use

| Field | What |
|---|---|
| `Cmd` | `[]string{"sh", "-c", "..."}` — note: `["echo", "hi"]` runs `echo` directly, no shell. |
| `AttachStdout` / `AttachStderr` | Set both. Default is no output captured. |
| `AttachStdin` | Set if you want to pipe data in via `resp.Conn`. |
| `Tty` | Allocate a PTY. If true, output is NOT multiplexed (same gotcha as logs); use `io.Copy`, not `StdCopy`. |
| `User` | Run as a different user (e.g. `"nobody"`). |
| `WorkingDir` | Override the container's working dir. |
| `Env` | Extra env vars (overrides container's). |

## TODO

1. Start a victim container: `docker run -d --name worker alpine sleep 600`.
2. Uncomment the TODOs and run: `go run . worker -- sh -c 'echo hi; echo oops >&2; exit 7'`.
3. Confirm the exit code propagates: `echo $?` should be 7.
4. Try without `--`: how does shell expansion bite you?
5. Add `AttachStdin: true` and pipe data: `echo hello | go run . worker -- cat`.
   You'll need to copy `os.Stdin` to `resp.Conn` in a goroutine; close `resp.CloseWrite()` when done.
