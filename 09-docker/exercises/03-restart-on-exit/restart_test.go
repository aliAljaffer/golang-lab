//go:build exercise

package restart

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
)

// ---- ShouldRestart ---------------------------------------------------------

func dieMsg(id, exitCode string) events.Message {
	return events.Message{
		Type:   "container",
		Action: "die",
		Actor: events.Actor{
			ID:         id,
			Attributes: map[string]string{"exitCode": exitCode, "name": "victim"},
		},
	}
}

func TestShouldRestart_NonZeroExitTriggers(t *testing.T) {
	if !ShouldRestart(dieMsg("abc", "137")) {
		t.Error("ShouldRestart on exit=137 = false, want true")
	}
}

func TestShouldRestart_ZeroExitIgnored(t *testing.T) {
	// A clean `docker stop <c>` produces a die event with exitCode=0.
	if ShouldRestart(dieMsg("abc", "0")) {
		t.Error("ShouldRestart on exit=0 = true, want false (clean shutdown shouldn't restart)")
	}
}

func TestShouldRestart_NonContainerEventIgnored(t *testing.T) {
	m := events.Message{
		Type:   "image",
		Action: "pull",
	}
	if ShouldRestart(m) {
		t.Error("ShouldRestart on image event = true, want false")
	}
}

func TestShouldRestart_NonDieActionIgnored(t *testing.T) {
	m := events.Message{
		Type:   "container",
		Action: "start",
		Actor:  events.Actor{Attributes: map[string]string{"exitCode": "1"}},
	}
	if ShouldRestart(m) {
		t.Error("ShouldRestart on start action = true, want false")
	}
}

func TestShouldRestart_MissingExitCodeIgnored(t *testing.T) {
	m := events.Message{Type: "container", Action: "die", Actor: events.Actor{Attributes: map[string]string{}}}
	if ShouldRestart(m) {
		t.Error("ShouldRestart with no exitCode attr = true, want false (be conservative)")
	}
}

// ---- Run -------------------------------------------------------------------

// fakeAPI is a channel-backed stand-in for *client.Client. The test feeds
// events.Message into msgCh; Run consumes and calls ContainerStart.
type fakeAPI struct {
	msgCh    chan events.Message
	errCh    chan error
	starts   atomic.Int32
	startMu  sync.Mutex
	startIDs []string
	startErr error
}

func newFakeAPI() *fakeAPI {
	return &fakeAPI{msgCh: make(chan events.Message, 8), errCh: make(chan error, 1)}
}

func (f *fakeAPI) Events(_ context.Context, _ events.ListOptions) (<-chan events.Message, <-chan error) {
	return f.msgCh, f.errCh
}

func (f *fakeAPI) ContainerStart(_ context.Context, id string, _ container.StartOptions) error {
	f.starts.Add(1)
	f.startMu.Lock()
	f.startIDs = append(f.startIDs, id)
	f.startMu.Unlock()
	return f.startErr
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

func TestRun_RestartsOnNonZeroExit(t *testing.T) {
	api := newFakeAPI()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	done := make(chan error, 1)
	go func() { done <- Run(ctx, api) }()

	// Push a die-with-non-zero event.
	api.msgCh <- dieMsg("crashy", "1")

	if !waitFor(func() bool { return api.starts.Load() >= 1 }, 1*time.Second) {
		cancel()
		<-done
		t.Fatalf("ContainerStart was not called within 1s")
	}

	cancel()
	if err := <-done; err != nil {
		t.Fatalf("Run returned: %v", err)
	}

	api.startMu.Lock()
	defer api.startMu.Unlock()
	if len(api.startIDs) != 1 || api.startIDs[0] != "crashy" {
		t.Errorf("startIDs = %v, want [crashy]", api.startIDs)
	}
}

func TestRun_DoesNotRestartCleanExit(t *testing.T) {
	api := newFakeAPI()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	done := make(chan error, 1)
	go func() { done <- Run(ctx, api) }()

	api.msgCh <- dieMsg("ok-container", "0")

	// Give the loop a beat; we expect ZERO ContainerStart calls.
	time.Sleep(100 * time.Millisecond)

	cancel()
	<-done

	if got := api.starts.Load(); got != 0 {
		t.Errorf("ContainerStart called %d times, want 0", got)
	}
}

func TestRun_ContinuesAfterStartError(t *testing.T) {
	api := newFakeAPI()
	api.startErr = errors.New("daemon flaky")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	done := make(chan error, 1)
	go func() { done <- Run(ctx, api) }()

	api.msgCh <- dieMsg("a", "1")
	api.msgCh <- dieMsg("b", "1")

	if !waitFor(func() bool { return api.starts.Load() >= 2 }, 1*time.Second) {
		cancel()
		<-done
		t.Fatalf("starts = %d, want 2 (Run must not abort on a failed ContainerStart)", api.starts.Load())
	}

	cancel()
	<-done
}

func TestRun_TransportErrorReturned(t *testing.T) {
	api := newFakeAPI()
	stub := errors.New("daemon went away")
	api.errCh <- stub

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	done := make(chan error, 1)
	go func() { done <- Run(ctx, api) }()

	select {
	case err := <-done:
		if !errors.Is(err, stub) {
			t.Errorf("Run = %v, want %v (transport errors propagate)", err, stub)
		}
	case <-time.After(1 * time.Second):
		cancel()
		t.Fatal("Run did not return after errCh delivered an error")
	}
}

func TestRun_CtxCancelReturnsNil(t *testing.T) {
	api := newFakeAPI()
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() { done <- Run(ctx, api) }()

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run on ctx-cancel = %v, want nil (clean shutdown)", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Run did not return after ctx cancel")
	}
}
