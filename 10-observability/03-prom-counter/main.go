// 03-prom-counter — a minimal HTTP server that exposes a counter at /metrics.
//
// What this example proves:
//   - `promauto.NewCounterVec(...)` registers a counter with the default
//     registry. `promauto` is the "register-on-construct" wrapper — without
//     it, you'd `prometheus.NewCounter(...)` then `prometheus.MustRegister(c)`.
//   - `promhttp.Handler()` is the canonical `/metrics` handler. It serves
//     the OpenMetrics text format that Prometheus scrapes.
//   - `counter.WithLabelValues("GET", "/healthz").Inc()` looks up (or creates)
//     the time-series for that label combination and increments it.
//
// The label-cardinality footgun:
//
//	Every unique combination of label values creates a new time series in
//	Prometheus. Labels like "method" (5 values) + "status" (~10 values) +
//	"endpoint" (~50) yield 2,500 series — fine. But adding "user_id" or
//	"request_id" as a label = unbounded cardinality = your Prometheus
//	memory blows up. Rule of thumb: labels should have <100 unique values
//	each, and the product across all labels should stay under ~10k.
//
// Run:
//
//	go run .
//	# in another terminal:
//	curl localhost:8080/hello
//	curl localhost:8080/hello?name=ada
//	curl localhost:8080/metrics | grep http_requests_total
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// httpRequests is the canonical "RED" counter: requests, by method+path+status.
// 3 labels — keep them low-cardinality (no user_id, no request_id, no raw URL).
var httpRequests = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Count of HTTP requests handled, by method/path/status.",
	},
	[]string{"method", "path", "status"},
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "world"
	}
	fmt.Fprintf(w, "hello %s\n", name)
	// TODO: httpRequests.WithLabelValues(r.Method, "/hello", "200").Inc()
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", helloHandler)
	// TODO: mux.Handle("/metrics", promhttp.Handler())

	addr := ":8080"
	log.Println("listening on", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}

	_ = promhttp.Handler
}
