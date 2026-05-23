package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// Sink is anything that can deliver an Alert. Implementations below.
type Sink interface {
	Send(ctx context.Context, alert Alert) error
}

// StdoutSink writes one JSON-encoded line per alert to Out. Used by --dry-run.
type StdoutSink struct {
	Out io.Writer
}

// Send writes one JSON line. Concurrency-safe iff Out is.
func (s *StdoutSink) Send(_ context.Context, alert Alert) error {
	// TODO: marshal alert to JSON.
	// TODO: write the bytes followed by a single newline to s.Out.
	// TODO: return any error from either step.
	return errors.New("StdoutSink.Send not implemented")
}

// WebhookSink POSTs JSON-encoded alerts to a Slack incoming-webhook URL.
// Non-2xx is an error. Network errors propagate. Client == nil means
// http.DefaultClient.
//
// The retry/backoff behavior is the student's job — there's a TODO inside
// Send below for it.
type WebhookSink struct {
	URL    string
	Client *http.Client

	// MaxRetries is the cap on retries for transient failures (5xx, network
	// errors). 0 means "no retries — fail fast on the first error". Tests
	// pass 0 unless they're specifically asserting retry behavior.
	MaxRetries int
}

// Send POSTs a single alert. Behavior contract:
//   - 2xx           -> nil
//   - 4xx           -> error, no retry (permanent failure)
//   - 5xx / netErr  -> retry up to MaxRetries times with backoff, then error
//   - ctx.Done()    -> abort and return ctx.Err()
func (s *WebhookSink) Send(ctx context.Context, alert Alert) error {
	// TODO: marshal alert into a buffer.
	// TODO: build a *http.Request with method POST, Content-Type application/json, body=buf, using ctx.
	// TODO: pick the client (s.Client or http.DefaultClient).
	// TODO: do the request; check status code; on 2xx return nil; on 4xx return a permanent error.
	// TODO: on 5xx or transport error, retry up to MaxRetries with exponential backoff (and ctx-aware sleep).
	// TODO: surface the final error if every attempt failed.
	_ = bytes.NewReader
	_ = json.Marshal
	return errors.New("WebhookSink.Send not implemented")
}
