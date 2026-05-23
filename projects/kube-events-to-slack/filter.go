package main

import (
	"time"

	corev1 "k8s.io/api/core/v1"
)

// Filter decides whether a given event is interesting enough to forward to
// the dedup+sink pipeline. Zero-value semantics: an empty Filter passes
// every event (no severity restriction, no namespace restriction, no age
// cutoff). This keeps tests and CLI defaults honest.
type Filter struct {
	// Severities is the set of event.Type values to accept ("Normal", "Warning").
	// nil or empty == accept all severities.
	Severities map[string]bool

	// Namespaces is the namespace allow-list.
	// nil or empty == accept all namespaces.
	Namespaces map[string]bool

	// MaxAge skips events whose EventAt() timestamp is older than MaxAge
	// relative to Now(). Zero means "no cutoff".
	MaxAge time.Duration

	// Now is injected so tests can pin the age comparison. nil == time.Now.
	Now func() time.Time
}

// ShouldAlert returns true iff the event passes every active filter.
//
// Order is whatever's cheapest first; the tests don't care about the order,
// only about the truth table:
//
//   - if Severities is non-empty and event.Type ∉ Severities -> false
//   - if Namespaces is non-empty and event.InvolvedObject.Namespace ∉ Namespaces -> false
//   - if MaxAge > 0 and (Now() - EventAt(event)) > MaxAge -> false
//   - otherwise -> true
func (f Filter) ShouldAlert(e *corev1.Event) bool {
	// TODO: guard against nil event.
	// TODO: enforce the severity allow-list (skip if Severities is empty).
	// TODO: enforce the namespace allow-list (skip if Namespaces is empty).
	// TODO: enforce MaxAge using EventAt(e) and (f.Now or time.Now)().
	// TODO: return true if everything passes.
	return false
}
