// 02-chi-router — migration from stdlib mux to chi.
//
// Why bother? chi gives you:
//   - Cleaner path params: chi.URLParam(r, "id") instead of r.PathValue("id")
//   - Sub-routers / route groups for shared middleware
//   - A larger ecosystem of middleware (rate limit, CORS, etc.)
//
// Goal: serve GET /users/{id} and POST /users using chi.
//
// Run:
//   go run .
//   curl -i http://localhost:8080/users/42
//   curl -i -X POST http://localhost:8080/users -d '{"name":"ali"}'
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()

	// TODO: r.Get("/users/{id}", func(w http.ResponseWriter, req *http.Request) {
	//          id := chi.URLParam(req, "id")
	//          fmt.Fprintf(w, "user %s", id)
	//       })
	// TODO: r.Post("/users", func(w http.ResponseWriter, req *http.Request) {
	//          w.WriteHeader(http.StatusCreated)
	//          fmt.Fprintln(w, "created")
	//       })
	// TODO: r.Route("/admin", func(sub chi.Router) {
	//          sub.Get("/ping", func(w http.ResponseWriter, req *http.Request) {
	//              fmt.Fprintln(w, "pong (admin)")
	//          })
	//       })
	// TODO: log.Fatal(http.ListenAndServe(":8080", r))

	_ = r
	_ = fmt.Fprintf
	_ = log.Fatal
	_ = http.StatusOK
}
