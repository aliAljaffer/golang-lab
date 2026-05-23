// 01-rate-limiter — token-bucket implemented with a buffered channel.
//
// You implement:
//   - New(capacity, refillEvery) *Bucket   — start a refiller goroutine
//   - (*Bucket).Allow() bool               — non-blocking try-take a token
//   - (*Bucket).Wait(ctx) error            — blocking take, honors ctx
//   - (*Bucket).Stop()                     — halt the refiller goroutine
//
// The classic token-bucket-with-channels trick:
//   tokens chan struct{}   // buffered, capacity = bucket size
//   - Send a token into the channel to "add" one (drop if full via select+default)
//   - Receive from the channel to "consume" one
//   - A goroutine ticks every `refillEvery` and tries to add a token
//   - Allow() is `select { case <-tokens: return true; default: return false }`
package bucket

import (
	"context"
	"errors"
	"time"
)

type Bucket struct {
	tokens chan struct{}
	stop   chan struct{}
}

// New starts a Bucket pre-filled to `capacity`, with a refill of 1 token every
// `refillEvery`. Caller must call Stop to release the refiller goroutine.
func New(capacity int, refillEvery time.Duration) *Bucket {
	b := &Bucket{
		tokens: make(chan struct{}, capacity),
		stop:   make(chan struct{}),
	}
	// Pre-fill to capacity (initial burst allowance).
	for i := 0; i < capacity; i++ {
		b.tokens <- struct{}{}
	}
	// TODO: start the refiller. Two design points the tests care about:
	//   - it has to stop when Stop() is called (b.stop is your signal).
	//   - when the bucket is already full, adding another token must NOT
	//     block — overflow is silently dropped, not queued. The non-blocking
	//     send pattern (select with a default) is the standard trick.

	_ = time.NewTicker
	return b
}

// Allow returns true if a token was consumed, false otherwise.
// Non-blocking.
func (b *Bucket) Allow() bool {
	// TODO: try-take. The "try" half is the key word — receiving from the
	//   tokens channel takes one, but you must not block when there's
	//   nothing to take. Same non-blocking pattern as the refiller's send.
	return false
}

// Wait blocks until a token is available, ctx is cancelled, or Stop is called.
// Returns nil on success, ctx.Err() on cancel, or ErrStopped if the bucket is stopped.
func (b *Bucket) Wait(ctx context.Context) error {
	// TODO: three things can wake this caller up — pick whichever fires
	//   first and return the matching value (nil / ctx.Err() / ErrStopped).
	return errors.New("Wait: not implemented")
}

// Stop halts the refiller goroutine. Idempotent? No — calling twice will panic
// on the second close. Tests cover this.
func (b *Bucket) Stop() {
	// TODO: signal the refiller to exit. Whatever you do here must also
	//   wake any caller currently blocked in Wait — both watch the same
	//   channel.
}

// ErrStopped is returned by Wait if the bucket is Stop'd while a caller is blocked.
var ErrStopped = errors.New("bucket stopped")
