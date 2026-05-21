// 01-hello-server — the simplest possible HTTP server using stdlib `net/http`.
//
// Goal: serve "hello, world" on GET /hello, listen on :8080.
//
// Run:
//   go run .
//   curl -i http://localhost:8080/hello
//
// Notes for veterans of Python/Node/Java:
//   - There is no framework. `http.ServeMux` is the router.
//   - `http.HandleFunc` adapts a plain function into an `http.Handler`.
//   - Since Go 1.22 the mux supports method+path patterns: "GET /hello".
package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	// TODO: mux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
	//          fmt.Fprintln(w, "hello, world")
	//       })
	// TODO: also add "GET /echo" that returns the value of ?msg= from the query string.
	// TODO: log.Fatal(http.ListenAndServe(":8080", mux))

	_ = mux
	_ = fmt.Fprintln
	_ = log.Fatal
}
