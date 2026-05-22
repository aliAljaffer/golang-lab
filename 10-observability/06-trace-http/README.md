# 06 — propagating trace context over HTTP

A trace is only useful if it spans services. This example shows how the
`traceparent` header carries the trace context across an HTTP boundary so
the receiving service's span is linked to the caller's span.

## The W3C `traceparent` header

```
traceparent: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01
             ^^ ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^ ^^^^^^^^^^^^^^^^ ^^
             version trace-id (16 bytes hex)   span-id (8 hex)  flags
```

W3C standardized this in 2020. Every modern tracing library reads/writes
it. The OTel propagator is `propagation.TraceContext{}`.

## The two halves

| Half | What it does | API |
|---|---|---|
| Server | Read `traceparent` from incoming headers → ctx | `otelhttp.NewHandler(h, name)` |
| Client | Write `traceparent` from ctx → outgoing headers | `otelhttp.NewTransport(rt)` |

```go
// Server wrap
mux.Handle("/x", otelhttp.NewHandler(http.HandlerFunc(myHandler), "GET /x"))

// Client wrap
client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
```

That's it. `otelhttp` ALSO auto-starts a server span (for the wrapped
handler) and a client span (for each Do call). So you get two spans for
free per service hop, on top of any spans your handler creates manually.

## The global propagator gotcha

```go
otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
    propagation.TraceContext{},  // W3C traceparent
    propagation.Baggage{},        // OTel baggage (key/value attrs that ride along)
))
```

`otelhttp` uses `otel.GetTextMapPropagator()` by default — and the default
default is **no-op**. If you skip this `SetTextMapPropagator` line, headers
never get injected and your traces don't cross service boundaries. The
single most common "why isn't my distributed trace working" cause.

## Three spans per request in this example

When `curl /parent` hits this server:

1. `GET /parent` — server span (created by `otelhttp.NewHandler`).
2. `parent-business-logic` — manual span inside `parentHandler`.
3. `HTTP GET` — client span (created by `otelhttp.NewTransport` when
   `client.Do` runs).
4. `GET /child` — server span (created by the other `otelhttp.NewHandler`).
5. `child-business-logic` — manual span inside `childHandler`.

5 spans, one trace, one trace ID. The `traceparent` header is what stitches
3→4 across the network.

## Baggage

`propagation.Baggage{}` (the second arg to the composite propagator) carries
arbitrary key/value pairs across services in a separate `baggage:` header.
Use for "tenant_id", "feature_flag", or anything you want to read in
downstream services without re-fetching. Costs bytes on every hop; don't
abuse it.

## Common shape with HTTP middleware order

```go
mux.Handle("/x",
    otelhttp.NewHandler(            // 1. extract traceparent, start server span
        loggingMiddleware(           // 2. start request log line (now has trace_id)
            metricsMiddleware(       // 3. record duration / count
                http.HandlerFunc(myHandler),
            ),
        ),
        "GET /x",
    ),
)
```

Always wrap `otelhttp.NewHandler` OUTERMOST. Inner middleware can then read
the span context from `r.Context()` and add `trace_id` to log lines.

## TODO

1. Uncomment the TODO blocks (the propagator setup, the child span in
   `childHandler`, the client.Do call in `parentHandler`).
2. Run `go run .`. In another terminal: `curl localhost:8080/parent`.
3. In the server's stdout, find the five spans above. Verify all share one
   `TraceID`. Verify the child server span's `ParentSpanID` matches the
   parent's client span ID.
4. Comment out the `otel.SetTextMapPropagator` line and re-run. Observe
   the child span's `ParentSpanID` is zero — the trace is broken.
5. Manually inspect the outgoing header with `tcpdump -A -i lo port 8080`
   or by adding a `roundTripperFunc` that logs `req.Header`. Confirm
   `traceparent` is present.
