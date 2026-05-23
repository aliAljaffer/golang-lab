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

	// Running state — you'll need somewhere to accumulate raw line bytes,
	// a count of how many bytes that is so far, and the timestamp of the
	// first line in the current batch (for the age-based flush).
}

// Add appends line to the current batch. If this push causes the running
// byte count to meet or exceed MaxBytes, Add finalizes a Batch (gzips,
// computes Castagnoli CRC, stamps KeySuffix) and resets the running state.
//
//	(Batch, true)  -> caller should hand the batch to the uploader
//	(Batch, false) -> still accumulating; nothing to upload yet
func (b *Batcher) Add(line []byte) (Batch, bool) {
	// TODO: append line + '\n' to the running buffer. Two edges to handle:
	//   stamp firstAt on the empty→non-empty transition (so MaybeFlushByAge
	//   has something to measure), and trigger finalize when the raw byte
	//   count crosses MaxBytes.
	return Batch{}, false
}

// MaybeFlushByAge returns a finalized Batch iff at least one line is buffered
// AND Now() - firstAt >= MaxAge. Caller polls this on a ticker.
func (b *Batcher) MaybeFlushByAge() (Batch, bool) {
	// TODO: time-driven flush — only fires when there's something to flush
	//   AND the buffered lines are old enough. Polled on a ticker, so it
	//   must be cheap on the empty path.
	return Batch{}, false
}

// Flush is the unconditional flush used by Run on shutdown. Returns
// (_, false) iff the batch is empty (no upload to do).
func (b *Batcher) Flush() (Batch, bool) {
	// TODO: unconditional flush. The only branch is "is anything buffered?"
	//   — if not, return (_, false) so Run doesn't try to upload an empty
	//   object on shutdown.
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
	// TODO: gzip the buffered lines, CRC the gzipped output, build the key
	//   suffix, capture line count, then RESET running state before returning.
	//   Two gotchas pinned by tests:
	//     - the gzip.Writer must be closed before you CRC; otherwise the
	//       trailer isn't written and the bytes won't match what GCS sees.
	//     - the CRC table MUST be Castagnoli (crc32.MakeTable(crc32.Castagnoli)),
	//       not the default IEEE polynomial. Mixing them up is the load-
	//       bearing GCS pitfall from section 07-gcp.
	return Batch{}
}

// silence unused-import noise while the file is unimplemented; remove once Add/etc reference time.
var _ = time.Second

// silence unused-package noise — Castagnoli usage MUST live in finalize.
// (Removing this var once you import "hash/crc32" in finalize.)
var _ = errors.New
