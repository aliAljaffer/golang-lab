//go:build exercise

package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAllow_UnderLimit(t *testing.T) {
	l := New(3, time.Second)
	for i := 0; i < 3; i++ {
		if !l.Allow("ip-a") {
			t.Errorf("call %d: Allow=false, want true (limit=3)", i+1)
		}
	}
}

func TestAllow_OverLimitRejected(t *testing.T) {
	l := New(2, time.Second)
	if !l.Allow("ip-a") {
		t.Fatal("first call rejected")
	}
	if !l.Allow("ip-a") {
		t.Fatal("second call rejected")
	}
	if l.Allow("ip-a") {
		t.Error("third call: Allow=true, want false (over limit)")
	}
	if l.Allow("ip-a") {
		t.Error("fourth call: Allow=true, want false (still over limit)")
	}
}

func TestAllow_ResetsAfterWindow(t *testing.T) {
	now := time.Unix(0, 0)
	l := New(1, time.Second)
	l.Now = func() time.Time { return now }

	if !l.Allow("ip-a") {
		t.Fatal("first call rejected")
	}
	if l.Allow("ip-a") {
		t.Fatal("second call within window: want rejected")
	}

	// Advance time past the window — counter resets.
	now = now.Add(1500 * time.Millisecond)
	if !l.Allow("ip-a") {
		t.Error("after window: Allow=false, want true (counter should reset)")
	}
}

func TestAllow_IsolatesPerKey(t *testing.T) {
	l := New(1, time.Second)
	if !l.Allow("ip-a") {
		t.Fatal("ip-a first call rejected")
	}
	if l.Allow("ip-a") {
		t.Fatal("ip-a second call: want rejected")
	}
	if !l.Allow("ip-b") {
		t.Error("ip-b first call: rejected — limiter should isolate per key")
	}
}

func TestMiddleware_429sOverLimit(t *testing.T) {
	l := New(1, time.Second)
	h := Middleware(l)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))

	srv := httptest.NewServer(h)
	defer srv.Close()

	resp1, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatalf("first GET: %v", err)
	}
	resp1.Body.Close()
	if resp1.StatusCode != 200 {
		t.Errorf("first GET status = %d, want 200", resp1.StatusCode)
	}

	resp2, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatalf("second GET: %v", err)
	}
	resp2.Body.Close()
	if resp2.StatusCode != http.StatusTooManyRequests {
		t.Errorf("second GET status = %d, want 429", resp2.StatusCode)
	}
}
