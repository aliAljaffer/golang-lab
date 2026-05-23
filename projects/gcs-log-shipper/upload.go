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
	// TODO: for attempt := 0; attempt <= u.MaxRetries; attempt++ {
	// TODO:     if ctx.Err() != nil { return ctx.Err() }
	// TODO:     err := u.Inner.putObject(ctx, u.Bucket, key, body)
	// TODO:     if err == nil { return nil }
	// TODO:     if !IsTransient(err) { return err }
	// TODO:     if attempt == u.MaxRetries { return err }
	// TODO:     // backoff = u.BaseBackoff * 2^attempt + jitter (use math/rand or crypto/rand for jitter).
	// TODO:     if sleepErr := u.Sleep(ctx, backoff); sleepErr != nil { return sleepErr }
	// TODO: }
	// TODO: return nil // unreachable
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
	// TODO: if err == nil { return false }
	// TODO: if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) { return false }
	// TODO: st, ok := status.FromError(err)
	// TODO: if !ok { return true } // unknown shape -> assume transient (transport hiccup).
	// TODO: switch st.Code() { case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted, codes.Aborted: return true }
	// TODO: return false
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
	// TODO: w := a.Client.Bucket(bucket).Object(key).NewWriter(ctx)
	// TODO: w.ContentEncoding = "gzip"          // body is already gzipped
	// TODO: w.ContentType     = "application/octet-stream"
	// TODO: if _, err := w.Write(body); err != nil { _ = w.Close(); return err }
	// TODO: return w.Close()                    // Close() is what COMMITS the upload.
	return errors.New("gcsClientAdapter.putObject not implemented")
}
