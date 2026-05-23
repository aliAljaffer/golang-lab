package main

import (
	"errors"
	"time"

	corev1 "k8s.io/api/core/v1"
)

// Alert is the payload one detection emits. The Sink decides how to serialize
// it (JSON-per-line for stdout; JSON POST body for the Slack webhook).
//
// The field set is the public contract: tests pin this shape.
type Alert struct {
	Text      string    `json:"text"`      // human-readable summary
	Namespace string    `json:"namespace"` // from event.InvolvedObject.Namespace
	Kind      string    `json:"kind"`     // e.g. "Pod"
	Name      string    `json:"name"`     // e.g. "boomy-7c4f"
	Reason    string    `json:"reason"`    // e.g. "BackOff"
	Severity  string    `json:"severity"`  // "Normal" or "Warning"
	Count     int32     `json:"count"`     // event.Count (number of repetitions)
	Age       string    `json:"age"`       // human-readable, computed from now - lastTimestamp
	At        time.Time `json:"at"`        // the last-seen timestamp
}

// FormatSlackMessage converts a corev1.Event into the Alert shape the sinks
// emit. now is injected (rather than time.Now()) so tests can pin the Age
// string deterministically.
//
// Required fields on the Alert:
//   - Text: a one-line human summary. Suggested form:
//     "[Warning] default/boomy-7c4f: BackOff (x5, 2m ago)"
//   - Namespace: e.event.InvolvedObject.Namespace
//   - Kind:      event.InvolvedObject.Kind
//   - Name:      event.InvolvedObject.Name
//   - Reason:    event.Reason
//   - Severity:  event.Type  ("Normal" / "Warning")
//   - Count:     event.Count
//   - Age:       human-readable duration since the event's last-seen time,
//                using EventAt(event) as the timestamp source.
//   - At:        the same timestamp returned by EventAt(event).
func FormatSlackMessage(e *corev1.Event, now time.Time) Alert {
	// TODO: populate every field per the spec above. The Text format is
	//   yours, but the test checks that it includes severity, namespace/name,
	//   reason, count, and age — so don't drop any of those. Age comes from
	//   now - EventAt(e), which is why `now` is injected (so a frozen clock
	//   makes the assertion deterministic).
	return Alert{}
}

// EventAt returns the most-recently-known timestamp for the event. Kubernetes
// stores three different timestamps on a corev1.Event depending on the source:
//   - EventTime (the new events.k8s.io path)
//   - LastTimestamp (the legacy core/v1 path; set on repeats)
//   - FirstTimestamp (the legacy core/v1 path; set on first occurrence)
//
// Pick them in that priority order; whichever is non-zero wins.
func EventAt(e *corev1.Event) time.Time {
	// TODO: pick the first non-zero of EventTime / LastTimestamp / FirstTimestamp.
	//   The priority order matters — newer kubelets populate EventTime; legacy
	//   sources only set Last/First. Skipping EventTime would lose seconds on
	//   modern clusters; falling through to FirstTimestamp is the worst-case
	//   sane default.
	return time.Time{}
}

// DedupKey returns the per-event dedup key. Two distinct events should share
// a key iff a human would consider them "the same alert repeating".
//
// Suggested: "<namespace>/<involvedObject.Kind>/<involvedObject.Name>:<reason>"
//
// Tests pin this format — keep it stable.
func DedupKey(e *corev1.Event) string {
	// TODO: build the key from InvolvedObject's namespace / kind / name plus
	//   the reason. Format pinned by the test — keep it stable across changes.
	return ""
}

// silence unused-import noise during scaffolding
var _ = errors.New
