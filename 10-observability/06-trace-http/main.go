// 06-trace-http — propagate trace context across an HTTP boundary.
//
// What this example proves:
//   - The OTel **propagator** is what serializes a span's TraceID + SpanID
//     into HTTP headers (`traceparent` / `tracestate`, per W3C) on the
//     client side, and parses them back into a ctx on the server side.
//   - `otelhttp.NewHandler(h, "name")` wraps a server handler to extract
//     the incoming traceparent header AND auto-start a server span for
//     the request. Mirror: `otelhttp.NewTransport(rt)` does the same on
//     the client side.
//   - The whole point: if Service A calls Service B, B's server span
//     becomes a child of A's client span, in the SAME trace.
//
// The propagator must be set globally:
//
//	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
//	    propagation.TraceContext{},  // W3C traceparent
//	    propagation.Baggage{},        // OTel baggage
//	))
//
//	If you forget this, span context never reaches the wire. otelhttp uses
//	the globally-configured propagator by default.
//
// Run:
//
//	go run .
//	# in another terminal:
//	curl localhost:8080/parent
//	# observe: 3 spans in stdout — client(parent) -> server(child) -> business(grandchild)
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func initTracer() (*sdktrace.TracerProvider, error) {
	exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)
	// TODO: otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
	// TODO:     propagation.TraceContext{}, propagation.Baggage{},
	// TODO: ))
	return tp, nil
}

// childHandler simulates the "downstream service". It's wrapped in
// otelhttp.NewHandler so the incoming traceparent header becomes a real
// span context in its r.Context().
func childHandler(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("example/child")
	// TODO: _, span := tracer.Start(r.Context(), "child-business-logic")
	// TODO: defer span.End()
	time.Sleep(10 * time.Millisecond)
	fmt.Fprintln(w, "ok")
	_ = tracer
}

// parentHandler simulates the "upstream service". It calls childHandler
// via a tracing HTTP client (otelhttp.NewTransport) so the outgoing request
// carries traceparent.
func parentHandler(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("example/parent")
	ctx, span := tracer.Start(r.Context(), "parent-business-logic")
	defer span.End()

	// HTTP client wrapped with otelhttp.NewTransport — outgoing requests get
	// traceparent injected from ctx automatically.
	client := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
	// TODO: req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/child", nil)
	// TODO: resp, err := client.Do(req)
	// TODO: if err != nil { http.Error(w, err.Error(), 500); return }
	// TODO: defer resp.Body.Close()
	// TODO: io.Copy(w, resp.Body)

	_ = ctx
	_ = client
	_ = io.Copy
}

func main() {
	tp, err := initTracer()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = tp.Shutdown(shutdownCtx)
	}()

	mux := http.NewServeMux()
	// otelhttp wraps the handler: parses incoming traceparent, starts a
	// server span, makes ctx available via r.Context().
	mux.Handle("/parent", otelhttp.NewHandler(http.HandlerFunc(parentHandler), "GET /parent"))
	mux.Handle("/child", otelhttp.NewHandler(http.HandlerFunc(childHandler), "GET /child"))

	addr := ":8080"
	log.Println("listening on", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}

	_ = propagation.TraceContext{}
}
