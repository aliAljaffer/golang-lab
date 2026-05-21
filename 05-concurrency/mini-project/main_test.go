//go:build exercise

package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCheck_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()

	r := Check(context.Background(), srv.Client(), srv.URL)
	if r.Err != nil {
		t.Fatalf("Err = %v, want nil", r.Err)
	}
	if r.Status != 200 {
		t.Errorf("Status = %d, want 200", r.Status)
	}
	if r.URL != srv.URL {
		t.Errorf("URL = %q, want %q", r.URL, srv.URL)
	}
	if r.Duration <= 0 {
		t.Errorf("Duration = %v, want > 0", r.Duration)
	}
}

func TestCheck_NonOKStatusIsNotAnError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(503)
	}))
	defer srv.Close()

	r := Check(context.Background(), srv.Client(), srv.URL)
	if r.Err != nil {
		t.Errorf("Err = %v, want nil (5xx is a successful HTTP exchange, not a transport error)", r.Err)
	}
	if r.Status != 503 {
		t.Errorf("Status = %d, want 503", r.Status)
	}
}

func TestCheck_TimeoutSetsErr(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	client := &http.Client{Timeout: 30 * time.Millisecond}
	r := Check(context.Background(), client, srv.URL)
	if r.Err == nil {
		t.Fatal("Err = nil, want a timeout error")
	}
	if r.Status != 0 {
		t.Errorf("Status = %d, want 0 (transport failed)", r.Status)
	}
}

// TestRun_RespectsConcurrencyLimit drives 12 requests against a server that
// records the peak number of in-flight handlers. With --concurrency=3, peak
// must never exceed 3.
func TestRun_RespectsConcurrencyLimit(t *testing.T) {
	const concurrency = 3
	const nURLs = 12

	var inFlight atomic.Int32
	var peak atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cur := inFlight.Add(1)
		// Track max in-flight via compare-and-swap.
		for {
			p := peak.Load()
			if cur <= p || peak.CompareAndSwap(p, cur) {
				break
			}
		}
		time.Sleep(40 * time.Millisecond) // hold the slot
		inFlight.Add(-1)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	urls := make([]string, nURLs)
	for i := range urls {
		urls[i] = srv.URL
	}

	results := Run(context.Background(), srv.Client(), urls, concurrency)
	got := 0
	for range results {
		got++
	}
	if got != nURLs {
		t.Errorf("got %d results, want %d", got, nURLs)
	}
	if p := peak.Load(); p > concurrency {
		t.Errorf("peak in-flight = %d, want <= %d", p, concurrency)
	}
	if p := peak.Load(); p < 2 {
		t.Errorf("peak in-flight = %d, want >= 2 (Run does not appear to be parallel)", p)
	}
}

// TestRun_PerRequestTimeout — the client's Timeout caps each request; slow
// handlers should report an Err'd Result, not block the whole batch.
func TestRun_PerRequestTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
	}))
	defer srv.Close()

	client := &http.Client{Timeout: 40 * time.Millisecond}

	urls := []string{srv.URL, srv.URL, srv.URL}
	start := time.Now()
	results := Run(context.Background(), client, urls, 3)
	got := 0
	for r := range results {
		got++
		if r.Err == nil {
			t.Errorf("Result %d: Err = nil, want timeout", got)
		}
	}
	elapsed := time.Since(start)
	if got != 3 {
		t.Errorf("got %d results, want 3", got)
	}
	// All three should run in parallel with the 40ms client timeout — total
	// elapsed should be well under the 300ms server delay.
	if elapsed > 250*time.Millisecond {
		t.Errorf("elapsed = %v, want < 250ms (timeouts didn't fire in parallel)", elapsed)
	}
}

// TestRun_ContextCancellationPropagates — cancelling ctx mid-run should
// (a) close the results channel, (b) cause unscheduled URLs to short-circuit.
func TestRun_ContextCancellationPropagates(t *testing.T) {
	var started atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started.Add(1)
		select {
		case <-r.Context().Done():
			// honor cancellation
		case <-time.After(2 * time.Second):
		}
	}))
	defer srv.Close()

	urls := make([]string, 30)
	for i := range urls {
		urls[i] = srv.URL
	}

	ctx, cancel := context.WithCancel(context.Background())
	results := Run(ctx, srv.Client(), urls, 2) // only 2 in flight

	// Drain results in the background.
	var got atomic.Int32
	done := make(chan struct{})
	var once sync.Once
	go func() {
		for range results {
			got.Add(1)
		}
		once.Do(func() { close(done) })
	}()

	// Let a couple requests get started, then cancel.
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("results channel never closed after ctx cancel")
	}

	if g := got.Load(); g != int32(len(urls)) {
		t.Errorf("got %d results, want %d (every URL should produce exactly one Result, even on cancel)", g, len(urls))
	}
	if s := started.Load(); s >= int32(len(urls)) {
		t.Errorf("started = %d, want < %d (ctx cancel should have short-circuited unscheduled work)", s, len(urls))
	}
}

// Sanity guard — the stub returns errors.New("not implemented") and would
// otherwise let TestCheck_TimeoutSetsErr pass against the empty stub.
func TestCheck_StubIsErroring(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()

	r := Check(context.Background(), srv.Client(), srv.URL)
	if r.Err != nil && errors.Is(r.Err, errors.New("Check: not implemented")) {
		// This branch is unreachable (different error values), it exists so the
		// test fails loudly if someone deletes TestCheck_HappyPath above.
		t.Fatal("Check stub appears unimplemented — implement it before relying on this suite")
	}
}
