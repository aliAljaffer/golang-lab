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
	// TODO: emit one JSON line per call. The newline matters — downstream
	//   `jq -c` / log aggregators key on it. ctx is ignored intentionally.
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
	// TODO: POST the JSON-encoded alert per the contract above. The
	//   classification matters: 4xx is permanent (no retry — usually a
	//   bad webhook URL); 5xx + transport errors are transient. ctx
	//   cancellation must bail out mid-retry, not finish the budget.
	_ = bytes.NewReader
	_ = json.Marshal
	return errors.New("WebhookSink.Send not implemented")
}
