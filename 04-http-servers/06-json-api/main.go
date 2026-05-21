// 06-json-api — POST endpoint with JSON body + validation + JSON errors.
//
// Goal: POST /users { "name": "...", "email": "..." } returns 201 + the created user.
//   - Reject body > 1 MiB (DoS guard).
//   - Reject unknown JSON fields (catches typos in API consumers' code).
//   - Validate required fields, return 400 + JSON error.
//
// Run:
//   go run .
//   curl -i -X POST http://localhost:8080/users -H 'Content-Type: application/json' \
//        -d '{"name":"ali","email":"ali@example.com"}'
//   curl -i -X POST http://localhost:8080/users -d '{"name":""}'             # 400
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type apiError struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	// TODO: cap body size: r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	// TODO: dec := json.NewDecoder(r.Body)
	// TODO: dec.DisallowUnknownFields()
	// TODO: var u User; if err := dec.Decode(&u); err != nil { writeJSON(w, 400, apiError{err.Error()}); return }
	// TODO: validate u.Name != "" and u.Email contains "@" — otherwise 400 with apiError
	// TODO: pretend to persist; writeJSON(w, 201, u)

	_ = json.NewDecoder
	writeJSON(w, http.StatusNotImplemented, apiError{Error: "TODO"})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /users", createUser)

	fmt.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
