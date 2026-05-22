# Exercise 03 — restart on exit

Watch the daemon's event stream; restart any container that exits non-zero.
This is the "auto-restart supervisor" pattern, written from scratch — the
same job `docker run --restart=on-failure` does, but observable and
extensible (you can layer in backoff, opt-in labels, max-restarts, etc.).

## What you implement

```go
type DockerAPI interface {
    Events(ctx, opts) (<-chan events.Message, <-chan error)
    ContainerStart(ctx, id, opts) error
}

func ShouldRestart(msg events.Message) bool
func Run(ctx context.Context, api DockerAPI) error
```

The split exists for testability: `ShouldRestart` is pure (events → bool),
`Run` glues it to the SDK. Tests cover both surfaces.

## Why "die" alone isn't enough

`docker stop my-container` produces this event:

```
Type: "container"  Action: "die"  Attributes: {exitCode: "0", ...}
```

Restarting that container would be wrong — the user explicitly stopped it.
The `exitCode != "0"` check is the actual signal.

## Contract pinned by tests

`ShouldRestart`:
- Type=container, Action=die, exitCode != "0" → **true**
- exitCode == "0" → **false** (clean shutdown)
- Different Type (image / network / etc.) → **false**
- Different Action (start, stop, kill, ...) → **false**
- Missing exitCode attribute → **false** (be conservative)

`Run`:
- Crashing container → exactly one `ContainerStart` call with its ID.
- Clean exit → zero `ContainerStart` calls.
- `ContainerStart` failure → log and CONTINUE (don't abort the supervisor on a flaky daemon).
- Transport error from `errCh` → propagate.
- `ctx.Done()` → return nil (clean shutdown).

## Production hardening (NOT in scope)

- Exponential backoff on rapid crash loops.
- Max-restarts-per-window (avoid pegging the daemon on a wedged image).
- Opt-in label filter: only restart containers with `auto-restart=true`.
- Persist the restart history so a process restart doesn't reset counters.

The daemon's built-in `--restart=on-failure[:N]` covers all of this already.
This exercise replicates a slice for the pedagogical value, not to ship.

## Run the failing suite

```bash
go test -tags=exercise ./09-docker/exercises/03-restart-on-exit/
```
