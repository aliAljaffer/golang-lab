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
//     (the next batch will get its own retry budget inside S3Uploader.Put).
//   - On ctx.Done, the batcher loop performs one final Flush() before exiting
//     so no buffered lines are lost.
//   - Run returns nil after every tailer + the batcher loop have exited.
//
// keyPrefix is prepended to each Batch.KeySuffix to form the final object key.
func Run(ctx context.Context, tailers []*Tailer, batcher *Batcher, uploader Uploader, keyPrefix string, errOut io.Writer) error {
	// TODO: linesCh := make(chan Line, 1024) (or some sensible buffer).
	// TODO: start one goroutine per Tailer; collect their errors via a sync.WaitGroup or errgroup.Group.
	// TODO: launch a goroutine that closes(linesCh) once every tailer goroutine has returned. (so the batcher loop drains on shutdown).
	// TODO: in the main goroutine (or another), run the batcher loop:
	// TODO:   ticker := time.NewTicker(batcher.MaxAge) (clamp to a sensible min like 100ms).
	// TODO:   defer ticker.Stop().
	// TODO:   loop {
	// TODO:     select {
	// TODO:       case line, ok := <-linesCh:
	// TODO:         if !ok { /* tailers done */ goto finalFlush }
	// TODO:         if batch, full := batcher.Add(line.Body); full {
	// TODO:             upload(ctx, uploader, keyPrefix, batch, errOut)
	// TODO:         }
	// TODO:       case <-ticker.C:
	// TODO:         if batch, ok := batcher.MaybeFlushByAge(); ok {
	// TODO:             upload(ctx, uploader, keyPrefix, batch, errOut)
	// TODO:         }
	// TODO:       case <-ctx.Done():
	// TODO:         goto finalFlush
	// TODO:     }
	// TODO:   }
	// TODO: finalFlush: if batch, ok := batcher.Flush(); ok { upload(...) }
	// TODO: wait for tailer goroutines; return any aggregated error.
	_ = fmt.Sprintf // keep fmt live for when you add error wrapping
	return errors.New("Run not implemented")
}

// upload composes the final object key and hands the batch to the uploader.
// Errors are written to errOut; they do not propagate (so a single bad
// upload doesn't kill the whole shipper).
func upload(ctx context.Context, u Uploader, keyPrefix string, b Batch, errOut io.Writer) {
	// TODO: key := keyPrefix + b.KeySuffix (or path.Join, if you want).
	// TODO: if err := u.Put(ctx, key, b.Body); err != nil { fmt.Fprintln(errOut, "upload:", err) }.
}
