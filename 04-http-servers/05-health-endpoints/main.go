// 05-health-endpoints — /healthz vs /readyz.
//
// Liveness (/healthz): "is the process alive?" Used by k8s to decide whether to restart.
//   Should be cheap and never depend on external systems — otherwise a DB blip
//   triggers a restart loop.
//
// Readiness (/readyz): "should I receive traffic right now?" Used by k8s to add/remove
//   the pod from the Service endpoints. Can check external dependencies (DB, queue).
//
// During graceful shutdown, flip readiness to false BEFORE calling srv.Shutdown.
// k8s will drain traffic ~one probe period later, then your handlers will be idle
// when you actually shut down.
//
// Run:
//   go run .
//   curl -i http://localhost:8080/healthz   # always 200
//   curl -i http://localhost:8080/readyz    # 200 until you "fail" the DB check
package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

// readiness is the live "should I take traffic?" flag. Toggle it from your shutdown handler.
// atomic.Bool is the right type — readers run on every request.
var readiness atomic.Bool

func livez(w http.ResponseWriter, r *http.Request) {
	// TODO: write 200 + "ok\n". Never check external systems here.
	_, _ = fmt.Fprintln(w, "TODO")
}

func readyz(w http.ResponseWriter, r *http.Request) {
	// TODO: if !readiness.Load() { http.Error(w, "not ready", 503); return }
	// TODO: optionally probe a dependency here (DB ping, etc.) — anything that returns an
	//       error means 503.
	// TODO: write "ready\n".
	_, _ = fmt.Fprintln(w, "TODO")
}

func main() {
	readiness.Store(true)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", livez)
	mux.HandleFunc("GET /readyz", readyz)

	// A toy "trip the readiness" endpoint to demo the toggle.
	mux.HandleFunc("POST /admin/drain", func(w http.ResponseWriter, r *http.Request) {
		readiness.Store(false)
		fmt.Fprintln(w, "drained: now reporting not-ready")
	})

	log.Fatal(http.ListenAndServe(":8080", mux))
}
