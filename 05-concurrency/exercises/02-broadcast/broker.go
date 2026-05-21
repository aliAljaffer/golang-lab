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
	// TODO: return &Broker{subs: map[chan string]struct{}{}}

	return &Broker{}
}

// Subscribe returns a new receive-only channel that will see every message
// Publish'd after this call. The channel is closed when Unsubscribe(ch) is
// called or when the broker is Close'd.
func (b *Broker) Subscribe() <-chan string {
	b.mu.Lock()
	defer b.mu.Unlock()

	// TODO: if b.closed { return a pre-closed channel so the caller's `range` exits cleanly }
	// TODO: ch := make(chan string, subBufSize)
	// TODO: b.subs[ch] = struct{}{}
	// TODO: return ch

	return nil
}

// Unsubscribe removes ch from the broker and closes it.
// Calling Unsubscribe with a channel that's already been removed (or never
// subscribed) is a no-op.
func (b *Broker) Unsubscribe(ch <-chan string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// TODO: for k := range b.subs {
	// TODO:     if (<-chan string)(k) == ch {
	// TODO:         delete(b.subs, k)
	// TODO:         close(k)
	// TODO:         return
	// TODO:     }
	// TODO: }

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

	// TODO: for ch := range b.subs {
	// TODO:     select {
	// TODO:     case ch <- msg:
	// TODO:     default:
	// TODO:         // subscriber too slow — drop
	// TODO:     }
	// TODO: }

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

	// TODO: for ch := range b.subs {
	// TODO:     close(ch)
	// TODO:     delete(b.subs, ch)
	// TODO: }
}
