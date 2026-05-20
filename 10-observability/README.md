# 10 — Observability

> Status: ☐ not started — see [`PLAN.md`](./PLAN.md)

## What you'll learn

- Structured logging with stdlib `log/slog` (added in Go 1.21)
- Exporting Prometheus metrics with `prometheus/client_golang`
- Distributed tracing basics with OpenTelemetry

## Mental model from other languages

| Concept | Go | Python | TS / Node |
|---|---|---|---|
| Structured logging | `log/slog` (stdlib!) | `structlog` / `loguru` | `pino` / `winston` |
| Prometheus metrics | `prometheus/client_golang` | `prometheus_client` | `prom-client` |
| Tracing | `go.opentelemetry.io/otel` | `opentelemetry-python` | `@opentelemetry/sdk-node` |

## The DevOps angle

This is where your code becomes operable. Logs, metrics, and traces are how SREs answer "what's broken?". Wiring them up consistently across your services is half the job of being a platform engineer. The good news: Go's ecosystem here is mature and the patterns are stable across the industry.

See [`PLAN.md`](./PLAN.md).
