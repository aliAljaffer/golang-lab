// 03-middleware — the `func(http.Handler) http.Handler` pattern.
//
// Three classic middlewares:
//   - withLogging: prints "<method> <path> -> <status> <duration>"
//   - withRecovery: catches panics and returns 500 instead of crashing the process
//   - withRequestID: assigns or echoes an X-Request-ID header
//
// The chain is bottom-up: withLogging(withRecovery(withRequestID(mux))).
// Each middleware wraps the next handler.
//
// Run:
//   go run .
//   curl -i http://localhost:8080/hi
//   curl -i http://localhost:8080/boom        # demonstrates recovery
package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"time"
)

// statusRecorder wraps http.ResponseWriter so we can record the status code
// for the logging middleware. The stdlib does not expose it after WriteHeader.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// withLogging logs method, path, status code, and elapsed time for each request.
func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: started := time.Now()
		// TODO: sr := &statusRecorder{ResponseWriter: w, status: 200}
		// TODO: next.ServeHTTP(sr, r)
		// TODO: log.Printf("%s %s -> %d %s", r.Method, r.URL.Path, sr.status, time.Since(started))

		next.ServeHTTP(w, r)
	})
}

// withRecovery turns panics into 500 responses so one bad handler doesn't kill the server.
func withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: defer func() {
		//           if rec := recover(); rec != nil {
		//               log.Printf("panic: %v", rec)
		//               http.Error(w, "internal server error", http.StatusInternalServerError)
		//           }
		//       }()
		next.ServeHTTP(w, r)
	})
}

// withRequestID either uses an incoming X-Request-ID or generates a fresh one,
// then echoes it back on the response. (Future examples will stash it in context.)
func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			// TODO: generate 16 random bytes -> hex.EncodeToString into id
			_ = rand.Read
			_ = hex.EncodeToString
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /hi", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hi")
	})
	mux.HandleFunc("GET /boom", func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})

	// TODO: build the chain. Order matters — outermost wrapper runs first on the
	// way in and last on the way out. Recommended:
	//   handler := withLogging(withRecovery(withRequestID(mux)))
	// Logging on the outside so it sees the final status; recovery just inside
	// so the logger can still report the 500 it produced.

	handler := http.Handler(mux)
	_ = withLogging
	_ = withRecovery
	_ = withRequestID
	_ = time.Now

	log.Fatal(http.ListenAndServe(":8080", handler))
}
