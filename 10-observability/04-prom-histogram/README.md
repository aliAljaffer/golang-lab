# 04 — Prometheus histogram

A counter says *how many*. A histogram says *how long* (or *how big*).
This example records HTTP request latency.

## What a histogram actually is

Internally, it's:

- A counter for each bucket (`{le="0.001"}`, `{le="0.002"}`, ..., `{le="+Inf"}`)
- A `_sum` counter (total observed value)
- A `_count` counter (number of observations)

The `_bucket` series are cumulative — every observation that falls into
the `le="0.005"` bucket is also counted in `le="0.01"`, `le="0.025"`, etc.
That's what makes `histogram_quantile()` work.

## Buckets matter

```go
// DEFAULT — fine for typical web latencies in seconds
// .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10

// EXPONENTIAL — programmable doubling
prometheus.ExponentialBuckets(0.001, 2, 14)
// → 0.001, 0.002, 0.004, 0.008, 0.016, ..., ~8.2s

// LINEAR — even spacing
prometheus.LinearBuckets(0, 0.1, 11)
// → 0, 0.1, 0.2, ..., 1.0
```

Rules of thumb:

- Pick buckets that **bracket** your expected p50–p99. Outside that range,
  the histogram tells you nothing useful.
- 10–15 buckets is the sweet spot. Fewer = poor quantile accuracy; more
  = wasted memory.
- **Seconds**, not milliseconds. Prometheus convention.

## Observe = time.Since(start).Seconds()

The standard pattern, top-of-handler:

```go
func myHandler(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    defer httpDuration.WithLabelValues(r.Method, "/foo").Observe(time.Since(start).Seconds())
    // ... do work ...
}
```

The `defer` is what makes this clean — early returns and panics still
record the latency.

## Reading quantiles in Prometheus

```promql
# p95 latency over 5 minutes
histogram_quantile(0.95,
    sum by (le, path) (rate(http_request_seconds_bucket[5m]))
)
```

- The outer `histogram_quantile(0.95, ...)` interpolates the bucket boundaries.
- `sum by (le, path)` aggregates across replicas — this is the part Summary
  metrics can't do, which is why histograms are preferred.
- `rate(..._bucket[5m])` gives per-second observation rates within each
  bucket. Use `rate`, not `increase`, for histogram math.

## Histogram vs Summary

| | Histogram | Summary |
|---|---|---|
| Quantiles | Computed in Prometheus | Computed in client |
| Aggregable | Yes (sum by le) | No |
| Memory | Cheap (counters) | Expensive (sliding window) |
| Verdict | **Use this** | Legacy; avoid in new code |

## In-flight gauge

A frequent companion to a duration histogram:

```go
var inFlight = promauto.NewGauge(prometheus.GaugeOpts{...})

inFlight.Inc()
defer inFlight.Dec()
```

Gauges go up AND down. Use them for "current concurrency" / "queue depth".
The mini-project for this section adds one.

## TODO

1. Uncomment the TODOs. Run `go run .`.
2. Generate load: `for i in 1 2 3 4 5 6 7 8 9 10; do curl -s localhost:8080/work; done`.
3. `curl localhost:8080/metrics | grep http_request_seconds` — observe
   `_bucket`, `_sum`, `_count` series.
4. Switch the buckets to `LinearBuckets(0, 0.01, 20)` and re-run. Reason
   about which is better for the 0–200ms range we're actually observing.
