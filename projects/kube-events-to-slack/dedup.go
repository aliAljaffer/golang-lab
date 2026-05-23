package main

import (
	"sync"
	"time"
)

// Deduper rate-limits alerts per key. Use NewDeduper to build one. Safe for
// concurrent use — the informer handler can be invoked from multiple
// goroutines and ShouldAlert is the gate.
//
// Shape is intentionally identical to 08-kubernetes/mini-project/crashloop-alert
// — same clock-injection trick (Now func() time.Time) so tests stay
// deterministic.
type Deduper struct {
	mu       sync.Mutex
	cooldown time.Duration
	lastSent map[string]time.Time
	Now      func() time.Time // injected for tests; defaults to time.Now
}

// NewDeduper returns a Deduper with the given cooldown. Now defaults to
// time.Now.
func NewDeduper(cooldown time.Duration) *Deduper {
	return &Deduper{
		cooldown: cooldown,
		lastSent: map[string]time.Time{},
		Now:      time.Now,
	}
}

// ShouldAlert returns true iff no alert for key was sent within the cooldown.
// Side effect: on a true return, records the alert time so the next call
// within the cooldown returns false.
func (d *Deduper) ShouldAlert(key string) bool {
	// TODO: read+update lastSent under d.mu — the handler can be called from
	//   several goroutines. On a true return, you also have to RECORD now,
	//   so the next call within the cooldown is rejected. The test pins
	//   both halves: "first call true, second within cooldown false,
	//   call after cooldown true again".
	return false
}
