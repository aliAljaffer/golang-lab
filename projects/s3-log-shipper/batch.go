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
//   - MD5 is the hex md5 of Body. S3 returns the same md5 as the object's
//     ETag for single-part PutObject uploads, so this is the only client-side
//     hash that's directly comparable post-upload. This is the load-bearing
//     S3 quirk (mirrors the Castagnoli CRC requirement in gcs-log-shipper)
//     — pinned by TestBatcher_MD5MatchesGzippedBody.
type Batch struct {
	Body      []byte
	KeySuffix string
	Count     int
	MD5       string // lower-case hex; matches S3 ETag for single-part PUT
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
// computes hex md5 of the gzipped body, stamps KeySuffix) and resets the
// running state.
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

// finalize gzips the accumulated lines, computes the hex md5 of the gzipped
// output, stamps KeySuffix, resets running state, and returns the Batch.
// Called by Add / MaybeFlushByAge / Flush.
//
// IMPORTANT: MD5 is hashed over the gzipped Body, NOT the raw lines. The
// reason is that S3's ETag-for-single-part is computed over the bytes S3
// actually received, which is the gzipped payload. A test pins this so the
// implementation can't drift to "md5 of the plaintext".
//
// KeySuffix format: "<Hostname>/<firstAt.UTC().UnixNano()>.gz".
func (b *Batcher) finalize() Batch {
	// TODO: gzip the buffered bytes into a bytes.Buffer; close the gzip.Writer to flush.
	// TODO: compute md5 of the gzipped bytes; hex-encode lower-case.
	// TODO: build KeySuffix from b.Hostname and the captured firstAt timestamp.
	// TODO: capture Count before resetting running state.
	// TODO: reset running state (lines empty, rawBytes 0, firstAt zero).
	// TODO: return the populated Batch.
	return Batch{}
}

// silence unused-import noise while the file is unimplemented; remove once Add/etc reference time.
var _ = time.Second

// silence unused-package noise — md5/hex usage MUST live in finalize.
var _ = errors.New
