// Package dbspan instruments a fake "DB call" with an OTel span.
//
// Surface:
//
//	type DB interface {
//	    Query(ctx context.Context, sql string) (rows int, err error)
//	}
//	func Query(ctx context.Context, tracer trace.Tracer, db DB, sql string) (int, error)
//
// What Query must do:
//
//  1. Start a span named "db.query" against the provided tracer; defer End.
//  2. Set attribute `db.statement` to the SQL.
//  3. Call db.Query; on success, set `db.rows_affected`.
//  4. On error, record the error on the span (span.RecordError(err)) AND
//     set the span status to codes.Error (span.SetStatus(codes.Error, ...)).
//     Production tracing UIs (Jaeger, Tempo, Honeycomb) light up errored
//     spans red — this is what makes them visible.
//
// Why semantic attribute names matter:
//
//	`db.statement`, `db.system`, `db.rows_affected` are OpenTelemetry
//	semantic conventions. Pre-built dashboards in your backend recognize
//	them. If you make up your own names ("query.text", "rows.changed"),
//	you re-implement the dashboard. Use OTel's vocabulary.
package dbspan

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// DB is the narrow interface — exactly one method that this package needs.
// Production passes *sql.DB (wrapped), tests pass a fake.
type DB interface {
	Query(ctx context.Context, sql string) (rows int, err error)
}

// Query runs sql against db, instrumented with an OTel span.
func Query(ctx context.Context, tracer trace.Tracer, db DB, sql string) (int, error) {
	// TODO: start a "db.query" span, set db.statement, call db.Query, then
	//   either set db.rows_affected (success) or RecordError + SetStatus
	//   codes.Error (failure). The error-recording path is the load-bearing
	//   bit — without SetStatus(codes.Error), the span shows up green in
	//   Jaeger / Tempo / Honeycomb and you can't tell anything broke.

	_ = ctx
	_ = tracer
	_ = sql
	_ = attribute.Int
	_ = codes.Error
	_ = trace.Tracer(nil)
	return 0, nil
}
