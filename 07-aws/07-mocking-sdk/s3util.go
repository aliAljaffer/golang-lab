// 07-mocking-sdk — define a tiny interface for the S3 operations you use,
// then write tests against a fake. No real AWS calls.
//
// The pattern: do NOT take `*s3.Client` in your business code. Take an
// interface that names only the methods you call. Production passes the
// real client (it satisfies the interface automatically). Tests pass a fake.
//
// Why this works: SDK v2 method signatures are stable enough that an interface
// like the one below tracks `(*s3.Client).GetObject` 1:1.
package s3util

import (
	"context"
	"errors"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3GetObjectAPI is the slice of the S3 API this package needs.
// `*s3.Client` satisfies it automatically.
type S3GetObjectAPI interface {
	GetObject(ctx context.Context, in *s3.GetObjectInput, opts ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

// FetchKey reads s3://<bucket>/<key> into a byte slice. Bounded by maxBytes
// to keep callers from OOMing on a giant object.
func FetchKey(ctx context.Context, api S3GetObjectAPI, bucket, key string, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		return nil, errors.New("maxBytes must be > 0")
	}
	out, err := api.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, err
	}
	defer out.Body.Close()
	return io.ReadAll(io.LimitReader(out.Body, maxBytes))
}
