# Plan: 04-http-servers

## Concepts to cover

- [ ] `http.HandleFunc`, `http.ServeMux` — stdlib basics
- [ ] Why use a router lib (`chi`, `gorilla/mux`) — path params, method routing
- [ ] Middleware pattern: `func(http.Handler) http.Handler`
- [ ] Common middleware: logging, recovery (panic → 500), request ID, timeout, CORS, auth
- [ ] Graceful shutdown: signal handling + `srv.Shutdown(ctx)`
- [ ] Health endpoints: `/healthz` (liveness) vs `/readyz` (readiness)
- [ ] Reading JSON request bodies, validating, returning JSON errors
- [ ] Webhook verification (HMAC-SHA256) — GitHub, Slack, Stripe pattern
- [ ] Server timeouts: `ReadTimeout`, `WriteTimeout`, `IdleTimeout` (matters in prod)

## Examples to build

| Folder | Demonstrates |
|---|---|
| `01-hello-server/` | `net/http` stdlib server, one route |
| `02-chi-router/` | Migration to chi: path params, route groups |
| `03-middleware/` | Logging + recovery + request-id middleware chain |
| `04-graceful-shutdown/` | SIGTERM-aware shutdown with in-flight request draining |
| `05-health-endpoints/` | livez/readyz with dependency checks |
| `06-json-api/` | POST endpoint with body validation + JSON errors |
| `07-webhook-receiver/` | Verify GitHub-style HMAC signature, reject bad sigs |

## Mini-project

**`webhook-runner`** — HTTP server that receives webhooks (HMAC-verified), looks up a configured command from a YAML config, runs it, and returns the exit code + truncated output. Like a tiny CI runner.

Tests verify:
- Rejects unsigned/badly-signed requests with 401
- Runs configured commands, captures exit codes
- Graceful shutdown drains in-flight executions

## Exercises

1. **`01-rate-limit-middleware`** — implement an IP-based rate limiter as middleware
2. **`02-basic-auth`** — middleware that gates routes behind basic auth (constant-time comparison)
3. **`03-request-tracing`** — middleware that injects a request ID into context and logs the full lifecycle

## Status

- [x] Concepts in README walkthrough
- [x] Examples 01-07 scaffolded (TODO blocks — user to fill in)
- [x] Mini-project `webhook-runner` scaffolded + tests written (failing)
- [x] Exercises scaffolded

## Session Log

When a Claude session does work in this section, append an entry to the root [`SESSIONS.md`](../SESSIONS.md) before ending — do **not** log session history in this file. `PLAN.md` is the plan; `SESSIONS.md` is the history. Tick the Status boxes above as items complete.
