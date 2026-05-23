package main

import (
	"bytes"
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Uploader is anything that can push a single batch to a remote sink. The
// production impl is *S3Uploader; tests use a capture impl that records
// the (key, body) pairs it received.
type Uploader interface {
	Put(ctx context.Context, key string, body []byte) error
}

// putter is the low-level S3 write surface — the slice S3Uploader uses.
// Production passes *s3ClientAdapter wrapping a real *s3.Client; tests
// pass a fake that returns programmed errors so retry/backoff is pinnable
// without ever touching S3.
//
// Why this isn't just "Uploader inside Uploader": the retry policy is the
// thing tests want to verify in isolation. Splitting it out keeps the test
// surface small (the fake only has to return the right errors in the right
// order) and matches the gcs-log-shipper shape so the diff between the two
// capstones is small.
type putter interface {
	putObject(ctx context.Context, bucket, key string, body []byte) error
}

// S3Uploader wraps a putter with retry/backoff semantics.
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
type S3Uploader struct {
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
func (u *S3Uploader) Put(ctx context.Context, key string, body []byte) error {
	// TODO: implement the retry loop. The behaviour contract above is the
	//   spec; the test file pins each branch (success, transient-then-success,
	//   permanent, exhausted-retries, ctx-cancel mid-sleep).
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
	return errors.New("S3Uploader.Put not implemented")
}

// IsTransient returns true for errors that warrant a retry: S3 5xx-class
// codes (ServiceUnavailable, InternalError), request timeouts (RequestTimeout),
// throttling (Throttling, ThrottlingException, SlowDown, TooManyRequests),
// or generic transport errors with no embedded smithy code.
//
// Returns false for permanent errors: AccessDenied, NoSuchBucket,
// NoSuchKey, InvalidAccessKeyId, SignatureDoesNotMatch, and any other
// client-fault (4xx) error code. Also false for ctx.Err() — there's no point
// retrying a cancelled context.
//
// Use `github.com/aws/smithy-go` — `errors.As(err, &smithy.APIError)` to
// extract the error code from a wrapped SDK error. If the error has no
// smithy code at all, assume "transport hiccup" and treat it as transient.
func IsTransient(err error) bool {
	// TODO: classify err. The interesting decisions:
	//   - ctx errors are NOT transient — retrying a cancelled call is silly.
	//   - errors that don't carry a smithy.APIError (raw transport errors)
	//     are conservatively treated as transient — they're usually DNS,
	//     EOF, TLS hiccups worth one more shot.
	//   - the docstring above names the exact smithy error codes that count
	//     as transient. Everything else (AccessDenied, NoSuchBucket, …) is
	//     permanent.
	//   See `github.com/aws/smithy-go` — errors.As against smithy.APIError
	//   gives you ErrorCode() and ErrorFault().
	return false
}

// s3ClientAdapter is the production putter: drives a real *s3.Client.
// Tests do NOT use this — they wire a fake putter into S3Uploader directly.
//
// Unlike the GCS sibling, S3 PutObject commits when it returns success —
// there's no separate Close() ritual.
type s3ClientAdapter struct {
	Client *s3.Client
}

// putObject writes body to s3://bucket/key and returns any error.
func (a *s3ClientAdapter) putObject(ctx context.Context, bucket, key string, body []byte) error {
	// TODO: call PutObject on a.Client. Don't forget Content-Encoding=gzip
	//   on the request — the batches arrive here already gzipped, and the
	//   consumer needs to know that so they get the bytes uncompressed
	//   automatically. ContentType is also worth setting.
	_ = bytes.NewReader
	return errors.New("s3ClientAdapter.putObject not implemented")
}
