# Mini-project — instrumented webhook-runner

The webhook-runner from [`04-http-servers/mini-project`](../../04-http-servers/mini-project/)
with logs, metrics, and traces wired in. Only the cross-cutting observability
surface is new; the HMAC + command-execution shape carries over.

## What's different

| Aspect | 04-http-servers | 10-observability |
|---|---|---|
| Config source | YAML file | Hard-coded map (the YAML lesson belongs to 04) |
| Logging | `log.Println` | `*slog.Logger` with request_id + trace_id on every line |
| `/metrics` | Not exposed | `promhttp.HandlerFor(...)` on a custom registry |
| Traces | None | OTel spans: `POST /webhook` → `verify-hmac` → `run-job` |
| Graceful shutdown | The lesson there | Not the lesson here — removed |

## Testable surface

```go
type Job struct { Command []string }

type Metrics struct {
    Requests *prometheus.CounterVec   // method, path, status
    Duration *prometheus.HistogramVec // method, path
    InFlight prometheus.Gauge
    Jobs     *prometheus.CounterVec   // job, result (ok|fail|unknown)
}
func NewMetrics() *Metrics

type ServerOpts struct {
    Secret    []byte
    Jobs      map[string]Job
    MaxOutput int
    Logger    *slog.Logger
    Metrics   *Metrics
    Tracer    trace.Tracer
}
func newServer(ServerOpts) http.Handler

func VerifyHMAC(ctx, tracer, secret, body, header) bool
func RunJob(ctx, tracer, j, maxOutput) (exitCode, output, error)
```

Tests pass a `*tracetest.InMemoryExporter`, a `bytes.Buffer`-backed logger,
and inspect the registry directly. **No real network, no real daemon.**

## The observability middleware (key piece)

```go
func observability(opts ServerOpts) func(http.Handler) http.Handler
```

For each request:

1. Generate `request_id` (atomic counter — request IDs should be opaque;
   in production use UUID or ULID).
2. Start the server span. The returned ctx carries the span.
3. Build a per-request logger pre-bound with `request_id` and (if the span
   is sampled) `trace_id`. Stash on ctx via the unexported `ctxKey{}`.
4. `InFlight.Inc()` / `defer InFlight.Dec()`.
5. Run the wrapped handler with the new ctx.
6. Record duration histogram + requests counter using a `statusRecorder`
   that captures the status code the handler wrote.
7. `defer span.End()`.

This shape is **the canonical "instrument an HTTP server" middleware** in
Go. Three concerns colocated, in one place, where you can read them.

## Why a custom registry (and not `prometheus.DefaultRegisterer`)

Tests construct multiple servers. With `promauto` + the default registry,
the second server's `MustRegister` panics ("duplicate metrics collector
registration"). Holding a registry on `*Metrics` is the cleanest fix —
production code does the same to support hot-reloads.

## What the tests verify

| Test | Concept |
|---|---|
| `TestVerifyHMAC_*` (3) | HMAC pure-function correctness (carryover from 04) |
| `TestRunJob_*` (3) | Process spawning + exit-code capture + truncation |
| `TestServer_HappyPath` | Valid request increments `http_requests_total{status=200}` AND `webhook_jobs_total{result=ok}` |
| `TestServer_BadSignatureReturns401AndCounts` | 401 path produces the right counter |
| `TestServer_UnknownJobReturns404AndCountsUnknown` | 404 path increments `result=unknown` |
| `TestServer_FailingJobCountsFail` | Non-zero exit code → `result=fail` (still HTTP 200) |
| `TestServer_MetricsEndpointExposesSeries` | `/metrics` returns OpenMetrics with our 4 names |
| `TestServer_RequestLogIncludesRequestID` | Slog lines have `request_id` AND `trace_id` |
| `TestServer_SpansCreated` | All 3 spans present: server, verify-hmac, run-job |
| `TestServer_InFlightReturnsToZero` | Deferred Dec() runs (no leak) |

## How to run (once you've implemented `VerifyHMAC` and `RunJob`)

```bash
WEBHOOK_SECRET=swordfish go run ./10-observability/mini-project &

# fire a request with a valid signature (use the helper from main_test.go's sign())
PAYLOAD='{"job":"echo"}'
SIG=$(printf '%s' "$PAYLOAD" | openssl dgst -sha256 -hmac swordfish | sed 's/^.* /sha256=/')
curl -s -X POST -H "X-Hub-Signature-256: $SIG" -d "$PAYLOAD" localhost:8080/webhook

# observe metrics
curl -s localhost:8080/metrics | grep -E '^(http_requests_total|webhook_jobs)'
```

## Notes

- **`request_id` is just an atomic counter** in this scaffold. Production
  should use UUID or read the upstream `X-Request-ID` header. The atomic
  counter keeps tests deterministic ("first request is id=1").
- **`trace_id` only appears in logs when the span is sampled.** With the
  in-memory exporter and `sdktrace.AlwaysSample()` (default for the test
  TracerProvider), every span is sampled and every log line has it.
- **The job counter uses `result=ok|fail|unknown`** as a bounded enum —
  exactly the kind of low-cardinality label that's safe to use.
- **`/metrics` is NOT wrapped by the observability middleware.** Scraping
  yourself is a useless data point and inflates the histogram.
