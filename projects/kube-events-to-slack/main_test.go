//go:build exercise

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// ---- helpers ---------------------------------------------------------------

// fixedNow at 2026-05-22T12:00:00Z — every age-sensitive test uses this as
// "now" so durations are exact.
var fixedNow = time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)

func nowFn() time.Time { return fixedNow }

// event builds a corev1.Event with sane defaults plus the given overrides.
// Reason / Type / InvolvedObject / Count / LastTimestamp are the
// load-bearing fields.
func event(ns, name, kind, reason, severity string, count int32, lastSeen time.Time) *corev1.Event {
	return &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name + ".event",
		},
		InvolvedObject: corev1.ObjectReference{
			Namespace: ns,
			Kind:      kind,
			Name:      name,
		},
		Reason:        reason,
		Message:       reason + " happened",
		Type:          severity,
		Count:         count,
		LastTimestamp: metav1.NewTime(lastSeen),
	}
}

// ---- Filter ---------------------------------------------------------------

func TestFilter_SeverityMatch(t *testing.T) {
	f := Filter{
		Severities: map[string]bool{"Warning": true},
		Now:        nowFn,
	}
	warn := event("default", "p", "Pod", "BackOff", "Warning", 1, fixedNow)
	normal := event("default", "p", "Pod", "Pulled", "Normal", 1, fixedNow)

	if !f.ShouldAlert(warn) {
		t.Error("ShouldAlert(Warning) = false, want true")
	}
	if f.ShouldAlert(normal) {
		t.Error("ShouldAlert(Normal) = true, want false (severity allow-list)")
	}
}

func TestFilter_NamespaceAllowList(t *testing.T) {
	f := Filter{
		Namespaces: map[string]bool{"kube-system": true},
		Now:        nowFn,
	}
	good := event("kube-system", "p", "Pod", "BackOff", "Warning", 1, fixedNow)
	bad := event("default", "p", "Pod", "BackOff", "Warning", 1, fixedNow)

	if !f.ShouldAlert(good) {
		t.Error("ShouldAlert(kube-system) = false, want true")
	}
	if f.ShouldAlert(bad) {
		t.Error("ShouldAlert(default) = true, want false (namespace allow-list)")
	}
}

func TestFilter_AgeCutoff(t *testing.T) {
	f := Filter{
		MaxAge: 5 * time.Minute,
		Now:    nowFn,
	}
	fresh := event("default", "p", "Pod", "BackOff", "Warning", 1, fixedNow.Add(-2*time.Minute))
	stale := event("default", "p", "Pod", "BackOff", "Warning", 1, fixedNow.Add(-10*time.Minute))

	if !f.ShouldAlert(fresh) {
		t.Error("ShouldAlert(fresh, age=2m, max=5m) = false, want true")
	}
	if f.ShouldAlert(stale) {
		t.Error("ShouldAlert(stale, age=10m, max=5m) = true, want false")
	}
}

func TestFilter_EmptyMeansPassAll(t *testing.T) {
	f := Filter{Now: nowFn} // no severities, no namespaces, MaxAge=0
	e := event("anywhere", "p", "Pod", "Anything", "Normal", 1, fixedNow.Add(-365*24*time.Hour))

	if !f.ShouldAlert(e) {
		t.Error("zero-value Filter rejected an event, want pass-all")
	}
}

// ---- Deduper ---------------------------------------------------------------

func TestDeduper_FirstAlertPasses(t *testing.T) {
	d := NewDeduper(time.Minute)
	d.Now = nowFn

	if !d.ShouldAlert("k") {
		t.Error("first ShouldAlert = false, want true")
	}
}

func TestDeduper_BlocksWithinCooldown(t *testing.T) {
	now := fixedNow
	d := NewDeduper(time.Minute)
	d.Now = func() time.Time { return now }

	_ = d.ShouldAlert("k")
	now = now.Add(30 * time.Second) // still in cooldown
	if d.ShouldAlert("k") {
		t.Error("ShouldAlert inside cooldown = true, want false")
	}
}

func TestDeduper_AlertsAgainAfterCooldown(t *testing.T) {
	now := fixedNow
	d := NewDeduper(time.Minute)
	d.Now = func() time.Time { return now }

	_ = d.ShouldAlert("k")
	now = now.Add(2 * time.Minute) // past cooldown
	if !d.ShouldAlert("k") {
		t.Error("ShouldAlert past cooldown = false, want true")
	}
}

func TestDeduper_PerKeyIsolation(t *testing.T) {
	d := NewDeduper(time.Hour)
	d.Now = nowFn

	if !d.ShouldAlert("a") {
		t.Error("ShouldAlert(a) #1 = false, want true")
	}
	if !d.ShouldAlert("b") {
		t.Error("ShouldAlert(b) #1 = false, want true (per-key isolation)")
	}
	if d.ShouldAlert("a") {
		t.Error("ShouldAlert(a) #2 = true within cooldown, want false")
	}
}

// ---- Sinks -----------------------------------------------------------------

func TestStdoutSink_WritesLine(t *testing.T) {
	var buf bytes.Buffer
	s := &StdoutSink{Out: &buf}

	err := s.Send(context.Background(), Alert{
		Namespace: "ns", Name: "p", Kind: "Pod", Reason: "BackOff",
		Severity: "Warning", Count: 3, Age: "2m0s", At: fixedNow,
	})
	if err != nil {
		t.Fatalf("StdoutSink.Send: %v", err)
	}
	line := strings.TrimRight(buf.String(), "\n")
	if line == "" {
		t.Fatal("StdoutSink wrote nothing")
	}
	var got Alert
	if err := json.Unmarshal([]byte(line), &got); err != nil {
		t.Fatalf("output is not valid JSON (%q): %v", line, err)
	}
	if got.Namespace != "ns" || got.Name != "p" || got.Reason != "BackOff" {
		t.Errorf("decoded = %+v, want ns=ns name=p reason=BackOff", got)
	}
}

func TestWebhookSink_PostsJSON(t *testing.T) {
	var (
		gotBody    Alert
		gotMethod  string
		gotCT      string
		decoded    bool
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotCT = r.Header.Get("Content-Type")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err == nil {
			decoded = true
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s := &WebhookSink{URL: srv.URL, Client: srv.Client()}
	err := s.Send(context.Background(), Alert{Namespace: "ns", Name: "p", Reason: "BackOff", Severity: "Warning"})
	if err != nil {
		t.Fatalf("WebhookSink.Send: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if !strings.Contains(gotCT, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", gotCT)
	}
	if !decoded || gotBody.Namespace != "ns" || gotBody.Name != "p" {
		t.Errorf("decoded body = %+v (decoded=%v), want ns=ns name=p", gotBody, decoded)
	}
}

func TestWebhookSink_NonOKIsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	s := &WebhookSink{URL: srv.URL, Client: srv.Client(), MaxRetries: 0}
	if err := s.Send(context.Background(), Alert{Namespace: "ns"}); err == nil {
		t.Error("WebhookSink.Send on 500 with MaxRetries=0 returned nil, want error")
	}
}

// ---- FormatSlackMessage ---------------------------------------------------

func TestFormatSlackMessage_Shape(t *testing.T) {
	lastSeen := fixedNow.Add(-2 * time.Minute)
	e := event("default", "boomy", "Pod", "BackOff", "Warning", 5, lastSeen)

	a := FormatSlackMessage(e, fixedNow)

	if a.Namespace != "default" {
		t.Errorf("Namespace = %q, want default", a.Namespace)
	}
	if a.Kind != "Pod" {
		t.Errorf("Kind = %q, want Pod", a.Kind)
	}
	if a.Name != "boomy" {
		t.Errorf("Name = %q, want boomy", a.Name)
	}
	if a.Reason != "BackOff" {
		t.Errorf("Reason = %q, want BackOff", a.Reason)
	}
	if a.Severity != "Warning" {
		t.Errorf("Severity = %q, want Warning", a.Severity)
	}
	if a.Count != 5 {
		t.Errorf("Count = %d, want 5", a.Count)
	}
	if !a.At.Equal(lastSeen) {
		t.Errorf("At = %v, want %v (event LastTimestamp)", a.At, lastSeen)
	}
	if a.Age == "" {
		t.Error("Age = \"\", want a non-empty duration string like \"2m0s\"")
	}
	// Text must include the load-bearing fields so a Slack reader can grok
	// the alert without unfolding the JSON.
	for _, want := range []string{"Warning", "default", "boomy", "BackOff"} {
		if !strings.Contains(a.Text, want) {
			t.Errorf("Text = %q, want it to contain %q", a.Text, want)
		}
	}
}

func TestDedupKey_PinsFormat(t *testing.T) {
	e := event("default", "boomy", "Pod", "BackOff", "Warning", 1, fixedNow)
	got := DedupKey(e)
	want := "default/Pod/boomy:BackOff"
	if got != want {
		t.Errorf("DedupKey = %q, want %q", got, want)
	}
}

// ---- End-to-end with fake clientset ---------------------------------------

type recordingSink struct {
	mu     sync.Mutex
	alerts []Alert
}

func (r *recordingSink) Send(_ context.Context, a Alert) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.alerts = append(r.alerts, a)
	return nil
}

func (r *recordingSink) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.alerts)
}

func (r *recordingSink) names() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]string, len(r.alerts))
	for i, a := range r.alerts {
		out[i] = a.Name
	}
	return out
}

// TestRun_WarningFires_NormalIsSilent — the load-bearing end-to-end test.
// A Warning event must fire exactly one alert; a Normal event must be silent
// when the filter is Severities={Warning}.
func TestRun_WarningFires_NormalIsSilent(t *testing.T) {
	warn := event("default", "boomy", "Pod", "BackOff", "Warning", 1, fixedNow.Add(-30*time.Second))
	normal := event("default", "ok", "Pod", "Pulled", "Normal", 1, fixedNow.Add(-30*time.Second))

	clientset := fake.NewSimpleClientset(warn, normal)
	sink := &recordingSink{}
	deduper := NewDeduper(time.Hour) // never re-fire in this test
	filter := Filter{
		Severities: map[string]bool{"Warning": true},
		MaxAge:     time.Hour,
		Now:        time.Now, // production Now — the events use real-ish recent timestamps
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	done := make(chan error, 1)
	go func() {
		done <- Run(ctx, clientset, filter, deduper, sink, io.Discard)
	}()

	if !waitFor(func() bool { return sink.count() >= 1 }, 2*time.Second) {
		cancel()
		<-done
		t.Fatalf("sink received %d alert(s), want ≥ 1", sink.count())
	}

	cancel()
	if err := <-done; err != nil {
		t.Fatalf("Run returned: %v", err)
	}

	if sink.count() != 1 {
		t.Fatalf("sink.count = %d (%v), want 1 (Warning only)", sink.count(), sink.names())
	}
	if sink.alerts[0].Name != "boomy" {
		t.Errorf("alerted on %q, want \"boomy\"", sink.alerts[0].Name)
	}
}

func waitFor(cond func() bool, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return cond()
}
