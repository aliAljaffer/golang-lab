//go:build exercise

package dbspan

import (
	"context"
	"errors"
	"testing"

	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// fakeDB implements DB. By default returns rows=42, err=nil. Override err
// to simulate a failed query.
type fakeDB struct {
	rows int
	err  error
}

func (f *fakeDB) Query(_ context.Context, _ string) (int, error) {
	return f.rows, f.err
}

func newRecorder(t *testing.T) (*tracetest.InMemoryExporter, *sdktrace.TracerProvider) {
	t.Helper()
	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })
	return exp, tp
}

func TestQuery_SpanNamed(t *testing.T) {
	exp, tp := newRecorder(t)
	tracer := tp.Tracer("test")
	_, _ = Query(context.Background(), tracer, &fakeDB{rows: 1}, "SELECT 1")

	spans := exp.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("got %d spans, want 1", len(spans))
	}
	if spans[0].Name != "db.query" {
		t.Errorf("span name = %q, want %q", spans[0].Name, "db.query")
	}
}

func TestQuery_DBStatementAttributeSet(t *testing.T) {
	exp, tp := newRecorder(t)
	tracer := tp.Tracer("test")
	sql := "SELECT name FROM users WHERE id = ?"
	_, _ = Query(context.Background(), tracer, &fakeDB{rows: 1}, sql)

	spans := exp.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("got %d spans, want 1", len(spans))
	}
	var found string
	for _, kv := range spans[0].Attributes {
		if string(kv.Key) == "db.statement" {
			found = kv.Value.AsString()
			break
		}
	}
	if found != sql {
		t.Errorf("db.statement = %q, want %q", found, sql)
	}
}

func TestQuery_RowsAffectedSetOnSuccess(t *testing.T) {
	exp, tp := newRecorder(t)
	tracer := tp.Tracer("test")
	_, _ = Query(context.Background(), tracer, &fakeDB{rows: 42}, "SELECT *")

	spans := exp.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("got %d spans, want 1", len(spans))
	}
	var found int64
	var ok bool
	for _, kv := range spans[0].Attributes {
		if string(kv.Key) == "db.rows_affected" {
			found = kv.Value.AsInt64()
			ok = true
			break
		}
	}
	if !ok {
		t.Fatal("db.rows_affected attribute missing on successful query")
	}
	if found != 42 {
		t.Errorf("db.rows_affected = %d, want 42", found)
	}
}

func TestQuery_ErrorRecordedAndStatusSet(t *testing.T) {
	exp, tp := newRecorder(t)
	tracer := tp.Tracer("test")
	stub := errors.New("connection refused")
	_, err := Query(context.Background(), tracer, &fakeDB{err: stub}, "SELECT *")
	if !errors.Is(err, stub) {
		t.Errorf("Query returned err = %v, want %v", err, stub)
	}

	spans := exp.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("got %d spans, want 1", len(spans))
	}
	if spans[0].Status.Code != codes.Error {
		t.Errorf("span status = %v, want codes.Error", spans[0].Status.Code)
	}
	if len(spans[0].Events) == 0 {
		t.Error("span has no events; RecordError should have added one")
	}
}

func TestQuery_ReturnsUnderlyingResult(t *testing.T) {
	_, tp := newRecorder(t)
	tracer := tp.Tracer("test")
	got, err := Query(context.Background(), tracer, &fakeDB{rows: 7}, "SELECT *")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if got != 7 {
		t.Errorf("rows = %d, want 7", got)
	}
}
