// Package ratelog implements a per-key rate limiter for noisy logs.
//
// The use case:
//
//	A worker hits "connection refused" 10,000 times a second when the
//	downstream is down. You want to know about it — but ONE log line per
//	second is enough. ratelog.Allow("conn-refused") returns true for the
//	first call within a window, false for the rest, true again after the
//	window expires.
//
// Surface:
//
//	type Limiter struct { ... }
//	func New(window time.Duration) *Limiter
//	func (l *Limiter) Allow(key string) bool
//
// Idiomatic usage:
//
//	limiter := ratelog.New(1 * time.Second)
//	for err := range errors {
//	    if limiter.Allow("conn-refused") {
//	        logger.Error("downstream connection refused", slog.Any("err", err))
//	    }
//	    // ... handle err ...
//	}
//
// The clock is injectable via Now so tests can advance time without sleeping.
package ratelog

import (
	"sync"
	"time"
)

// Limiter rate-limits per key.
//
// Concurrency: safe for use from multiple goroutines.
//
// Memory: keys are never evicted in this implementation. For long-running
// processes with unbounded key churn, swap in an LRU. For the typical case
// of a fixed set of error keys, this is fine.
type Limiter struct {
	Window time.Duration
	Now    func() time.Time // defaults to time.Now in New

	mu       sync.Mutex
	lastSeen map[string]time.Time // pre-declared so go vet -tags=exercise is happy
}

// New returns a Limiter with the given window. The Now func defaults to
// time.Now; replace it in tests to drive a fake clock.
func New(window time.Duration) *Limiter {
	return &Limiter{
		Window:   window,
		Now:      time.Now,
		lastSeen: make(map[string]time.Time),
	}
}

// Allow reports whether the caller should log on this call. Returns true:
//
//   - on the first call for this key, OR
//   - if the time since the last allowed call for this key exceeds Window.
//
// When Allow returns true, it stamps the current time for this key.
func (l *Limiter) Allow(key string) bool {
	// TODO: read+update lastSeen under l.mu. On the allowed path, RECORD
	//   now — otherwise the limiter never engages. The "first call always
	//   allows" case is the unseen-key branch; don't special-case it,
	//   just treat a missing entry the same as "old enough".
	_ = key
	return false
}
