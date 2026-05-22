# 10 — Observability

> Status: ☑ scaffolded — examples + mini-project + exercises ready; implementation + walkthrough pending. See [`PLAN.md`](./PLAN.md).

The three pillars: logs (`log/slog`), metrics (`prometheus/client_golang`), and traces (`go.opentelemetry.io/otel`). Wiring them consistently across services is half the job of being a platform engineer; the good news is that Go's ecosystem here is mature and the patterns are stable across the industry.

This section walks each pillar independently, then composes all three in the mini-project — which rebuilds `04-http-servers`'s `webhook-runner` with the full observability surface.

---

## What you'll learn

- Structured logging with the stdlib `log/slog` (Go 1.21+) — the package that killed `zap`/`logrus`/`zerolog` for new code
- `slog.Group`, `AddSource: true`, JSON vs Text handler, and the levels-live-on-the-handler model
- The `WithLogger`/`FromContext` pattern for carrying a request-scoped logger
- Prometheus metrics: `Counter`, `CounterVec`, `Histogram`, `HistogramVec`, `Gauge` — and the cardinality trap (never label by `user_id` or raw URL)
- The RED method (`http_requests_total{method,path,status}` + duration histogram)
- Why Histograms beat Summaries (aggregation across replicas)
- Distributed tracing with OpenTelemetry: exporter → TracerProvider → Tracer → spans
- W3C `traceparent` propagation (`propagation.NewCompositeTextMapPropagator(TraceContext{}, Baggage{})`) and the "global propagator default is no-op" gotcha
- `otelhttp.NewHandler` (server) + `otelhttp.NewTransport` (client) for transparent HTTP auto-instrumentation
- Stitching `trace_id` into log lines so a logs-platform link drops you into the right trace

---

## Mental model from other languages

| Concept             | Go                                | Python                            | TS / Node                         |
| ------------------- | --------------------------------- | --------------------------------- | --------------------------------- |
| Structured logging  | `log/slog` (stdlib)               | `structlog` / `loguru`            | `pino` / `winston`                |
| Prometheus metrics  | `prometheus/client_golang`        | `prometheus_client`               | `prom-client`                     |
| Histogram bucket    | `prometheus.ExponentialBuckets`   | `prometheus_client.Histogram`     | `Histogram` (manual buckets)      |
| Tracing             | `go.opentelemetry.io/otel`        | `opentelemetry-python`            | `@opentelemetry/sdk-node`         |
| HTTP auto-instrument | `otelhttp.NewHandler/Transport`  | `opentelemetry-instrumentation-requests` | `@opentelemetry/instrumentation-http` |
| Logger from context | `slog.FromContext` (your helper)  | `structlog.contextvars`           | `AsyncLocalStorage`               |

**Go's twist:** `log/slog` arrived in 2023 and immediately became the recommended choice for new code. The third-party logging libraries (`zap`, `logrus`, `zerolog`) all still work, but the stdlib option is now competitive on performance and unbeatable on "no extra dep." All three pillars now use stdlib *or* a single first-party-blessed library — no more "which logger does this team use?" debate.

---

## The DevOps angle

This is where your code becomes operable. Logs, metrics, and traces are how SREs answer "what's broken?" — and the answer is much faster when all three are stitched together by `request_id` and `trace_id`.

The non-obvious production details:

- **Cardinality is the headline cost** of Prometheus. Labeling by `user_id` or raw `URL` blows up the time-series count and OOMs the scraper. Label by *bounded* dimensions (method, route template, status).
- **`/metrics` lives outside any observability middleware.** Every scrape would otherwise inflate `http_requests_total` and skew the duration histogram. Outer-mux pattern: instrument `/...`, plain-serve `/metrics`.
- **Always set a custom registry on shared structs in tests.** `promauto` registers against the default registry; the second `NewCounterVec` with the same name panics. Tests want isolated registries; production benefits when reloads spin up a fresh one.
- **`stdouttrace` for examples; real OTLP for prod.** The exporter is the only thing that changes — TracerProvider construction is identical. Examples 05/06 use `stdouttrace.WithPrettyPrint()` so spans land in the terminal.
- **`trace_id` in logs requires the span to be sampled.** Production with `ParentBased(TraceIDRatioBased(0.01))` only stamps ~1% of log lines with `trace_id`. That's fine — the goal is "for any log you care about, you *can* jump to the trace," not "every log has a trace."

---

## Walkthrough

Read these in order. Each is a runnable example demonstrating one specific concept. Examples 03/04/06 bind to `:8080` and serve real HTTP (`curl localhost:8080/metrics` works directly — no Prometheus server needed). Example 05 writes span JSON straight to stdout, also no infra.

1. [`01-slog-basics/`](./01-slog-basics/) — `slog.NewTextHandler` / `slog.NewJSONHandler`, the "levels live on the handler" model, `slog.Group`, `AddSource: true`. Why `slog` killed `zap`/`logrus`/`zerolog` for new code.
2. [`02-slog-context/`](./02-slog-context/) — the `WithLogger`/`FromContext` pattern via an unexported `ctxKey struct{}`. `FromContext` falls back to `slog.Default()` so callers never nil-check. `logger.With(...)` is a builder (returns new logger), not mutation.
3. [`03-prom-counter/`](./03-prom-counter/) — `promauto.NewCounterVec` + `promhttp.Handler()` + the canonical RED counter `http_requests_total{method,path,status}`. The cardinality-trap section is the headline — never label by `user_id` / raw URL.
4. [`04-prom-histogram/`](./04-prom-histogram/) — `HistogramVec` + `ExponentialBuckets(0.001, 2, 14)` + the `defer Observe(time.Since(start).Seconds())` idiom + the seconds-not-millis Prometheus convention. README contrasts Histogram vs Summary — "use Histogram, period."
5. [`05-otel-tracing/`](./05-otel-tracing/) — the three things: **exporter** (`stdouttrace`), **TracerProvider** (`sdktrace.NewTracerProvider`), **Tracer** (`tp.Tracer("...")`). `tracer.Start(ctx, name)` returns BOTH a new ctx AND the span — and you MUST use the returned ctx for children. `tp.Shutdown(ctx)` is not optional — without it the batcher drops the final batch.
6. [`06-trace-http/`](./06-trace-http/) — `otelhttp.NewHandler` server-side + `otelhttp.NewTransport` client-side + `propagation.NewCompositeTextMapPropagator(TraceContext{}, Baggage{})`. The W3C `traceparent` header anatomy. The "global propagator default is no-op" gotcha that breaks 100% of first-time setups.

---

## Mini-project: [`webhook-runner-instrumented`](./mini-project/)

Rebuilds `04-http-servers/mini-project` (`webhook-runner`) with full observability. `Metrics` struct holds `Requests` (`CounterVec method/path/status`) + `Duration` (`HistogramVec`) + `InFlight` (`Gauge`) + `Jobs` (`CounterVec job/result` with `ok|fail|unknown` enum), plus its own `*prometheus.Registry` so tests spin up independent servers without collisions.

The `observability(opts)` middleware does the cross-cutting work — generates `request_id` (atomic counter, deterministic for tests), starts the server span, builds a `*slog.Logger` pre-bound with `request_id` AND `trace_id` (the latter pulled from `span.SpanContext().TraceID()`), stashes it on ctx via unexported `ctxKey{}`, increments `InFlight` with defer-Dec, captures status via a `statusRecorder` wrapper, observes duration on End.

12 tests pin the contracts: HMAC verification, subprocess exit codes (process-failed-is-HTTP-200, not 5xx), the 4 metric series names, `request_id` AND `trace_id` both appearing in JSON log lines, 3 expected spans via `tracetest.NewInMemoryExporter`, and `InFlight` returns to zero (the deferred-Dec invariant).

Spec and starter in [`mini-project/`](./mini-project/).

---

## Exercises

See [`exercises/`](./exercises/):

1. **[`01-log-context-key`](./exercises/01-log-context-key/)** — `WithRequestID`/`RequestIDFromContext` + `WithLogger`/`LoggerFromContext` + `Middleware(base, idGen)`. The `idGen` is injected the same way the clock is in `06-testing/02-fake-clock` — tests use a constant `"GENERATED"`; production uses `uuid.New().String`. Tests pin that an incoming `X-Request-ID` header is honored (not regenerated over) and echoed on the response.
2. **[`02-rate-limited-logging`](./exercises/02-rate-limited-logging/)** — `Limiter` with injectable `Now func() time.Time` + per-key `lastSeen` map. Tests pin first-call-allowed, second-within-window-blocked, after-window-allowed-again, per-key isolation. The README mentions wrapping as a `slog.Handler` as the natural follow-up.
3. **[`03-trace-sql-call`](./exercises/03-trace-sql-call/)** — `Query(ctx, tracer, db, sql)` wraps a DB call with a `db.query` span, sets `db.statement` attribute, `db.rows_affected` on success, AND `span.RecordError(err) + span.SetStatus(codes.Error, ...)` on failure. The both-RecordError-AND-SetStatus contract is pinned — RecordError adds the event; SetStatus paints the span red; production needs both.

---

## Further reading

- [`log/slog` docs](https://pkg.go.dev/log/slog) — stdlib structured logging
- [`prometheus/client_golang` docs](https://pkg.go.dev/github.com/prometheus/client_golang/prometheus) — metrics library; the `promhttp` subpackage handles `/metrics`
- [Prometheus naming best practices](https://prometheus.io/docs/practices/naming/) — including the seconds-not-millis convention
- [OpenTelemetry Go docs](https://opentelemetry.io/docs/languages/go/) — start with "Getting Started" and the Manual Instrumentation page
- [OTel semantic conventions](https://github.com/open-telemetry/semantic-conventions) — the `db.statement` / `http.route` / etc. attribute names; dashboards in Jaeger/Tempo/Honeycomb auto-recognize these
- [W3C Trace Context spec](https://www.w3.org/TR/trace-context/) — the `traceparent` / `tracestate` header format example 06 emits
