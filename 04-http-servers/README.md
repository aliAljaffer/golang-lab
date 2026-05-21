# 04 — HTTP Servers

> Status: ☑ scaffolded — examples + mini-project + exercises ready to implement. See [`PLAN.md`](./PLAN.md).

## What you'll learn

- Building HTTP servers with stdlib `net/http`
- Routing with `chi` or `gorilla/mux`
- Middleware (logging, recovery, auth)
- Graceful shutdown with `context.Context`
- Health/readiness probes (Kubernetes-friendly)
- Receiving and verifying webhooks (HMAC)

## Mental model from other languages

| Concept | Go | Python | TS / Node | Java |
|---|---|---|---|---|
| Minimal server | `net/http` | Flask | Express | Spring Boot |
| Router | `chi` / `gorilla/mux` | Flask / FastAPI routes | Express routes | Spring `@RestController` |
| Middleware | `func(next http.Handler) http.Handler` | Flask `before_request` | Express `app.use` | Servlet filters |
| Graceful shutdown | `srv.Shutdown(ctx)` | uvicorn lifespan | `server.close()` + drain | Spring lifecycle |

## The DevOps angle

You'll write servers for: webhook receivers (deploys, GitHub events), health-check sidecars, internal metrics scrapers, k8s admission controllers, custom operators. The patterns here are the foundation for all of those.

See [`PLAN.md`](./PLAN.md).
