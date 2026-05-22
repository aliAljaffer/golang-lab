# Exercises — 09-docker

Each subfolder is an exercise with failing tests. Run them with:

```bash
go test -tags=exercise ./09-docker/exercises/...
```

| # | Folder | What you build |
|---|---|---|
| 01 | `01-container-stats/` | `CPUPercent(curr, prev)` and `MemoryPercent(s)` — the math behind `docker stats`. |
| 02 | `02-buildkit-tar/` | `BuildContext([]File) ([]byte, error)` — assemble an in-memory tar ready for `cli.ImageBuild`. |
| 03 | `03-restart-on-exit/` | `ShouldRestart(events.Message)` + `Run(ctx, api)` — supervise containers via the event stream. |

Each exercise defines its OWN narrow interface (`DockerAPI` with the 1–3
methods it needs). That's the production pattern — pass `*client.Client` in
production, a fake in tests.

See parent [`PLAN.md`](../PLAN.md).
