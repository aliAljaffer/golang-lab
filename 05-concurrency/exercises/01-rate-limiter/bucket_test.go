//go:build exercise

package bucket

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestAllow_BucketInitiallyFull(t *testing.T) {
	b := New(3, time.Hour) // refill so slow it can't refill mid-test
	defer b.Stop()

	for i := 0; i < 3; i++ {
		if !b.Allow() {
			t.Errorf("call %d: Allow=false, want true (bucket pre-filled to 3)", i+1)
		}
	}
	if b.Allow() {
		t.Error("call 4: Allow=true, want false (bucket should be empty)")
	}
}

func TestAllow_RefillsOverTime(t *testing.T) {
	b := New(1, 30*time.Millisecond)
	defer b.Stop()

	if !b.Allow() {
		t.Fatal("first Allow: rejected")
	}
	if b.Allow() {
		t.Fatal("second Allow immediately: should be empty")
	}

	// Wait long enough for at least one refill tick.
	time.Sleep(80 * time.Millisecond)

	if !b.Allow() {
		t.Error("after refill window: Allow=false, want true")
	}
}

func TestWait_ReturnsWhenTokenAvailable(t *testing.T) {
	b := New(1, time.Hour)
	defer b.Stop()

	// Consume the one initial token.
	if !b.Allow() {
		t.Fatal("setup: first Allow rejected")
	}

	// Wait should block until we manually inject a token by stopping & starting,
	// or until a refill arrives. Here we use a fresh bucket and call Wait
	// immediately — first token is already present.
	b2 := New(1, time.Hour)
	defer b2.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if err := b2.Wait(ctx); err != nil {
		t.Errorf("Wait err = %v, want nil (token was already in the bucket)", err)
	}
}

func TestWait_HonorsContextCancel(t *testing.T) {
	b := New(1, time.Hour)
	defer b.Stop()

	// Drain.
	if !b.Allow() {
		t.Fatal("setup: first Allow rejected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := b.Wait(ctx)
	if err == nil {
		t.Fatal("Wait err = nil, want ctx.DeadlineExceeded")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Wait err = %v, want context.DeadlineExceeded", err)
	}
	if elapsed := time.Since(start); elapsed > 200*time.Millisecond {
		t.Errorf("Wait took %v, expected ~30ms (ctx wasn't honored)", elapsed)
	}
}

func TestWait_ReturnsErrStoppedOnStop(t *testing.T) {
	b := New(1, time.Hour)
	if !b.Allow() {
		t.Fatal("setup: first Allow rejected")
	}

	var waitErr atomic.Value
	done := make(chan struct{})
	go func() {
		waitErr.Store(b.Wait(context.Background()))
		close(done)
	}()

	time.Sleep(20 * time.Millisecond)
	b.Stop()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Wait did not return after Stop")
	}

	got := waitErr.Load().(error)
	if !errors.Is(got, ErrStopped) {
		t.Errorf("Wait err = %v, want ErrStopped", got)
	}
}
