// Package cleanup deletes S3 objects under a prefix that are older than a
// cutoff time. Classic janitorial tool — every team has a `tmp/` prefix that
// nobody wants to clean up by hand.
//
// Exercise surface:
//
//	type S3API interface { ListObjectsV2 paginatable; DeleteObject }
//	func Cleanup(ctx, api, bucket, prefix, cutoff, dryRun) ([]string, error)
//
// Returns the keys that were (or would have been) deleted. Order matches
// what the paginator returned.
package cleanup

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3API is the slice of *s3.Client this package uses.
type S3API interface {
	ListObjectsV2(ctx context.Context, in *s3.ListObjectsV2Input, opts ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
	DeleteObject(ctx context.Context, in *s3.DeleteObjectInput, opts ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

// Cleanup deletes every object in <bucket> under <prefix> whose LastModified
// is BEFORE cutoff. If dryRun is true, no DeleteObject calls are made — the
// return list is what WOULD have been deleted.
//
// Hints:
//   - List with &s3.ListObjectsV2Input{Bucket: &bucket, Prefix: &prefix}
//     and the paginator. AWS does the prefix filter server-side.
//   - obj.LastModified is *time.Time. Use .Before(cutoff) to compare.
//   - obj.Key is *string. The DeleteObjectInput needs both Bucket + Key.
//   - If dryRun, append the key to the result but skip the DeleteObject call.
//   - If a DeleteObject fails, return the error (and the keys deleted so far).
func Cleanup(ctx context.Context, api S3API, bucket, prefix string, cutoff time.Time, dryRun bool) ([]string, error) {
	if bucket == "" {
		return nil, errors.New("bucket must not be empty")
	}
	// TODO: implement.
	return nil, errors.New("Cleanup not implemented")
}
