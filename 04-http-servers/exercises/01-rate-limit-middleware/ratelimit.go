// 01-rate-limit-middleware — IP-based rate limiter as middleware.
//
// You implement:
//   - Limiter struct with Allow(key string) bool — pure logic, clock-injectable
//   - Middleware(l *Limiter) func(http.Handler) http.Handler — thin HTTP wrapper that keys by RemoteAddr
//
// Algorithm: fixed-window per key. Within a window of `Window`, a key is allowed
// up to `Limit` requests; the (Limit+1)th and beyond are rejected. When the
// window expires, the counter resets.
//
// Tests in ratelimit_test.go drive the design.
package ratelimit

import (
	"net/http"
	"sync"
	"time"
)

// Limiter is a fixed-window per-key counter. Safe for concurrent use.
type Limiter struct {
	Limit  int
	Window time.Duration
	// Now is the clock. Leave nil to use time.Now. Tests override it.
	Now func() time.Time

	mu      sync.Mutex
	buckets map[string]*bucket
}

type bucket struct {
	count       int
	windowStart time.Time
}

// New returns a Limiter ready to use.
func New(limit int, window time.Duration) *Limiter {
	return &Limiter{
		Limit:   limit,
		Window:  window,
		buckets: map[string]*bucket{},
	}
}

// Allow records a request for `key` and returns true iff it's within the limit.
// Concurrency-safe.
func (l *Limiter) Allow(key string) bool {
	// TODO: fixed-window check + record. Under l.mu (concurrent requests),
	//   look up the bucket for `key`, reset its counter when the window has
	//   elapsed, then admit-or-reject. The reset has to use the same `now`
	//   value you used to make the admit decision — otherwise two calls
	//   straddling a window boundary can race.
	return false
}

func (l *Limiter) now() time.Time {
	if l.Now != nil {
		return l.Now()
	}
	return time.Now()
}

// Middleware returns net/http middleware that 429s when Allow returns false.
// Keyed by r.RemoteAddr (host:port). For production you'd usually strip the port
// or honor X-Forwarded-For — kept simple here.
func Middleware(l *Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: key by r.RemoteAddr (port included is fine for this
			//   exercise), reject with 429 when Allow returns false, pass
			//   through otherwise.
			next.ServeHTTP(w, r)
		})
	}
}
