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
	// TODO: start a goroutine that ticks every refillEvery and tries to add a token:
	// TODO:   go func() {
	// TODO:     t := time.NewTicker(refillEvery)
	// TODO:     defer t.Stop()
	// TODO:     for {
	// TODO:       select {
	// TODO:       case <-b.stop:
	// TODO:         return
	// TODO:       case <-t.C:
	// TODO:         select {
	// TODO:         case b.tokens <- struct{}{}:  // add a token if room
	// TODO:         default:                       // bucket full — drop the tick
	// TODO:         }
	// TODO:       }
	// TODO:     }
	// TODO:   }()

	_ = time.NewTicker
	return b
}

// Allow returns true if a token was consumed, false otherwise.
// Non-blocking.
func (b *Bucket) Allow() bool {
	// TODO: select {
	// TODO: case <-b.tokens:
	// TODO:     return true
	// TODO: default:
	// TODO:     return false
	// TODO: }
	return false
}

// Wait blocks until a token is available, ctx is cancelled, or Stop is called.
// Returns nil on success, ctx.Err() on cancel, or ErrStopped if the bucket is stopped.
func (b *Bucket) Wait(ctx context.Context) error {
	// TODO: select {
	// TODO: case <-b.tokens:    return nil
	// TODO: case <-ctx.Done():  return ctx.Err()
	// TODO: case <-b.stop:      return ErrStopped
	// TODO: }
	return errors.New("Wait: not implemented")
}

// Stop halts the refiller goroutine. Idempotent? No — calling twice will panic
// on the second close. Tests cover this.
func (b *Bucket) Stop() {
	// TODO: close(b.stop)
}

// ErrStopped is returned by Wait if the bucket is Stop'd while a caller is blocked.
var ErrStopped = errors.New("bucket stopped")
