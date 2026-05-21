// 04-graceful-shutdown — drain in-flight requests on SIGTERM.
//
// Why this matters in prod:
//   - Kubernetes sends SIGTERM and waits up to terminationGracePeriodSeconds before SIGKILL.
//   - A naive `http.ListenAndServe` aborts active requests immediately on signal.
//   - srv.Shutdown(ctx) stops accepting new connections, then waits for in-flight handlers
//     to return (until ctx is cancelled).
//
// Run:
//   go run .
//   # in another shell: curl http://localhost:8080/slow & ; sleep 1 ; kill -TERM $(pgrep -f 04-graceful-shutdown)
//   # the in-flight /slow should still complete.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /slow", func(w http.ResponseWriter, r *http.Request) {
		// Simulate a long handler. Use r.Context() so a hard-shutdown would still cancel us.
		select {
		case <-time.After(5 * time.Second):
			fmt.Fprintln(w, "done")
		case <-r.Context().Done():
			http.Error(w, "cancelled", http.StatusServiceUnavailable)
		}
	})
	mux.HandleFunc("GET /fast", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// TODO: run srv.ListenAndServe() in a goroutine; send any non-ErrServerClosed
	// error onto an `errCh` so main can react.
	//
	// TODO: signal.Notify on SIGINT + SIGTERM. Block on either:
	//         - a signal arriving (start shutdown)
	//         - an error from errCh (server died — log + exit)
	//
	// TODO: on signal, create `shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)`
	//       call srv.Shutdown(shutdownCtx). If it returns an error, log it.
	//       Then exit cleanly.

	_ = srv
	_ = signal.Notify
	_ = syscall.SIGTERM
	_ = context.WithTimeout
	_ = os.Interrupt
	log.Println("started")
}
