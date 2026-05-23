// Package inventory lists every object across every bucket in a project
// and renders the result as CSV.
//
// Exercise surface:
//
//	type GCSAPI interface { Buckets; Objects }
//	type Row struct { Bucket, Name string; Size int64; Updated time.Time }
//	func Inventory(ctx, api, project) ([]Row, error)
//	func WriteCSV(w io.Writer, rows []Row) error
//
// Tests pass a fake GCSAPI. There is no `main.go` — wire your own CLI if you
// want. The GCP-flavoured cousin of 07-aws/exercises/01-bucket-inventory.
package inventory

import (
	"context"
	"encoding/csv"
	"errors"
	"io"
	"time"
)

// BucketAttrs is the minimal projection of *storage.BucketAttrs needed here.
type BucketAttrs struct {
	Name    string
	Created time.Time
}

// ObjectAttrs is the minimal projection of *storage.ObjectAttrs needed here.
type ObjectAttrs struct {
	Name    string
	Size    int64
	Updated time.Time
}

// GCSAPI is the slice of GCS this package uses. Same wrapper-style pattern
// as 07-mocking-gcs — we don't try to interface *storage.Client directly.
type GCSAPI interface {
	// ListBuckets returns every bucket in <project>, in iterator order.
	ListBuckets(ctx context.Context, project string) ([]BucketAttrs, error)
	// ListObjects returns every object in <bucket> (no prefix filter needed
	// for an inventory), in iterator order.
	ListObjects(ctx context.Context, bucket string) ([]ObjectAttrs, error)
}

// Row is one CSV line: which bucket, which object, how big, when last updated.
type Row struct {
	Bucket  string
	Name    string
	Size    int64
	Updated time.Time
}

// Inventory enumerates every bucket and every object across every bucket.
// Rows must be in stable order: buckets in the order ListBuckets returned
// them; objects in the order ListObjects returned them.
//
// Hints:
//   - Call api.ListBuckets(ctx, project) once.
//   - For each bucket, call api.ListObjects(ctx, bucketName).
//   - Flatten into []Row preserving order.
//   - If a per-bucket ListObjects errors (different region, no perm), the
//     simplest contract is "fail fast" — return the error. (A production
//     tool might log + skip the bucket; the tests pin fail-fast.)
func Inventory(ctx context.Context, api GCSAPI, project string) ([]Row, error) {
	// TODO: implement.
	return nil, errors.New("Inventory not implemented")
}

// WriteCSV writes header + rows. Header: bucket,name,size,updated.
// updated is RFC3339. Use encoding/csv (not Sprintf into a string).
func WriteCSV(w io.Writer, rows []Row) error {
	// TODO: implement using encoding/csv.NewWriter.
	_ = csv.NewWriter
	return errors.New("WriteCSV not implemented")
}
