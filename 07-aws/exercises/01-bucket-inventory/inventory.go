// Package inventory lists every object across every bucket in an account
// and renders the result as CSV.
//
// Exercise surface:
//
//	type S3API interface { ListBuckets; ListObjectsV2 paginatable }
//	type Row struct { Bucket, Key string; Size int64; LastModified time.Time }
//	func Inventory(ctx, api) ([]Row, error)
//	func WriteCSV(w io.Writer, rows []Row) error
//
// Tests pass a fake S3API. There is no `main.go` — wire your own CLI if you want.
package inventory

import (
	"context"
	"encoding/csv"
	"errors"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3API is the slice of *s3.Client this package uses.
type S3API interface {
	ListBuckets(ctx context.Context, in *s3.ListBucketsInput, opts ...func(*s3.Options)) (*s3.ListBucketsOutput, error)
	ListObjectsV2(ctx context.Context, in *s3.ListObjectsV2Input, opts ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
}

// Row is one CSV line: which bucket, which key, how big, when last modified.
type Row struct {
	Bucket       string
	Key          string
	Size         int64
	LastModified time.Time
}

// Inventory enumerates every bucket and every object across every bucket.
// Rows must be in stable order: buckets in the order ListBuckets returned
// them; keys in the order the paginator returned them.
//
// Hints:
//   - ListBuckets returns *s3.ListBucketsOutput with .Buckets []s3types.Bucket
//   - For each bucket, page with s3.NewListObjectsV2Paginator(api, &s3.ListObjectsV2Input{Bucket: &name})
//   - Skip buckets that return an error (different region, no permission) —
//     but the simpler "fail fast" return-on-error is also fine for now.
func Inventory(ctx context.Context, api S3API) ([]Row, error) {
	// TODO: implement.
	return nil, errors.New("Inventory not implemented")
}

// WriteCSV writes header + rows. Header: bucket,key,size,last_modified.
// last_modified is RFC3339. Use encoding/csv (not Sprintf into a string).
func WriteCSV(w io.Writer, rows []Row) error {
	// TODO: implement using encoding/csv.NewWriter.
	_ = csv.NewWriter
	return errors.New("WriteCSV not implemented")
}
