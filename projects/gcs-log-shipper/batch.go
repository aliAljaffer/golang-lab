package main

import (
	"errors"
	"time"
)

// Batch is a finished, gzipped batch ready for upload.
//
//   - Body is the gzipped concatenation of the lines (each line followed
//     by '\n'). Decompressing Body must round-trip the original input.
//   - KeySuffix is the deterministic per-batch portion of the object key:
//     "<hostname>/<unixnano>.gz". Run prepends the user's --key-prefix.
//   - Count is the number of lines that went into this batch (0 means the
//     batch was empty and should not have been emitted).
//   - CRC32C is the Castagnoli (NOT IEEE) CRC of Body. GCS uses Castagnoli
//     for its own object integrity hash, so this is the only CRC that's
//     comparable to what GCS returns post-upload. This is the load-bearing
//     GCS quirk from section 07-gcp — pinned by TestBatcher_CRCIsCastagnoli.
type Batch struct {
	Body      []byte
	KeySuffix string
	Count     int
	CRC32C    uint32
}

// Batcher accumulates Lines and flushes when either:
//
//   - the RAW (pre-gzip) byte count meets or exceeds MaxBytes, OR
//   - the age of the first line in the current batch meets or exceeds MaxAge.
//
// Add returns (Batch, true) when adding a line crosses the size threshold;
// the caller hands the Batch to the uploader and the Batcher's running state
// is reset.
//
// MaybeFlushByAge is the time-driven counterpart; the caller polls it on a
// ticker (Run does this). It returns (Batch, true) iff there is at least
// one buffered line AND Now() - firstLineAt >= MaxAge.
//
// Flush forces an immediate flush; on empty state it returns (_, false).
//
// Pure state machine. Not concurrency-safe — Run owns the only instance and
// calls these methods from a single goroutine.
type Batcher struct {
	MaxBytes int
	MaxAge   time.Duration
	Hostname string
	Now      func() time.Time

	// Running state. TODO: declare what you need; suggestions:
	//   - a [][]byte of buffered lines (or a bytes.Buffer of "line\n"… raw)
	//   - an int rawBytes counter
	//   - a time.Time firstAt
}

// Add appends line to the current batch. If this push causes the running
// byte count to meet or exceed MaxBytes, Add finalizes a Batch (gzips,
// computes Castagnoli CRC, stamps KeySuffix) and resets the running state.
//
//	(Batch, true)  -> caller should hand the batch to the uploader
//	(Batch, false) -> still accumulating; nothing to upload yet
func (b *Batcher) Add(line []byte) (Batch, bool) {
	// TODO: if running state is empty, stamp firstAt = b.Now().
	// TODO: append line (and a single '\n') to the buffer; bump rawBytes.
	// TODO: if rawBytes >= MaxBytes, return b.finalize(), true.
	// TODO: otherwise return Batch{}, false.
	return Batch{}, false
}

// MaybeFlushByAge returns a finalized Batch iff at least one line is buffered
// AND Now() - firstAt >= MaxAge. Caller polls this on a ticker.
func (b *Batcher) MaybeFlushByAge() (Batch, bool) {
	// TODO: if rawBytes == 0, return Batch{}, false.
	// TODO: if b.Now().Sub(firstAt) < b.MaxAge, return Batch{}, false.
	// TODO: return b.finalize(), true.
	return Batch{}, false
}

// Flush is the unconditional flush used by Run on shutdown. Returns
// (_, false) iff the batch is empty (no upload to do).
func (b *Batcher) Flush() (Batch, bool) {
	// TODO: if rawBytes == 0, return Batch{}, false.
	// TODO: return b.finalize(), true.
	return Batch{}, false
}

// finalize gzips the accumulated lines, computes the Castagnoli CRC of the
// gzipped output, stamps KeySuffix, resets running state, and returns the
// Batch. Called by Add / MaybeFlushByAge / Flush.
//
// IMPORTANT: the CRC32C table MUST be built with crc32.Castagnoli, NOT the
// default crc32.IEEE — GCS uses Castagnoli server-side, so an IEEE CRC will
// disagree with whatever GCS reports back.
//
// KeySuffix format: "<Hostname>/<firstAt.UTC().UnixNano()>.gz".
func (b *Batcher) finalize() Batch {
	// TODO: gzip the buffered bytes into a bytes.Buffer; close the gzip.Writer to flush.
	// TODO: compute Castagnoli CRC32C of the gzipped bytes.
	// TODO: build KeySuffix from b.Hostname and the captured firstAt timestamp.
	// TODO: capture Count before resetting running state.
	// TODO: reset running state (lines empty, rawBytes 0, firstAt zero).
	// TODO: return the populated Batch.
	return Batch{}
}

// silence unused-import noise while the file is unimplemented; remove once Add/etc reference time.
var _ = time.Second

// silence unused-package noise — Castagnoli usage MUST live in finalize.
// (Removing this var once you import "hash/crc32" in finalize.)
var _ = errors.New
