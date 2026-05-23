package main

import (
	"context"
	"errors"
	"fmt"
	"io"
)

// Run wires the pipeline:
//
//	tailers (one goroutine each, emit Line on linesCh)
//	     -> batcher (single goroutine: Add on every Line; ticker drives MaybeFlushByAge)
//	     -> uploader (Put on every emitted Batch)
//
// Lifecycle:
//   - One goroutine per Tailer; each calls t.Run(ctx, linesCh) and exits when
//     ctx is cancelled.
//   - A single batcher-loop goroutine selects across linesCh and a ticker.
//   - Uploader.Put errors are logged to errOut; they do NOT kill the pipeline
//     (the next batch will get its own retry budget inside GCSUploader.Put).
//   - On ctx.Done, the batcher loop performs one final Flush() before exiting
//     so no buffered lines are lost.
//   - Run returns nil after every tailer + the batcher loop have exited.
//
// keyPrefix is prepended to each Batch.KeySuffix to form the final object key.
func Run(ctx context.Context, tailers []*Tailer, batcher *Batcher, uploader Uploader, keyPrefix string, errOut io.Writer) error {
	// TODO: wire the topology described in the doc comment above. Things
	//   to get right (the test file pins each):
	//
	//   - linesCh as a buffered channel so a slow batcher doesn't block tails.
	//   - One goroutine per Tailer; close linesCh only after all of them
	//     have returned (a WaitGroup + closer goroutine is the usual shape).
	//   - The batcher loop selects over: incoming lines, an age-based ticker,
	//     and ctx.Done. Three exit signals, one finalFlush. Don't drop the
	//     final flush — the test asserts no buffered lines are lost on
	//     shutdown.
	//   - Uploader errors are logged to errOut, not returned — see the
	//     contract above.
	_ = fmt.Sprintf
	return errors.New("Run not implemented")
}

// upload composes the final object key and hands the batch to the uploader.
// Errors are written to errOut; they do not propagate (so a single bad
// upload doesn't kill the whole shipper).
func upload(ctx context.Context, u Uploader, keyPrefix string, b Batch, errOut io.Writer) {
	// TODO: build the object key from keyPrefix + b.KeySuffix and call u.Put.
	//   On error, log to errOut; never return — see the no-propagation
	//   contract above.
}
