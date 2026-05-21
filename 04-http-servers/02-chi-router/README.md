# 02 — chi router

Same server as 01, but routed through [`chi`](https://github.com/go-chi/chi).

## When you'd reach for chi

- You want **route groups** with shared middleware: e.g. `/admin/*` behind auth.
- You want **path params** without the Go 1.22 mux quirks (e.g. multi-segment captures).
- You want the **chi middleware ecosystem** — `middleware.Logger`, `middleware.Recoverer`, `middleware.Throttle`, etc.

For a single `/hello` endpoint, stdlib is enough. For anything more than a handful of routes with shared concerns, chi pays for itself fast.

## Comparison

| Concept | stdlib mux (Go 1.22+) | chi |
|---|---|---|
| Path param | `r.PathValue("id")` | `chi.URLParam(r, "id")` |
| Route group | not built in | `r.Route("/admin", func(sub chi.Router) {...})` |
| Method routing | `"GET /users"` pattern | `r.Get("/users", ...)` / `r.Post(...)` |
| Middleware | wrap by hand | `r.Use(mw)` |

## Run

```
go run .
curl -i http://localhost:8080/users/42
curl -i -X POST http://localhost:8080/users -d '{"name":"ali"}'
curl -i http://localhost:8080/admin/ping
```
