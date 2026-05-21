//go:build exercise

package broker

import (
	"sync"
	"testing"
	"time"
)

// drainWithin reads up to `want` messages from ch, blocking up to `d` total.
// Returns the messages it managed to collect.
func drainWithin(ch <-chan string, want int, d time.Duration) []string {
	got := make([]string, 0, want)
	deadline := time.After(d)
	for len(got) < want {
		select {
		case msg, ok := <-ch:
			if !ok {
				return got // channel closed
			}
			got = append(got, msg)
		case <-deadline:
			return got
		}
	}
	return got
}

func TestPublish_FansOutToAllSubscribers(t *testing.T) {
	b := New()
	defer b.Close()

	s1 := b.Subscribe()
	s2 := b.Subscribe()
	s3 := b.Subscribe()
	if s1 == nil || s2 == nil || s3 == nil {
		t.Fatal("Subscribe returned nil")
	}

	b.Publish("hello")

	for i, ch := range []<-chan string{s1, s2, s3} {
		got := drainWithin(ch, 1, 200*time.Millisecond)
		if len(got) != 1 || got[0] != "hello" {
			t.Errorf("sub %d: got %v, want [hello]", i+1, got)
		}
	}
}

func TestPublish_PreservesOrderPerSubscriber(t *testing.T) {
	b := New()
	defer b.Close()

	s := b.Subscribe()
	if s == nil {
		t.Fatal("Subscribe returned nil")
	}

	msgs := []string{"a", "b", "c", "d"}
	for _, m := range msgs {
		b.Publish(m)
	}

	got := drainWithin(s, len(msgs), 200*time.Millisecond)
	if len(got) != len(msgs) {
		t.Fatalf("got %v, want %v", got, msgs)
	}
	for i := range msgs {
		if got[i] != msgs[i] {
			t.Errorf("idx %d: got %q, want %q (order should be preserved)", i, got[i], msgs[i])
		}
	}
}

func TestUnsubscribe_StopsDelivery(t *testing.T) {
	b := New()
	defer b.Close()

	keep := b.Subscribe()
	drop := b.Subscribe()
	if keep == nil || drop == nil {
		t.Fatal("Subscribe returned nil")
	}

	b.Unsubscribe(drop)
	b.Publish("after-unsub")

	// `keep` should receive.
	got := drainWithin(keep, 1, 200*time.Millisecond)
	if len(got) != 1 || got[0] != "after-unsub" {
		t.Errorf("keep: got %v, want [after-unsub]", got)
	}

	// `drop` should have been closed by Unsubscribe — a receive returns ok=false.
	select {
	case _, ok := <-drop:
		if ok {
			t.Error("drop: received a message after Unsubscribe, channel should be closed")
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("drop: channel was not closed by Unsubscribe (receive blocked forever)")
	}
}

func TestClose_ClosesAllSubscribers(t *testing.T) {
	b := New()

	subs := []<-chan string{b.Subscribe(), b.Subscribe(), b.Subscribe()}
	for i, s := range subs {
		if s == nil {
			t.Fatalf("Subscribe %d returned nil", i)
		}
	}

	b.Close()

	for i, s := range subs {
		select {
		case _, ok := <-s:
			if ok {
				t.Errorf("sub %d: got a message, want closed channel", i)
			}
		case <-time.After(200 * time.Millisecond):
			t.Errorf("sub %d: not closed by Close()", i)
		}
	}
}

func TestPublish_ConcurrentPublishersAreSafe(t *testing.T) {
	b := New()
	defer b.Close()

	s := b.Subscribe()
	if s == nil {
		t.Fatal("Subscribe returned nil")
	}

	// 3 concurrent publishers, each sending 2 messages — 6 total expected.
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b.Publish("a")
			b.Publish("b")
		}()
	}
	wg.Wait()

	got := drainWithin(s, 6, 200*time.Millisecond)
	if len(got) != 6 {
		t.Errorf("got %d messages, want 6 (concurrent publish should be safe & lossless within buffer)", len(got))
	}
}
