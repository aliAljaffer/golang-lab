// 03-pipeline — a 3-stage channel pipeline: Source -> Square -> Sum.
//
// You implement:
//   - Source(ctx, nums)   <-chan int        // emit nums one at a time, close when done
//   - Square(ctx, in)     <-chan int        // for each n received, emit n*n
//   - Sum(ctx, in)        (int, error)      // accumulate until in closes
//
// Each stage:
//   - Returns its output channel synchronously, work happens on a goroutine.
//   - Closes its output channel exactly once, after the work goroutine returns.
//   - Bails out cleanly when ctx is cancelled: stop sending, close output, return.
//
// Composability is the point — a caller writes:
//     sum, err := Sum(ctx, Square(ctx, Source(ctx, nums)))
// and the three stages stream concurrently.
package pipeline

import (
	"context"
)

// Source emits each int in nums onto the returned channel, in order, then closes.
// Returns immediately; the actual sends happen on a goroutine.
func Source(ctx context.Context, nums []int) <-chan int {
	out := make(chan int)
	// TODO: go func() {
	// TODO:     defer close(out)
	// TODO:     for _, n := range nums {
	// TODO:         select {
	// TODO:         case out <- n:
	// TODO:         case <-ctx.Done():
	// TODO:             return
	// TODO:         }
	// TODO:     }
	// TODO: }()

	close(out)
	_ = ctx
	return out
}

// Square reads ints from `in`, squares each, and emits onto the returned channel.
// Closes its output when `in` is closed or ctx is cancelled.
func Square(ctx context.Context, in <-chan int) <-chan int {
	out := make(chan int)
	// TODO: go func() {
	// TODO:     defer close(out)
	// TODO:     for {
	// TODO:         select {
	// TODO:         case n, ok := <-in:
	// TODO:             if !ok { return }
	// TODO:             select {
	// TODO:             case out <- n * n:
	// TODO:             case <-ctx.Done():
	// TODO:                 return
	// TODO:             }
	// TODO:         case <-ctx.Done():
	// TODO:             return
	// TODO:         }
	// TODO:     }
	// TODO: }()

	close(out)
	_ = ctx
	return out
}

// Sum accumulates every value from `in` until the channel closes (returns nil err)
// or ctx is cancelled (returns the partial sum and ctx.Err()).
func Sum(ctx context.Context, in <-chan int) (int, error) {
	var total int
	// TODO: for {
	// TODO:     select {
	// TODO:     case v, ok := <-in:
	// TODO:         if !ok {
	// TODO:             return total, nil
	// TODO:         }
	// TODO:         total += v
	// TODO:     case <-ctx.Done():
	// TODO:         return total, ctx.Err()
	// TODO:     }
	// TODO: }

	_ = ctx
	_ = in
	return total, nil
}
