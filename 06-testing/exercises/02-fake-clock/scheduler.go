// Package scheduler decides whether a scheduled task should fire.
//
// The current implementation calls time.Now() directly, which makes it
// impossible to test deterministically without sleeping. Your job is to
// refactor so a fake clock can be injected.
package scheduler

import "time"

// Scheduler tracks when each task is allowed to fire.
//
// The Now field is the clock seam — tests overwrite it with a fake function
// returning a fixed time; production callers leave it pointing at time.Now.
// Storing `func() time.Time` is the lightest-weight clock seam in Go. You
// don't need a Clock interface unless you also need to fake other time-
// related methods (Sleep, After, Tick).
type Scheduler struct {
	Now       func() time.Time
	lastFired map[string]time.Time
}

// New returns a Scheduler ready to use.
func New() *Scheduler {
	return &Scheduler{
		Now:       time.Now,
		lastFired: map[string]time.Time{},
	}
}

// ShouldFire reports whether task `name` is eligible to fire, given that it
// should fire at most once per `cooldown`.
//
// PROBLEM: this calls time.Now() directly. Tests of the cooldown logic have
// to either sleep (slow + flaky) or live with non-determinism. The exercise
// is to replace the two time.Now() calls below with s.Now().
func (s *Scheduler) ShouldFire(name string, cooldown time.Duration) bool {
	now := time.Now() // TODO: replace with s.Now()
	last, ok := s.lastFired[name]
	if !ok {
		return true
	}
	return now.Sub(last) >= cooldown
}

// Fire records that `name` just fired. The next ShouldFire for the same name
// will return false until `cooldown` has elapsed.
func (s *Scheduler) Fire(name string) {
	s.lastFired[name] = time.Now() // TODO: replace with s.Now()
}
