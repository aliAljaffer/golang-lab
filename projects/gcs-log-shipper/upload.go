package main

import (
	"context"
	"errors"
	"time"

	"cloud.google.com/go/storage"
)

// Uploader is anything that can push a single batch to a remote sink. The
// production impl is *GCSUploader; tests use a capture impl that records
// the (key, body) pairs it received.
type Uploader interface {
	Put(ctx context.Context, key string, body []byte) error
}

// putter is the low-level GCS write surface — the slice GCSUploader uses.
// Production passes *gcsClientAdapter wrapping a real *storage.Client;
// tests pass a fake that returns programmed errors so retry/backoff is
// pinnable without ever touching GCS.
//
// Why this isn't just "Uploader inside Uploader": the retry policy is the
// thing tests want to verify in isolation. Splitting it out keeps the test
// surface small (the fake only has to return the right errors in the right
// order) and matches the s3-log-shipper shape.
type putter interface {
	putObject(ctx context.Context, bucket, key string, body []byte) error
}

// GCSUploader wraps a putter with retry/backoff semantics.
//
// Behavior contract:
//   - nil error                            -> return nil, no retry
//   - IsTransient(err)                     -> retry up to MaxRetries; backoff
//                                             exponentially from BaseBackoff
//                                             with jitter; ctx-aware Sleep
//   - permanent err (!IsTransient(err))    -> return immediately, no retry
//   - ctx.Done before / between attempts   -> return ctx.Err()
//
// Concurrency: safe iff the underlying putter is.
type GCSUploader struct {
	Inner       putter
	Bucket      string
	MaxRetries  int
	BaseBackoff time.Duration

	// Now and Sleep are injected so tests drive backoff deterministically.
	// Production wires Now = time.Now, Sleep = ctxSleep from main.go.
	Now   func() time.Time
	Sleep func(ctx context.Context, d time.Duration) error
}

// Put uploads body under bucket/key, with retry on transient errors.
func (u *GCSUploader) Put(ctx context.Context, key string, body []byte) error {
	// TODO: implement the retry loop. The behaviour contract above is the
	//   spec; the test file pins each branch.
	//
	//   Two things easy to get subtly wrong:
	//     - the loop bound: how many *attempts* does MaxRetries permit? The
	//       test name "retries N times then returns last error" tells you.
	//     - the sleep contract: u.Sleep is ctx-aware and returns an error
	//       when ctx fires. Propagate that, don't swallow it.
	//
	//   Backoff = BaseBackoff << attempt (or *2^attempt) + jitter. Jitter
	//   matters for thundering-herd in prod but tests only check that the
	//   backoff *grows* — pin it however the test allows.
	return errors.New("GCSUploader.Put not implemented")
}

// IsTransient returns true for errors that warrant a retry: gRPC
// Unavailable / DeadlineExceeded / ResourceExhausted / Aborted, or generic
// transport errors with no embedded gRPC code.
//
// Returns false for permanent errors: PermissionDenied, NotFound,
// InvalidArgument, AlreadyExists, FailedPrecondition. Also false for
// ctx.Err() — there's no point retrying a cancelled context.
//
// Use `google.golang.org/grpc/status` + `google.golang.org/grpc/codes` to
// extract the gRPC code from an error. If the error has no gRPC code at all,
// assume "transport hiccup" and treat it as transient.
func IsTransient(err error) bool {
	// TODO: classify err. The interesting decisions:
	//   - ctx errors are NOT transient — retrying a cancelled call is silly.
	//   - errors that don't carry a gRPC status (raw transport errors) are
	//     conservatively treated as transient — usually DNS/EOF/TLS hiccups
	//     worth one more shot.
	//   - the docstring above names the exact gRPC codes that count as
	//     transient. Everything else (PermissionDenied, NotFound, …) is
	//     permanent. See google.golang.org/grpc/status + .../codes.
	return false
}

// gcsClientAdapter is the production putter: drives a real *storage.Client.
// Tests do NOT use this — they wire a fake putter into GCSUploader directly.
//
// The Writer/Close ritual is the GCS-specific gotcha to internalize: writes
// are buffered until Close() succeeds. A bare `defer w.Close()` hides the
// commit error.
type gcsClientAdapter struct {
	Client *storage.Client
}

// putObject writes body to gs://bucket/key and returns the commit error.
func (a *gcsClientAdapter) putObject(ctx context.Context, bucket, key string, body []byte) error {
	// TODO: open a storage.Writer for gs://bucket/key. The GCS-specific
	//   gotcha: writes are BUFFERED until Close() returns successfully —
	//   that's the actual commit point. A bare `defer w.Close()` hides
	//   the commit error, so capture it explicitly. Also set ContentEncoding
	//   to "gzip" (body is already compressed) so reads decompress
	//   automatically.
	return errors.New("gcsClientAdapter.putObject not implemented")
}
