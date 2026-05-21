# 03 — Middleware

The most reused pattern in Go HTTP code:

```go
func middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // before
        next.ServeHTTP(w, r)
        // after
    })
}
```

It's just a function. `next` is the rest of the chain.

## Three classics

| Middleware     | Why                                                                                                          |
| -------------- | ------------------------------------------------------------------------------------------------------------ |
| **Logging**    | observability — without this, you have no idea what your server is doing in prod                             |
| **Recovery**   | one buggy handler shouldn't kill the whole process. `recover()` catches panics inside the deferred function. |
| **Request ID** | trace a request across services. Either honor an incoming `X-Request-ID` or mint a fresh one.                |

## Order matters

When you write `withLogging(withRecovery(withRequestID(mux)))`:

- Logging is **outermost** → it observes the final status code (after recovery converts the panic).
- Recovery is **inside** logging → so the 500 it produces is logged.
- Request-ID is **innermost** → so every log line and recovery report can include it (next example will wire context for that).

## How to capture the status code

`http.ResponseWriter` does not let you read back the status after `WriteHeader`. The trick is to wrap it in a small struct that intercepts `WriteHeader`:

```go
type statusRecorder struct {
    http.ResponseWriter
    status int
}

func (s *statusRecorder) WriteHeader(code int) {
    s.status = code
    s.ResponseWriter.WriteHeader(code)
}
```

This is the same trick `chi/middleware.Logger`, `gorilla/handlers.LoggingHandler`, and every prod logger uses.

## Comparison

| Concept        | Go                                      | Express                                                              | Flask                                        |
| -------------- | --------------------------------------- | -------------------------------------------------------------------- | -------------------------------------------- |
| Middleware     | `func(http.Handler) http.Handler`       | `app.use((req,res,next) => {...})`                                   | `@app.before_request` / `@app.after_request` |
| Panic recovery | `defer recover()` inside the middleware | unhandled exception → process crash unless you have an error handler | exception handler decorator                  |

## Run

```bash
go run .
curl -i http://localhost:8080/hi
curl -i http://localhost:8080/boom
```
