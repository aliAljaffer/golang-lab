# 05 — Health endpoints

`/healthz` ≠ `/readyz`. Kubernetes uses them for different decisions.

## The two probes

| Probe     | k8s field        | Question                        | If it fails                                    |
| --------- | ---------------- | ------------------------------- | ---------------------------------------------- |
| Liveness  | `livenessProbe`  | "is the process alive?"         | k8s restarts the pod                           |
| Readiness | `readinessProbe` | "should this pod take traffic?" | k8s removes the pod from the Service endpoints |

## Why they must differ

**Liveness must be cheap and self-contained.** If `/healthz` calls the DB and the DB has a 30s blip, every pod restarts — making the outage worse. Liveness usually returns 200 immediately.

**Readiness can be smart.** It can fail when:

- A dependency is down — pull this pod out of rotation
- The process is starting up and warming caches — don't send traffic yet
- The process is shutting down — flip readiness to false _before_ `srv.Shutdown(ctx)` so k8s drains traffic first

## The shutdown dance

```go
sig := <-sigCh
readiness.Store(false)             // 1. Tell k8s to stop sending requests.
time.Sleep(5 * time.Second)        // 2. Wait for endpoints to propagate.
srv.Shutdown(shutdownCtx)          // 3. Now drain in-flight + exit.
```

Without the readiness flip, you'd drain in-flight requests but k8s would keep sending new ones for a few seconds until the next probe period — and those would hit a closing server.

## Comparison

| Stack                | Liveness equivalent         | Readiness equivalent         |
| -------------------- | --------------------------- | ---------------------------- |
| Spring Boot Actuator | `/actuator/health/liveness` | `/actuator/health/readiness` |
| Node                 | hand-rolled                 | hand-rolled                  |
| Python (FastAPI)     | hand-rolled                 | hand-rolled                  |

## Run

```bash
go run .
curl -i http://localhost:8080/healthz       # 200
curl -i http://localhost:8080/readyz        # 200
curl -X POST http://localhost:8080/admin/drain
curl -i http://localhost:8080/readyz        # 503
curl -i http://localhost:8080/healthz       # still 200 — process is alive
```
