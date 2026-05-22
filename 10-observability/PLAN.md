# Plan: 10-observability

## Concepts to cover

### Logging
- [ ] `log/slog` basics: handlers (text vs JSON), levels, attributes
- [ ] Default logger vs context-bound loggers
- [ ] Structured fields: `slog.String`, `slog.Int`, `slog.Group`
- [ ] Source location: `slog.HandlerOptions{AddSource: true}`
- [ ] Custom handlers
- [ ] Why `log/slog` killed third-party loggers (zap, logrus) for new code

### Metrics
- [ ] Counter, Gauge, Histogram, Summary — when to use each
- [ ] Labels — and why high-cardinality labels are dangerous
- [ ] `promhttp.Handler()` — exposing `/metrics`
- [ ] Custom collectors

### Tracing
- [ ] Span basics, parent/child relationships
- [ ] Context propagation across goroutines and HTTP boundaries
- [ ] OTLP exporter to a local Jaeger / Tempo / collector
- [ ] Sampling

## Examples to build

| Folder | Demonstrates |
|---|---|
| `01-slog-basics/` | text + JSON handlers, levels, attributes |
| `02-slog-context/` | logger-in-context pattern |
| `03-prom-counter/` | minimal HTTP handler that increments a counter |
| `04-prom-histogram/` | request duration histogram with proper buckets |
| `05-otel-tracing/` | trace a fake "service" with parent + child spans, export to stdout |
| `06-trace-http/` | propagate context via HTTP headers |

## Mini-project

Instrument the HTTP server from `04-http-servers/mini-project` (`webhook-runner`) with:
- Structured logging (slog) at request boundaries
- Prometheus metrics: request count, request duration histogram, in-flight gauge
- OpenTelemetry tracing with spans for: HTTP handler → HMAC verification → command execution

Tests verify metrics are exposed and contain expected series.

## Exercises

1. **`01-log-context-key`** — implement a middleware that extracts a `request_id` from context and adds it to every log line within that request
2. **`02-rate-limited-logging`** — given noisy code, add log-rate-limiting (1 per second per error key)
3. **`03-trace-sql-call`** — instrument a fake DB call with a span; capture latency

## Status

- [ ] Concepts in README walkthrough
- [x] Examples 01-06 built
- [x] Mini-project: instrument webhook-runner
- [x] Exercises scaffolded

## Session Log

When a Claude session does work in this section, append an entry to the root [`SESSIONS.md`](../SESSIONS.md) before ending — do **not** log session history in this file. `PLAN.md` is the plan; `SESSIONS.md` is the history. Tick the Status boxes above as items complete.
