# 04 — HTTP Servers

> Status: ☑ scaffolded — examples + mini-project + exercises ready to implement. See [`PLAN.md`](./PLAN.md).

The other half of section 03. You'll write servers for: webhook receivers (deploy hooks, GitHub events), health-check sidecars, internal metrics scrapers, Kubernetes admission controllers, and one-off operators. The patterns here — handlers, middleware, graceful shutdown, health endpoints, HMAC verification — are the foundation for all of them.

The stdlib `net/http` is enough for a single-route server. Anything with path parameters, method routing, or middleware composition wants a third-party router — this section uses `chi`, which is the closest thing to a community default.

---

## What you'll learn

- `http.HandleFunc` and `http.ServeMux` — stdlib basics
- Why a router lib (`chi`, `gorilla/mux`) — path params, per-method handlers, route groups
- The middleware pattern: `func(http.Handler) http.Handler`
- Common middleware: structured logging, panic recovery (panic → 500), request ID, timeout, CORS, auth
- Graceful shutdown: signal handling + `srv.Shutdown(ctx)`
- Health endpoints — the Kubernetes split: `/healthz` (liveness) vs `/readyz` (readiness with dependency checks)
- Reading JSON request bodies, validating, returning structured JSON errors
- Webhook verification with HMAC-SHA256 (the GitHub / Slack / Stripe pattern)
- Server timeouts: `ReadTimeout`, `WriteTimeout`, `IdleTimeout` — the production-only details that the default `http.Server{}` skips

---

## Mental model from other languages

| Concept            | Go                                       | Python                  | TS / Node               | Java                    |
| ------------------ | ---------------------------------------- | ----------------------- | ----------------------- | ----------------------- |
| Minimal server     | `net/http`                               | Flask                   | Express                 | Spring Boot             |
| Router             | `chi` / `gorilla/mux`                    | Flask / FastAPI routes  | Express routes          | Spring `@RestController` |
| Middleware         | `func(next http.Handler) http.Handler`   | Flask `before_request`  | Express `app.use`       | Servlet filters         |
| Path params        | `chi.URLParam(r, "id")`                  | `<int:id>` in route     | `req.params.id`         | `@PathVariable`         |
| Graceful shutdown  | `srv.Shutdown(ctx)`                      | uvicorn lifespan        | `server.close()` + drain | Spring lifecycle       |
| JSON request body  | `json.NewDecoder(r.Body).Decode(&v)`     | `request.json`          | `req.body` (after parser) | Jackson `@RequestBody` |

**The cultural difference:** there is no Rails/Django/Spring monolith for Go. The community deliberately stays close to `net/http` — you compose `http.Handler` values rather than configure a framework. The upside: nothing magical, you can read the entire stack; the downside: every project rolls its own logging, error format, and request-id middleware unless they pick `chi` + its middleware bundle.

---

## The DevOps angle

Webhook receivers are the workhorse: GitHub fires "deployment requested" → your tool verifies the HMAC signature, looks up the configured command, runs it, reports back. The same shape covers PagerDuty incident webhooks, Slack slash commands, Stripe payment events.

The non-obvious production details:

- **Always set `ReadTimeout` and `WriteTimeout`.** A slow client holding a TCP connection open ties up a goroutine forever. The default `http.Server{}` lets it.
- **`srv.Shutdown(ctx)` is non-negotiable in containers.** SIGTERM arrives, you have ~30 seconds before SIGKILL; if you don't drain in-flight requests, you serve 502s during every deploy.
- **`/healthz` should not check downstream dependencies.** If your DB is down, the k8s liveness probe restarts your pod, then your replacement also can't reach the DB and gets restarted, and you cascade-fail your way through the deployment. Liveness = "this process is alive and not deadlocked." Readiness (`/readyz`) is where dependency checks belong.
- **HMAC verification with `hmac.Equal`, never `bytes.Equal`.** The constant-time comparison is what stops timing-attack signature recovery — see example 07.

---

## Walkthrough

Read these in order. Each is a runnable example demonstrating one specific concept.

1. [`01-hello-server/`](./01-hello-server/) — `net/http` stdlib server, one route. Shows the `http.HandlerFunc` adapter and the `http.ListenAndServe(addr, mux)` shape that every Go server starts from.
2. [`02-chi-router/`](./02-chi-router/) — migration to `chi`: path params (`/users/{id}`), route groups, per-method handlers (`r.Get`, `r.Post`). Why anything bigger than a single route grows out of `http.ServeMux`.
3. [`03-middleware/`](./03-middleware/) — the canonical middleware chain: structured logging + panic recovery + request-id injection. The composition (`Use(Logger, Recoverer, RequestID)`) reads top-to-bottom.
4. [`04-graceful-shutdown/`](./04-graceful-shutdown/) — SIGTERM-aware shutdown that drains in-flight requests. `signal.NotifyContext` + `srv.Shutdown(ctx)` is the modern recipe.
5. [`05-health-endpoints/`](./05-health-endpoints/) — `/healthz` (process alive) vs `/readyz` (dependencies reachable). The deliberate split is what keeps Kubernetes from cascade-restarting your fleet when a database flickers.
6. [`06-json-api/`](./06-json-api/) — POST endpoint with body validation + JSON errors. `DisallowUnknownFields()`, max-bytes reading via `http.MaxBytesReader` (defends against unbounded uploads), structured error envelope.
7. [`07-webhook-receiver/`](./07-webhook-receiver/) — verify a GitHub-style HMAC signature, reject bad sigs with 401. The `hmac.Equal` constant-time comparison is the whole point.

---

## Mini-project: [`webhook-runner`](./mini-project/)

An HTTP server that receives webhooks (HMAC-verified), looks up a configured command in a YAML file, runs it, and returns the exit code + truncated output. A tiny CI runner.

The point: this exercises the whole section together — routing, HMAC verification, request body parsing, subprocess execution (from section 02), graceful shutdown, and structured logging. Tests pin: reject unsigned/badly-signed with 401, runs configured commands and captures exit codes, graceful shutdown drains in-flight executions.

Spec and starter in [`mini-project/`](./mini-project/). Section 10 (`webhook-runner-instrumented`) extends this one with full Prometheus + OpenTelemetry instrumentation.

---

## Exercises

See [`exercises/`](./exercises/):

1. **[`01-rate-limit-middleware`](./exercises/01-rate-limit-middleware/)** — IP-based rate limiter as middleware. Token bucket per remote address; the foundation of every "1000 req/min" feature flag.
2. **[`02-basic-auth`](./exercises/02-basic-auth/)** — middleware that gates routes behind HTTP Basic Auth using **constant-time** comparison. Same `hmac.Equal`-or-`subtle.ConstantTimeCompare` lesson as webhook verification.
3. **[`03-request-tracing`](./exercises/03-request-tracing/)** — middleware that injects a request ID into `context.Context` and logs the full lifecycle. Sets up the patterns that section 10 expands into real OpenTelemetry tracing.

---

## Further reading

- [`net/http` docs](https://pkg.go.dev/net/http) — the stdlib reference for both client (section 03) and server
- [`chi` docs](https://github.com/go-chi/chi) — the router used here; small, idiomatic, plays well with `net/http`
- [Cloudflare: timeouts](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/) — the same post as section 03, but the server timeouts are the half that matters here
- [Kubernetes liveness/readiness probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/) — the doc that motivates the `/healthz` vs `/readyz` split
- [GitHub webhooks: securing](https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries) — the HMAC scheme example 07 mirrors
