//go:build exercise

package scheduler

import (
	"testing"
	"time"
)

// These tests use a fake clock so they're deterministic AND fast.
// They will fail until you:
//   1. Add a `Now func() time.Time` field to Scheduler.
//   2. Default it to time.Now in New().
//   3. Replace the two time.Now() calls in ShouldFire/Fire with s.Now().

func TestShouldFire_FirstCallAlwaysFires(t *testing.T) {
	s := New()
	s.Now = func() time.Time { return time.Unix(1_700_000_000, 0) }

	if !s.ShouldFire("backup", time.Hour) {
		t.Fatal("first-ever ShouldFire should return true")
	}
}

func TestShouldFire_WithinCooldownReturnsFalse(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	s := New()
	s.Now = func() time.Time { return now }

	s.Fire("backup")
	// Advance the fake clock by 10 minutes — still inside the 1h cooldown.
	now = now.Add(10 * time.Minute)

	if s.ShouldFire("backup", time.Hour) {
		t.Error("ShouldFire returned true within cooldown")
	}
}

func TestShouldFire_AfterCooldownReturnsTrue(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	s := New()
	s.Now = func() time.Time { return now }

	s.Fire("backup")
	now = now.Add(2 * time.Hour) // well past the 1h cooldown

	if !s.ShouldFire("backup", time.Hour) {
		t.Error("ShouldFire returned false past cooldown")
	}
}

func TestShouldFire_PerNameIsolation(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	s := New()
	s.Now = func() time.Time { return now }

	s.Fire("backup")
	// "rotate" has never fired — should still be eligible.
	if !s.ShouldFire("rotate", time.Hour) {
		t.Error("ShouldFire(rotate) returned false, but rotate has never fired")
	}
}

func TestShouldFire_DefaultsToRealClock(t *testing.T) {
	// New() should populate Now with time.Now by default — so a fresh
	// Scheduler still works in production without anyone touching Now.
	s := New()
	if s.Now == nil {
		t.Fatal("New() did not initialize Now — production callers would panic on time.Now()")
	}
	if !s.ShouldFire("first", time.Nanosecond) {
		t.Error("first-call ShouldFire should be true even with the real clock")
	}
}
