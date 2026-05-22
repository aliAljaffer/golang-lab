//go:build exercise

package ratelog

import (
	"testing"
	"time"
)

// newFake builds a Limiter whose clock starts at t0 and is advanced by tick.
func newFake(window time.Duration) (*Limiter, func(time.Duration)) {
	now := time.Unix(1_700_000_000, 0)
	l := New(window)
	l.Now = func() time.Time { return now }
	tick := func(d time.Duration) { now = now.Add(d) }
	return l, tick
}

func TestAllow_FirstCallIsAllowed(t *testing.T) {
	l, _ := newFake(time.Second)
	if !l.Allow("k") {
		t.Error("first Allow(k) = false, want true")
	}
}

func TestAllow_SecondCallWithinWindowIsBlocked(t *testing.T) {
	l, tick := newFake(time.Second)
	l.Allow("k")
	tick(500 * time.Millisecond)
	if l.Allow("k") {
		t.Error("second Allow within 500ms (window=1s) = true, want false")
	}
}

func TestAllow_AfterWindowAllowedAgain(t *testing.T) {
	l, tick := newFake(time.Second)
	l.Allow("k")
	tick(1100 * time.Millisecond)
	if !l.Allow("k") {
		t.Error("Allow after window = false, want true")
	}
}

func TestAllow_PerKeyIsolation(t *testing.T) {
	l, _ := newFake(time.Second)
	if !l.Allow("a") {
		t.Fatal("first Allow(a) should be true")
	}
	if !l.Allow("b") {
		t.Error("Allow(b) right after Allow(a) = false, want true (keys are independent)")
	}
}

func TestAllow_RateLimitedKeyStaysBlockedWhileOtherKeysFree(t *testing.T) {
	l, tick := newFake(time.Second)
	l.Allow("a")
	tick(100 * time.Millisecond)
	if l.Allow("a") {
		t.Error("Allow(a) within window = true, want false")
	}
	if !l.Allow("b") {
		t.Error("Allow(b) = false, want true (different key)")
	}
}
