# 03 — Prometheus counter

The simplest metric: a counter that goes up. This example exposes
`http_requests_total{method,path,status}` and increments it per request.

## The two registration styles

```go
// Style 1 — manual: build, then register against a specific registry.
c := prometheus.NewCounterVec(opts, labels)
prometheus.MustRegister(c)

// Style 2 — promauto: build AND register against the default registry, one call.
c := promauto.NewCounterVec(opts, labels)
```

Use `promauto` unless you need a custom registry (multi-tenant servers,
test isolation). Most code does not.

## The `/metrics` endpoint

```go
mux.Handle("/metrics", promhttp.Handler())
```

That's it. `promhttp.Handler()` serves the default registry in OpenMetrics
text format. Prometheus will scrape this on whatever interval is configured
in its `scrape_configs`. The endpoint is plain HTTP — no auth, no token —
because Prometheus scraping is usually inside a trusted network. If yours
isn't, wrap it in BasicAuth middleware.

## What the output looks like

```text
# HELP http_requests_total Count of HTTP requests handled, by method/path/status.
# TYPE http_requests_total counter
http_requests_total{method="GET",path="/hello",status="200"} 3
```

The `# HELP` and `# TYPE` lines come from `CounterOpts.Help` and the type
of metric. Prometheus uses these to render the metric in its UI.

## Counter vs Gauge vs Histogram vs Summary

| Type | When | Examples |
|---|---|---|
| Counter | Monotonically increasing | requests handled, errors seen, bytes sent |
| Gauge | Up and down | in-flight requests, queue depth, memory usage |
| Histogram | Distributions, bucketed | request duration, payload size |
| Summary | Distributions, quantiles | (use Histogram instead in modern setups) |

**Use Histogram, not Summary.** Histograms are aggregable across instances;
Summary quantiles are per-instance only and don't compose. The metric
community's consensus: Summary is legacy. See 04 for the histogram pattern.

## The label-cardinality trap

```go
// BAD: user_id has unbounded unique values
prometheus.NewCounterVec(opts, []string{"method", "path", "user_id"})

// BAD: full URL with query string
prometheus.NewCounterVec(opts, []string{"method", "url"})  // url=/items?id=42, =43, ...

// GOOD: low-cardinality route template
prometheus.NewCounterVec(opts, []string{"method", "route", "status"})
// route="/items/{id}", not "/items/42"
```

Cardinality = product of unique values across labels. 10k series per
metric is a lot; 1M will OOM Prometheus. If you can't enumerate the value
space, it's not a label — log it instead.

## TODO

1. Uncomment the TODOs (`httpRequests.WithLabelValues` + the `/metrics`
   handler registration). Run `go run .`.
2. `curl` `/hello` a few times. `curl /metrics | grep http_requests_total`.
3. Add a 404 branch (return 404 if `name=forbidden`) and a `status` label
   value to match. Observe two series appear.
4. Try adding `r.URL.RawQuery` as a label — DON'T register the metric, just
   write the line that would. Then explain (to yourself, in a comment) why
   it's wrong.
