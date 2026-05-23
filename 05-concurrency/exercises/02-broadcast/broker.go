// 02-broadcast — fan-out one value to N subscribers.
//
// You implement:
//   - New() *Broker
//   - (*Broker).Subscribe() <-chan string     // returns a fresh receive-only channel
//   - (*Broker).Unsubscribe(ch <-chan string) // remove a subscriber + close its channel
//   - (*Broker).Publish(msg string)           // deliver msg to every subscriber
//   - (*Broker).Close()                       // close every subscriber channel; further Publish is a no-op
//
// Design notes:
//   - Per-subscriber channels are buffered (subBufSize) so a single slow consumer
//     doesn't block Publish for everyone. Tests stay within the buffer.
//   - A sync.RWMutex protects the subscriber set. Publish takes RLock (many can
//     publish concurrently); Subscribe/Unsubscribe/Close take Lock.
//   - Don't double-close a subscriber channel: Unsubscribe + Close after, or vice
//     versa, must be safe. Track which channels have been closed.
package broker

import (
	"sync"
)

const subBufSize = 8

type Broker struct {
	mu     sync.RWMutex
	subs   map[chan string]struct{}
	closed bool
}

// New returns a fresh Broker.
func New() *Broker {
	// TODO: make sure the subscriber set is usable before anyone calls
	//   Subscribe — a nil map will panic on the first write.
	return &Broker{}
}

// Subscribe returns a new receive-only channel that will see every message
// Publish'd after this call. The channel is closed when Unsubscribe(ch) is
// called or when the broker is Close'd.
func (b *Broker) Subscribe() <-chan string {
	b.mu.Lock()
	defer b.mu.Unlock()

	// TODO: hand back a buffered receive channel and remember it in b.subs.
	//   Edge case the tests cover: subscribing AFTER Close() — the caller
	//   still does `for msg := range ch`, so the channel you return must
	//   not leave them hanging forever. A pre-closed channel is the
	//   conventional answer.
	return nil
}

// Unsubscribe removes ch from the broker and closes it.
// Calling Unsubscribe with a channel that's already been removed (or never
// subscribed) is a no-op.
func (b *Broker) Unsubscribe(ch <-chan string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// TODO: find ch in b.subs, remove it, and close it. The annoying bit:
	//   b.subs is keyed by `chan string` but the parameter is `<-chan string`,
	//   so a direct map lookup won't work — you have to scan and compare via
	//   a directional conversion. Skip silently if not found (no-op contract).

	_ = ch
}

// Publish delivers msg to every current subscriber. If a subscriber's channel
// is full, the message for that subscriber is dropped (non-blocking send) —
// document this behavior so callers don't expect lossless delivery to slow
// consumers.
func (b *Broker) Publish(msg string) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return
	}

	// TODO: deliver msg to every current subscriber. A slow consumer must
	//   NOT block the others, so use a non-blocking send and drop on a full
	//   buffer — that's the lossy-but-live contract documented above.

	_ = msg
}

// Close removes and closes every subscriber. Further Publish calls are no-ops.
// Idempotent.
func (b *Broker) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}
	b.closed = true

	// TODO: tear down every subscriber so their `range` loops exit. The
	//   double-close concern (Unsubscribe + Close on the same channel) is
	//   already handled — Unsubscribe deletes from b.subs first, so here
	//   you can close whatever is still in the map.
}
