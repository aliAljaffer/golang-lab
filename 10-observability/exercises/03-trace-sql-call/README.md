# Exercise 03 — trace a DB call

Wrap a fake DB call with an OTel span that:

- Is named `db.query`.
- Carries `db.statement` (the SQL) as an attribute.
- Carries `db.rows_affected` on success.
- On error: records the error via `span.RecordError(err)` AND sets the span
  status to `codes.Error` via `span.SetStatus(codes.Error, msg)`.

## What you implement

```go
type DB interface {
    Query(ctx context.Context, sql string) (rows int, err error)
}

func Query(ctx context.Context, tracer trace.Tracer, db DB, sql string) (int, error)
```

## Semantic conventions

Use the OTel canonical attribute names:

| Attribute | Meaning |
|---|---|
| `db.statement` | The SQL text |
| `db.system` | `postgresql`, `mysql`, etc. (not in this exercise) |
| `db.rows_affected` | Rows returned by the query |
| `db.operation` | `SELECT`, `INSERT`, ... (could be parsed from the SQL) |

Pre-built dashboards in Jaeger / Tempo / Honeycomb / Datadog recognize
these. If you invent your own names, you re-implement the dashboards.

## RecordError vs SetStatus — both, not either

`span.RecordError(err)` adds an event with `exception.message`, `exception.type`,
and stack-trace fields. Useful for finding the error in dashboards.

`span.SetStatus(codes.Error, msg)` paints the span red and bubbles up
through trace-level error rates. Useful for alerting.

**Always do both** on a span error. The error-event without the status
means dashboards see a "successful" span with a stray exception annotation.

## Run the failing suite

```bash
go test -tags=exercise ./10-observability/exercises/03-trace-sql-call/
```
