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
	// TODO: client := h.Client (or http.DefaultClient if nil).
	// TODO: for {
	// TODO:     if err := ctx.Err(); err != nil { return err }.
	// TODO:     req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil).
	// TODO:     resp, err := client.Do(req).
	// TODO:     if err == nil {
	// TODO:         _ = resp.Body.Close().
	// TODO:         if resp.StatusCode >= 200 && resp.StatusCode < 300 { return nil }.
	// TODO:     }
	// TODO:     // Sleep before retrying. Sleep is ctx-aware; surfaces ctx.Err() if cancelled.
	// TODO:     if err := h.Sleep(ctx, h.Interval); err != nil { return err }.
	// TODO: }
	return errors.New("HTTPHealthChecker.Probe not implemented")
}
