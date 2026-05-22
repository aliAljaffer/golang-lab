# 05 — OpenTelemetry tracing

A span is a single timed operation. A trace is a tree of spans tied
together by a shared trace ID. This example builds a 3-span trace using
the stdout exporter so you can see the JSON.

## The three things you need to set up tracing

| Thing | What it is |
|---|---|
| Exporter | Where spans go. `stdouttrace`, `otlptracegrpc`, `otlptracehttp`, `jaeger`, ... |
| TracerProvider | Holds the exporter + sampler + resource attributes. Built once. |
| Tracer | `otel.Tracer("name")` — what you call to start spans. Cheap to fetch. |

Wire them up at the boundary (`main` or a setup helper), then code anywhere
in the service grabs a tracer by name and starts spans.

## The `Start` -> use returned ctx -> `End` rule

```go
ctx, span := tracer.Start(ctx, "operation")
defer span.End()

// Use `ctx` (not the outer ctx) for anything that should be a child of this span.
```

Three failure modes if you forget:

1. **No `defer span.End()`** → span never flushes; trace looks truncated.
2. **`_, span := tracer.Start(ctx, ...)` and you keep using the outer ctx** →
   children become siblings of the parent's parent, not children of THIS span.
3. **`ctx, span := ...` but you pass the OLD ctx to subcalls** → same bug.

The single rule: **use the returned ctx everywhere downstream of the Start.**

## Sampling

- `sdktrace.AlwaysSample()` — every span sampled. Fine for dev, expensive in prod.
- `sdktrace.NeverSample()` — drop everything. Useful for tests / benchmarks.
- `sdktrace.TraceIDRatioBased(0.01)` — sample ~1% of root traces.
- `sdktrace.ParentBased(...)` — defer to the parent's decision. Use this so
  a request that's already being traced upstream keeps its full trace.

Production default: `ParentBased(TraceIDRatioBased(0.01))` or similar.

## Attributes

```go
span.SetAttributes(
    attribute.String("user.id", "u1"),
    attribute.Int("payload.len", 256),
)
```

Span attributes are like log fields: searchable, filterable, indexed by
the backend. Use **OpenTelemetry semantic conventions** for canonical
attribute names (`http.method`, `db.system`, `net.peer.name`) so your
spans light up in pre-built dashboards.

## Shutdown matters

```go
defer func() {
    _ = tp.Shutdown(ctx)
}()
```

Without `Shutdown`, the batch exporter's last batch may never flush — you
see "the trace is missing my last few spans" in dashboards. Always
shutdown the TracerProvider on clean exit (or use `tp.ForceFlush(ctx)`
before non-clean exits).

## Span events vs span attributes vs logs

- **Attributes**: about the *whole span* (`http.status_code=500`).
- **Events**: timestamped marker inside a span (`span.AddEvent("cache_miss")`).
  Like a sub-span without the cost of a full span.
- **Logs**: stay in your logger. Cross-reference by `trace_id` (06 covers this).

## Compare to other ecosystems

|                 | Go (`otel/sdk`)      | Python (`opentelemetry-python`) | TS (`@opentelemetry/sdk-node`) |
|-----------------|----------------------|---------------------------------|-------------------------------|
| Init           | `sdktrace.NewTracerProvider` | `TracerProvider(...)`           | `NodeTracerProvider`           |
| Start span     | `tracer.Start(ctx, ...)` | `tracer.start_as_current_span` | `tracer.startSpan(...)`        |
| Propagate     | via `context.Context` | implicit (contextvars)          | implicit (AsyncLocalStorage)    |

Go's "pass ctx everywhere" feels heavy compared to Python's contextvar
magic, but it's *explicit* — you can always tell where the ctx-cut points
are by looking at function signatures.

## TODO

1. Uncomment the TODO blocks (in `initTracer`, `validate`, `fetchData`,
   `handleRequest`). Run `go run .`.
2. Observe three span JSON blobs. Note the shared `TraceID`. Note that the
   two child spans have `ParentSpanID` matching the root's `SpanID`.
3. Change `tracer.Start(ctx, "fetchData")` to use the outer ctx instead of
   the one returned by `handleRequest`'s Start. Re-run — see the parent
   linkage break.
4. Replace `stdouttrace` with `otlptracehttp.New(...)` pointing at
   `http://localhost:4318` and run a local Jaeger / Tempo collector to see
   the trace in a UI. (Optional follow-up; not required.)
