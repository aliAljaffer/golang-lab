# Exercises — 10-observability

Each subfolder is an exercise with failing tests. Run them with:

```bash
go test -tags=exercise ./10-observability/exercises/...
```

| # | Folder | What you build |
|---|---|---|
| 01 | `01-log-context-key/` | HTTP middleware that stashes request_id + a ctx-bound logger so downstream code logs with `request_id` automatically |
| 02 | `02-rate-limited-logging/` | Per-key rate limiter (`Allow("conn-refused")`) for noisy logs — 1 per window per key |
| 03 | `03-trace-sql-call/` | `Query(ctx, tracer, db, sql)` that instruments a DB call with `db.statement`, `db.rows_affected`, and proper error recording |

Each exercise defines its OWN narrow interface and uses an injectable clock /
ID generator so tests don't depend on wall time or random UUIDs.

See parent [`PLAN.md`](../PLAN.md).
