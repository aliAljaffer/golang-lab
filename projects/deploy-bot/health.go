package main

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// HealthChecker probes a URL until it returns a 2xx OR ctx is done.
// Production impl is *HTTPHealthChecker; tests use a fake returning a
// programmed sequence of attempt results.
type HealthChecker interface {
	Probe(ctx context.Context, url string) error
}

// HTTPHealthChecker polls url every Interval. Returns nil on the first 2xx;
// returns an error if ctx fires first or if Sleep returns an error.
//
// Backoff is fixed (Interval). If you wanted exponential backoff this would
// be the place — for the bootcamp scaffold, fixed-interval polling is
// enough and keeps the test surface deterministic.
//
// Sleep is injected so tests don't actually wait. Production wires
// Sleep = ctxSleep from main.go.
type HTTPHealthChecker struct {
	Client   *http.Client
	Interval time.Duration

	// Sleep yields the goroutine for d, or returns ctx.Err() if ctx fires
	// first. Tests pass a fastSleep that ignores d.
	Sleep func(ctx context.Context, d time.Duration) error
}

// Probe loops:
//
//	for {
//	    GET url
//	    if 2xx -> return nil
//	    if ctx done before next attempt -> return ctx.Err()
//	    Sleep(Interval)
//	}
//
// Behavior contract (each pinned by a test in main_test.go):
//   - 2xx on the first attempt    -> return nil; one GET happened.
//   - 5xx N times then 2xx        -> return nil after N+1 GETs.
//   - never-2xx + ctx times out   -> return a non-nil error containing
//                                    "health probe" (so the caller can log it).
//   - ctx cancelled mid-Sleep     -> Sleep returns ctx.Err(); Probe propagates.
func (h *HTTPHealthChecker) Probe(ctx context.Context, url string) error {
	// TODO: implement the poll loop. The contract above is the spec; the
	//   tests pin each branch. Two non-obvious bits:
	//     - any non-2xx (including transport errors) means "keep polling",
	//       not "fail" — the container might just not be ready yet.
	//     - h.Sleep is ctx-aware; if it returns an error, propagate it
	//       instead of looping again. That's what makes ctx-cancel-mid-sleep
	//       wake up promptly.
	return errors.New("HTTPHealthChecker.Probe not implemented")
}
