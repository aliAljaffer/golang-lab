// 05-otel-tracing — set up an OTel tracer with a stdout exporter; trace a
// "service" that has a parent span and two child spans.
//
// What this example proves:
//   - Three things are needed to start tracing: an **exporter** (where
//     spans go), a **tracer provider** (a `*sdktrace.TracerProvider`,
//     configured with the exporter), and a **tracer** (obtained from the
//     provider, this is what you actually use to start spans).
//   - `tracer.Start(ctx, "name")` returns BOTH a new ctx AND the span.
//     YOU MUST USE THE RETURNED CTX for any child operation; the parent
//     ctx still points at the parent span.
//   - `defer span.End()` — without this, the span never closes and never
//     ships to the exporter.
//   - Parent/child relationships are derived from the ctx, not passed
//     explicitly. This is the whole point of context propagation.
//
// Why stdout exporter for the example:
//
//	A real OTel deployment ships spans to a collector (OTLP gRPC / HTTP)
//	which then routes to Jaeger / Tempo / Honeycomb / etc. For learning,
//	stdouttrace prints the span JSON to your terminal — no infra needed.
//	Swap to `otlptracegrpc` for production.
//
// Run:
//
//	go run .
//	# expect 3 span JSON blobs: handleRequest (root), validate, fetchData
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func initTracer() (*sdktrace.TracerProvider, error) {
	// TODO: exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	// TODO: if err != nil { return nil, err }

	// TODO: tp := sdktrace.NewTracerProvider(
	// TODO:     sdktrace.WithBatcher(exp),
	// TODO:     // For learning, AlwaysSample so we see every span. In prod, use
	// TODO:     // ParentBased(TraceIDRatioBased(0.01)) or similar.
	// TODO:     sdktrace.WithSampler(sdktrace.AlwaysSample()),
	// TODO: )
	// TODO: otel.SetTracerProvider(tp)
	// TODO: return tp, nil

	_ = stdouttrace.New
	_ = sdktrace.NewTracerProvider
	_ = otel.SetTracerProvider
	return nil, nil
}

// validate is a "child operation". It starts a span using the ctx it receives —
// which carries the parent span — so its span will be linked as a child.
func validate(ctx context.Context, payload string) error {
	tracer := otel.Tracer("example/service")
	// TODO: _, span := tracer.Start(ctx, "validate")
	// TODO: defer span.End()
	// TODO: span.SetAttributes(attribute.Int("payload.len", len(payload)))
	time.Sleep(10 * time.Millisecond)
	_ = tracer
	_ = payload
	_ = attribute.Int
	return nil
}

func fetchData(ctx context.Context) (string, error) {
	tracer := otel.Tracer("example/service")
	// TODO: _, span := tracer.Start(ctx, "fetchData")
	// TODO: defer span.End()
	time.Sleep(20 * time.Millisecond)
	_ = tracer
	return "ok", nil
}

func handleRequest(ctx context.Context) error {
	tracer := otel.Tracer("example/service")
	// TODO: ctx, span := tracer.Start(ctx, "handleRequest")
	// TODO: defer span.End()
	// TODO: span.SetAttributes(attribute.String("user.id", "u1"))

	if err := validate(ctx, "hello"); err != nil {
		return err
	}
	if _, err := fetchData(ctx); err != nil {
		return err
	}
	_ = tracer
	return nil
}

func main() {
	tp, err := initTracer()
	if err != nil {
		log.Fatal(err)
	}
	if tp != nil {
		defer func() {
			// Shutdown flushes buffered spans. WITHOUT this, the batcher
			// may exit before flushing — you see no output.
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = tp.Shutdown(shutdownCtx)
		}()
	}

	if err := handleRequest(context.Background()); err != nil {
		log.Fatal(err)
	}
	fmt.Println("done")
}
