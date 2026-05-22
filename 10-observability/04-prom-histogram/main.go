// 04-prom-histogram — request-duration histogram with sensible buckets.
//
// What this example proves:
//   - `promauto.NewHistogramVec(...)` with explicit `Buckets` — the bucket
//     boundaries are NOT optional and the defaults are usually wrong for
//     HTTP request latency.
//   - `prometheus.ExponentialBuckets(start, factor, count)` is the right
//     primitive for "I want buckets at 1ms, 2.5ms, 5ms, ..." style coverage.
//   - The "observe a Duration" pattern: `time.Since(start).Seconds()`.
//     Prometheus seconds, not milliseconds — convention is base SI units.
//
// Why buckets matter:
//
//	A histogram with the wrong buckets is worse than no histogram. The default
//	buckets (.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10) are tuned for
//	web request latency in seconds — fine for that, useless for nanosecond-
//	level RPC. Pick buckets that bracket your expected p50–p99 range with
//	~10–15 boundaries.
//
// The histogram_quantile() trick (read in Prometheus, not here):
//
//	histogram_quantile(0.95, sum by (le) (rate(http_request_seconds_bucket[5m])))
//	     ^ desired quantile        ^ le is the bucket-boundary label
//
// Run:
//
//	go run .
//	# in another terminal:
//	for i in 1 2 3 4 5; do curl localhost:8080/work; done
//	curl localhost:8080/metrics | grep http_request_seconds
package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// httpDuration: a histogram of request seconds. Exponential buckets from 1ms,
// doubling 14 times — covers 1ms up to ~16s. Adjust to your real workload.
var httpDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_seconds",
		Help:    "HTTP request latency in seconds.",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 14),
	},
	[]string{"method", "path"},
)

func workHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		// TODO: httpDuration.WithLabelValues(r.Method, "/work").Observe(time.Since(start).Seconds())
		_ = start
	}()

	// Simulate variable latency.
	time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)
	fmt.Fprintln(w, "done")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/work", workHandler)
	// TODO: mux.Handle("/metrics", promhttp.Handler())

	addr := ":8080"
	log.Println("listening on", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}

	_ = promhttp.Handler
}
