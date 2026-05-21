# 01 — Hello server

The 30-second version of an HTTP server in Go.

## Things to notice

- `http.ServeMux` is the router. No framework, no decorators — it's a plain struct that satisfies `http.Handler`.
- `http.HandleFunc` registers a function as a handler. Behind the scenes it wraps it in `http.HandlerFunc`, which is just `type HandlerFunc func(ResponseWriter, *Request)` with a `ServeHTTP` method — the adapter pattern.
- Since Go 1.22 the stdlib mux supports `"GET /path"` patterns and path wildcards like `"GET /users/{id}"`. For most CRUD servers you no longer need a router library.
- `http.ListenAndServe(addr, handler)` blocks until the server errors. That's why production code (later examples) builds an `*http.Server` explicitly and runs `srv.ListenAndServe()` in a goroutine.

## Comparison

| Concept           | Go                                  | Python (Flask)         | TS (Express)           | Java (Spring)               |
| ----------------- | ----------------------------------- | ---------------------- | ---------------------- | --------------------------- |
| Minimal server    | `http.ListenAndServe(":8080", mux)` | `app.run()`            | `app.listen(8080)`     | `SpringApplication.run()`   |
| Route a path      | `mux.HandleFunc("GET /hello", h)`   | `@app.route("/hello")` | `app.get("/hello", h)` | `@GetMapping("/hello")`     |
| Handler signature | `func(w, r)`                        | `def handler():`       | `(req, res) => {}`     | `@GetMapping ... returns X` |

## Run

```bash
go run .
curl -i http://localhost:8080/hello
curl -i 'http://localhost:8080/echo?msg=hi'
```
