# 06 — events

`docker events` in Go. A long-lived subscription to the daemon's event bus —
container lifecycle, image pulls/pushes, volume mounts, network attachments,
plugin loads. Useful for any tool that wants to *react* rather than poll.

## The two-channel return

```go
msgCh, errCh := cli.Events(ctx, events.ListOptions{Filters: f})
```

Same pattern as `ContainerWait` (see 03-pull-and-run). Select on both:

- `msgCh` yields `events.Message` values.
- `errCh` yields a single error then closes (transport died, daemon went away).
- `ctx.Done()` is the third arm — your cancellation signal.

The error arm is non-optional. If you only select on `msgCh`, a daemon
restart looks like "events just... stopped" — your tool freezes.

## What's in an `events.Message`?

| Field | Notes |
|---|---|
| `Type` | `container`, `image`, `volume`, `network`, `daemon`, `plugin`, `service`, `node`. |
| `Action` | `create`, `start`, `die`, `kill`, `destroy`, `pull`, `push`, `tag`, ... — depends on Type. |
| `Actor.ID` | Container ID, image ref+digest, volume name, etc. |
| `Actor.Attributes` | Free-form `map[string]string`. For container events, includes `name`, `image`, often `exitCode` on `die`. |
| `Time`, `TimeNano` | Wall clock. |
| `Scope` | `local` vs `swarm`. |

The full per-type matrix is in the Docker API docs. The shortcut: run
`docker events` in one terminal and `docker run --rm alpine echo hi` in
another — you'll see the canonical sequence: `create → attach → start → die → destroy`.

## Filtering at the source

```go
filters.NewArgs(
    filters.Arg("type", "container"),
    filters.Arg("event", "die"),         // only "die" actions
    filters.Arg("label", "owner=me"),    // only containers with this label
)
```

Same DSL as `docker events --filter type=container --filter event=die`.
**Always filter at the source** for noisy daemons — every event hits the
wire, so filtering client-side wastes bandwidth and CPU.

## The `errCh` closure ambiguity

When you cancel the ctx, you typically see `errCh` deliver `context.Canceled`
(or just close, depending on SDK version). The `&& ctx.Err() == nil` guard
in the example main is the idiomatic way to distinguish "we cancelled" from
"the daemon hung up."

## TODO

1. Uncomment the loop and run.
2. In another terminal: `docker run --rm alpine echo hi`. Observe the
   create → start → die → destroy sequence.
3. Remove the `type=container` filter — observe image / network events too.
4. Add a filter for a specific image: `filters.Arg("image", "alpine:3")`.
5. Wire this into a goroutine and feed the events into a channel-driven state
   machine. (You're 90% of the way to `03-restart-on-exit`.)
