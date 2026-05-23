// 07-mocking-gcs — wrap *storage.Client behind a 3-method interface, then
// test consumers against a hand-rolled fake. No real GCS calls.
//
// The wrinkle vs 07-aws's mocking-sdk:
//   - AWS SDK v2 methods are typed `(ctx, *Input, ...func(*Options)) (*Output, error)`,
//     so you can write an interface that matches `*s3.Client` 1:1 and pass
//     the real client as the interface in production. Zero adapter code.
//   - GCS exposes `*storage.BucketHandle`, `*storage.ObjectHandle`,
//     `*storage.ObjectIterator`, `*storage.Reader`, `*storage.Writer` — all
//     concrete types with unexported fields. You can't pick off
//     `client.Bucket(b).Object(k).NewReader` as an interface method directly
//     because the chain returns concrete types you can't mock.
//   - The fix: wrap your own thin abstraction. Production passes a struct that
//     drives a real `*storage.Client`; tests pass a hand-rolled fake. The
//     abstraction lives in YOUR code, not GCS's.
//
// The interface below names three operations a typical tool needs:
// Read / Write / List. Pick the slice you need; don't make it bigger.
package gcsutil

import (
	"context"
	"errors"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

// GCSAPI is the slice of GCS the consumer needs. Both `*RealGCS` (production)
// and `*fakeGCS` (tests) implement it.
type GCSAPI interface {
	// Read fetches the bytes of gs://<bucket>/<key>. Capped at maxBytes.
	Read(ctx context.Context, bucket, key string, maxBytes int64) ([]byte, error)
	// Write replaces gs://<bucket>/<key> with body.
	Write(ctx context.Context, bucket, key string, body []byte) error
	// List returns one ObjectInfo per object in <bucket> under <prefix>.
	// Drains the iterator internally — the caller doesn't see the iterator
	// pattern at all.
	List(ctx context.Context, bucket, prefix string) ([]ObjectInfo, error)
}

// ObjectInfo is the minimal projection of *storage.ObjectAttrs the wrapper
// exposes. Add fields as consumers need them; keep it small.
type ObjectInfo struct {
	Name    string
	Size    int64
	Updated time.Time
	CRC32C  uint32 // GCS's default object integrity hash
}

// RealGCS is the production adapter. Implements GCSAPI by driving a real
// *storage.Client. Production code constructs one of these and passes it as
// GCSAPI down through the call stack.
type RealGCS struct {
	Client *storage.Client
}

// NewReal builds a RealGCS using ADC. The caller is responsible for
// `defer real.Client.Close()` (or the consumer wraps it).
func NewReal(ctx context.Context) (*RealGCS, error) {
	c, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return &RealGCS{Client: c}, nil
}

func (r *RealGCS) Read(ctx context.Context, bucket, key string, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		return nil, errors.New("maxBytes must be > 0")
	}
	rd, err := r.Client.Bucket(bucket).Object(key).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer rd.Close()
	return io.ReadAll(io.LimitReader(rd, maxBytes))
}

func (r *RealGCS) Write(ctx context.Context, bucket, key string, body []byte) error {
	w := r.Client.Bucket(bucket).Object(key).NewWriter(ctx)
	if _, err := w.Write(body); err != nil {
		_ = w.Close() // best-effort
		return err
	}
	return w.Close() // Close() is what commits — return its error
}

func (r *RealGCS) List(ctx context.Context, bucket, prefix string) ([]ObjectInfo, error) {
	it := r.Client.Bucket(bucket).Objects(ctx, &storage.Query{Prefix: prefix})
	var out []ObjectInfo
	for {
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return out, err
		}
		out = append(out, ObjectInfo{
			Name:    attrs.Name,
			Size:    attrs.Size,
			Updated: attrs.Updated,
			CRC32C:  attrs.CRC32C,
		})
	}
	return out, nil
}

// ---------------------------------------------------------------------------
// Consumer: a function that uses GCSAPI. Tests verify this against a fake.
// ---------------------------------------------------------------------------

// FetchKey reads gs://<bucket>/<key> with a size cap. Same shape as
// 07-aws/07-mocking-sdk's FetchKey — the pattern is identical, only the
// surface differs.
func FetchKey(ctx context.Context, api GCSAPI, bucket, key string, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		return nil, errors.New("maxBytes must be > 0")
	}
	return api.Read(ctx, bucket, key, maxBytes)
}

// TotalSize sums Size across every object in <bucket> with <prefix>. A
// realistic second consumer — demonstrates that one interface drives
// multiple consumers.
func TotalSize(ctx context.Context, api GCSAPI, bucket, prefix string) (int64, error) {
	objs, err := api.List(ctx, bucket, prefix)
	if err != nil {
		return 0, err
	}
	var total int64
	for _, o := range objs {
		total += o.Size
	}
	return total, nil
}
