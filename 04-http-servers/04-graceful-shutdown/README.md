# 04 — Graceful shutdown

The piece every prod HTTP server needs and every tutorial skips.

## Why

When Kubernetes rolls a pod:

1. It sends **SIGTERM** to your process.
2. It waits up to `terminationGracePeriodSeconds` (default 30s).
3. If you haven't exited, it sends **SIGKILL** — in-flight requests die mid-response.

Naive `http.ListenAndServe` aborts open connections the moment the process exits. Graceful shutdown means: stop accepting new requests, let active ones finish, then exit.

## The pattern

```go
srv := &http.Server{Addr: ":8080", Handler: mux}

errCh := make(chan error, 1)
go func() { errCh <- srv.ListenAndServe() }()

sigCh := make(chan os.Signal, 1)
signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

select {
case err := <-errCh:
    log.Fatal(err)
case <-sigCh:
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    if err := srv.Shutdown(ctx); err != nil {
        log.Printf("shutdown error: %v", err)
    }
}
```

## Details that bite

- `srv.ListenAndServe()` returns `http.ErrServerClosed` after a clean shutdown. That's **not** an error — filter it out.
- The shutdown context is the _grace deadline_. Pick a value shorter than k8s's `terminationGracePeriodSeconds`.
- Handlers that need to abort early on shutdown should select on `r.Context().Done()`. Shutdown cancels per-request contexts when it gives up waiting.
- Background workers (not handlers) need their own shutdown signal — `srv.Shutdown` only knows about HTTP.

## Comparison

| Stack            | Graceful shutdown hook                   |
| ---------------- | ---------------------------------------- |
| Go               | `srv.Shutdown(ctx)`                      |
| Node             | `server.close(callback)` + drain timeout |
| Python (uvicorn) | lifespan `shutdown` event                |
| Spring Boot      | `server.shutdown=graceful`               |

## Run

```bash
go run .

# in another terminal:
curl http://localhost:8080/slow &
sleep 1
kill -TERM $(pgrep -f 04-graceful-shutdown)
# /slow should still finish — process exits after.
```
