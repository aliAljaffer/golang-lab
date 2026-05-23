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
	// TODO: do the work on a goroutine, then return `out` synchronously.
	//   Two things every send-side stage owes its caller:
	//     1. close(out) exactly once, when there's nothing more to send.
	//     2. honour ctx — if the consumer disappears, don't block forever
	//        trying to send into a channel no one is reading.
	//   `select` with a ctx.Done() case on each send is the usual shape.

	close(out)
	_ = ctx
	return out
}

// Square reads ints from `in`, squares each, and emits onto the returned channel.
// Closes its output when `in` is closed or ctx is cancelled.
func Square(ctx context.Context, in <-chan int) <-chan int {
	out := make(chan int)
	// TODO: middle stage — needs both ctx-cancel awareness AND end-of-input
	//   detection. The two-value receive form `v, ok := <-in` is how you
	//   distinguish "closed and drained" from a live value; that's the
	//   signal to close `out` and return. Sends must also respect ctx,
	//   otherwise a vanished consumer will deadlock you.

	close(out)
	_ = ctx
	return out
}

// Sum accumulates every value from `in` until the channel closes (returns nil err)
// or ctx is cancelled (returns the partial sum and ctx.Err()).
func Sum(ctx context.Context, in <-chan int) (int, error) {
	var total int
	// TODO: terminal stage — accumulate until one of two things happens:
	//   the input is closed-and-drained (return total, nil), or ctx is
	//   cancelled (return whatever you've accumulated so far + ctx.Err()).
	//   The ctx-cancel test pins both: partial sum is preserved AND error
	//   is errors.Is(context.Canceled).

	_ = ctx
	_ = in
	return total, nil
}
