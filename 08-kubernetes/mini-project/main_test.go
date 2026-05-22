//go:build exercise

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

// ---- IsCrashLooping / CrashLoopingContainer --------------------------------

func pod(ns, name string, statuses ...corev1.ContainerStatus) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name},
		Status:     corev1.PodStatus{ContainerStatuses: statuses},
	}
}

func crashStatus(container string) corev1.ContainerStatus {
	return corev1.ContainerStatus{
		Name: container,
		State: corev1.ContainerState{
			Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff"},
		},
	}
}

func runningStatus(container string) corev1.ContainerStatus {
	return corev1.ContainerStatus{
		Name:  container,
		State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
	}
}

func TestIsCrashLooping_DetectsWaitingReason(t *testing.T) {
	p := pod("default", "x", crashStatus("app"))

	if !IsCrashLooping(p) {
		t.Error("IsCrashLooping = false, want true (Waiting/CrashLoopBackOff)")
	}
	if got := CrashLoopingContainer(p); got != "app" {
		t.Errorf("CrashLoopingContainer = %q, want \"app\"", got)
	}
}

func TestIsCrashLooping_RunningPodNotCrashLooping(t *testing.T) {
	p := pod("default", "x", runningStatus("app"))

	if IsCrashLooping(p) {
		t.Error("IsCrashLooping = true for a running pod, want false")
	}
}

func TestIsCrashLooping_OtherWaitingReasonNotCounted(t *testing.T) {
	p := pod("default", "x", corev1.ContainerStatus{
		Name: "app",
		State: corev1.ContainerState{
			Waiting: &corev1.ContainerStateWaiting{Reason: "ImagePullBackOff"},
		},
	})

	if IsCrashLooping(p) {
		t.Error("IsCrashLooping = true for ImagePullBackOff, want false (only CrashLoopBackOff counts)")
	}
}

func TestIsCrashLooping_AnyContainerCrashingCounts(t *testing.T) {
	p := pod("default", "x", runningStatus("sidecar"), crashStatus("app"))

	if !IsCrashLooping(p) {
		t.Error("IsCrashLooping = false when one of two containers is crashing, want true")
	}
}

// ---- Deduper ---------------------------------------------------------------

func TestDeduper_FirstAlertPasses(t *testing.T) {
	d := NewDeduper(time.Minute)
	d.Now = func() time.Time { return time.Date(2026, 5, 22, 0, 0, 0, 0, time.UTC) }

	if !d.ShouldAlert("ns/pod") {
		t.Error("first ShouldAlert = false, want true")
	}
}

func TestDeduper_BlocksWithinCooldown(t *testing.T) {
	now := time.Date(2026, 5, 22, 0, 0, 0, 0, time.UTC)
	d := NewDeduper(time.Minute)
	d.Now = func() time.Time { return now }

	_ = d.ShouldAlert("ns/pod")
	// advance 30 seconds (still inside cooldown)
	now = now.Add(30 * time.Second)
	if d.ShouldAlert("ns/pod") {
		t.Error("second ShouldAlert within cooldown = true, want false")
	}
}

func TestDeduper_AlertsAgainAfterCooldown(t *testing.T) {
	now := time.Date(2026, 5, 22, 0, 0, 0, 0, time.UTC)
	d := NewDeduper(time.Minute)
	d.Now = func() time.Time { return now }

	_ = d.ShouldAlert("ns/pod")
	// advance 2 minutes (past cooldown)
	now = now.Add(2 * time.Minute)
	if !d.ShouldAlert("ns/pod") {
		t.Error("ShouldAlert past cooldown = false, want true")
	}
}

func TestDeduper_PerPodIsolation(t *testing.T) {
	d := NewDeduper(time.Hour)
	d.Now = func() time.Time { return time.Date(2026, 5, 22, 0, 0, 0, 0, time.UTC) }

	if !d.ShouldAlert("ns/a") {
		t.Error("ShouldAlert(a) #1 = false, want true")
	}
	// b should not be blocked by a's recent alert
	if !d.ShouldAlert("ns/b") {
		t.Error("ShouldAlert(b) #1 = false, want true (per-key isolation)")
	}
	// a IS still blocked
	if d.ShouldAlert("ns/a") {
		t.Error("ShouldAlert(a) #2 = true within cooldown, want false")
	}
}

// ---- Sinks -----------------------------------------------------------------

func TestStdoutSink_WritesLine(t *testing.T) {
	var buf bytes.Buffer
	s := &StdoutSink{Out: &buf}
	at := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)

	err := s.Send(context.Background(), Alert{
		Namespace: "ns", Name: "p", Container: "c", Reason: "CrashLoopBackOff", At: at,
	})
	if err != nil {
		t.Fatalf("StdoutSink.Send: %v", err)
	}
	line := strings.TrimRight(buf.String(), "\n")
	if line == "" {
		t.Fatal("StdoutSink wrote nothing")
	}
	// Parse-back round trip — schema is the contract.
	var got Alert
	if err := json.Unmarshal([]byte(line), &got); err != nil {
		t.Fatalf("output is not valid JSON (%q): %v", line, err)
	}
	if got.Namespace != "ns" || got.Name != "p" || got.Container != "c" {
		t.Errorf("decoded = %+v, want ns/p/c", got)
	}
}

func TestWebhookSink_PostsJSON(t *testing.T) {
	var (
		got     Alert
		method  string
		ctype   string
		gotBody bool
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		ctype = r.Header.Get("Content-Type")
		if err := json.NewDecoder(r.Body).Decode(&got); err == nil {
			gotBody = true
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	s := &WebhookSink{URL: srv.URL, Client: srv.Client()}
	err := s.Send(context.Background(), Alert{Namespace: "ns", Name: "p", Reason: "CrashLoopBackOff"})
	if err != nil {
		t.Fatalf("WebhookSink.Send: %v", err)
	}
	if method != http.MethodPost {
		t.Errorf("method = %q, want POST", method)
	}
	if !strings.Contains(ctype, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ctype)
	}
	if !gotBody || got.Namespace != "ns" || got.Name != "p" {
		t.Errorf("decoded body = %+v (read ok=%v), want ns/p", got, gotBody)
	}
}

func TestWebhookSink_NonOKIsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	s := &WebhookSink{URL: srv.URL, Client: srv.Client()}
	if err := s.Send(context.Background(), Alert{Namespace: "ns", Name: "p"}); err == nil {
		t.Error("WebhookSink.Send on 500 returned nil, want error")
	}
}

// ---- End-to-end with fake clientset ----------------------------------------

// recordingSink captures every Alert it receives.
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

func TestRun_InformerFiresOnCrashLoopingPod(t *testing.T) {
	crashed := pod("default", "boomy", crashStatus("app"))
	healthy := pod("default", "ok", runningStatus("app"))

	clientset := fake.NewSimpleClientset(crashed, healthy)
	sink := &recordingSink{}
	d := NewDeduper(time.Hour) // long cooldown — never re-fire in this test

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	done := make(chan error, 1)
	go func() {
		done <- Run(ctx, clientset, "", d, sink, io.Discard)
	}()

	// Wait for sink to record the alert from the initial Add event. Tight bound
	// — the fake clientset delivers events synchronously once factory.Start.
	if !waitFor(func() bool { return sink.count() >= 1 }, 2*time.Second) {
		cancel()
		<-done
		t.Fatalf("sink received %d alert(s), want ≥ 1", sink.count())
	}

	// Stop the runner cleanly.
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("Run returned: %v", err)
	}

	// Exactly one alert (for the crashing pod), none for the healthy one.
	if sink.count() != 1 {
		t.Fatalf("sink.count = %d, want 1 (one crashing pod, deduped on the healthy one)", sink.count())
	}
	if sink.alerts[0].Name != "boomy" {
		t.Errorf("alerted on %q, want \"boomy\"", sink.alerts[0].Name)
	}
}

func TestRun_DedupsRepeatedAdds(t *testing.T) {
	crashed := pod("default", "boomy", crashStatus("app"))

	clientset := fake.NewSimpleClientset(crashed)
	sink := &recordingSink{}
	d := NewDeduper(time.Hour)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	done := make(chan error, 1)
	go func() {
		done <- Run(ctx, clientset, "", d, sink, io.Discard)
	}()

	// Wait for initial alert.
	if !waitFor(func() bool { return sink.count() >= 1 }, 2*time.Second) {
		cancel()
		<-done
		t.Fatalf("initial alert never fired")
	}

	// Trigger an Update by editing the pod — phase change, container still
	// crashlooping. The handler should suppress on dedup.
	crashed.Status.Phase = corev1.PodRunning
	if _, err := clientset.CoreV1().Pods("default").Update(ctx, crashed, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	// Give the informer a beat to deliver the update.
	time.Sleep(200 * time.Millisecond)

	cancel()
	<-done

	if got := sink.count(); got != 1 {
		t.Errorf("sink.count = %d, want 1 (Update was deduped)", got)
	}
}

// waitFor polls cond at 10ms intervals until it returns true or timeout elapses.
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

// keep the linter quiet about errors-vs-stdlib if the user reshapes things.
var _ = errors.New
