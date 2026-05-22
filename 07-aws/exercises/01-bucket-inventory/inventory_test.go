//go:build exercise

package inventory

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// fakeS3 returns canned ListBuckets + per-bucket ListObjectsV2 data.
type fakeS3 struct {
	buckets []string
	// objects[bucket] is the canned []s3types.Object that ListObjectsV2 returns.
	objects map[string][]s3types.Object
	// errPerBucket[bucket] makes ListObjectsV2 fail for that bucket.
	errPerBucket map[string]error
}

func (f *fakeS3) ListBuckets(_ context.Context, _ *s3.ListBucketsInput, _ ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
	out := &s3.ListBucketsOutput{}
	for _, b := range f.buckets {
		name := b
		out.Buckets = append(out.Buckets, s3types.Bucket{Name: &name})
	}
	return out, nil
}

func (f *fakeS3) ListObjectsV2(_ context.Context, in *s3.ListObjectsV2Input, _ ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	b := *in.Bucket
	if err, ok := f.errPerBucket[b]; ok {
		return nil, err
	}
	return &s3.ListObjectsV2Output{Contents: f.objects[b]}, nil
}

func ptr(s string) *string { return &s }
func i64(n int64) *int64   { return &n }

func TestInventory_EmptyAccount(t *testing.T) {
	f := &fakeS3{}

	rows, err := Inventory(context.Background(), f)
	if err != nil {
		t.Fatalf("Inventory: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("rows = %v, want empty", rows)
	}
}

func TestInventory_FlattensBucketsAndObjects(t *testing.T) {
	mod := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	f := &fakeS3{
		buckets: []string{"alpha", "beta"},
		objects: map[string][]s3types.Object{
			"alpha": {{Key: ptr("a/1.txt"), Size: i64(10), LastModified: &mod}},
			"beta":  {{Key: ptr("b.txt"), Size: i64(99), LastModified: &mod}},
		},
	}

	rows, err := Inventory(context.Background(), f)
	if err != nil {
		t.Fatalf("Inventory: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("len(rows) = %d, want 2", len(rows))
	}
	if rows[0].Bucket != "alpha" || rows[0].Key != "a/1.txt" || rows[0].Size != 10 {
		t.Errorf("row[0] = %+v, want {alpha a/1.txt 10}", rows[0])
	}
	if rows[1].Bucket != "beta" || rows[1].Key != "b.txt" || rows[1].Size != 99 {
		t.Errorf("row[1] = %+v, want {beta b.txt 99}", rows[1])
	}
}

func TestInventory_PreservesBucketOrder(t *testing.T) {
	mod := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	f := &fakeS3{
		buckets: []string{"zulu", "alpha", "mike"}, // intentionally not sorted
		objects: map[string][]s3types.Object{
			"zulu":  {{Key: ptr("z"), Size: i64(1), LastModified: &mod}},
			"alpha": {{Key: ptr("a"), Size: i64(1), LastModified: &mod}},
			"mike":  {{Key: ptr("m"), Size: i64(1), LastModified: &mod}},
		},
	}

	rows, err := Inventory(context.Background(), f)
	if err != nil {
		t.Fatalf("Inventory: %v", err)
	}
	got := []string{rows[0].Bucket, rows[1].Bucket, rows[2].Bucket}
	if got[0] != "zulu" || got[1] != "alpha" || got[2] != "mike" {
		t.Errorf("bucket order = %v, want [zulu alpha mike] (no sorting)", got)
	}
}

func TestInventory_ErrorFromOneBucketBubblesUp(t *testing.T) {
	mod := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	stub := errors.New("access denied")
	f := &fakeS3{
		buckets: []string{"good", "bad"},
		objects: map[string][]s3types.Object{
			"good": {{Key: ptr("k"), Size: i64(1), LastModified: &mod}},
		},
		errPerBucket: map[string]error{"bad": stub},
	}

	_, err := Inventory(context.Background(), f)
	if !errors.Is(err, stub) {
		t.Errorf("err = %v, want %v", err, stub)
	}
}

func TestWriteCSV_HeaderAndRows(t *testing.T) {
	mod := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	rows := []Row{
		{Bucket: "alpha", Key: "a/1.txt", Size: 10, LastModified: mod},
		{Bucket: "beta", Key: "b.txt", Size: 99, LastModified: mod},
	}
	var buf bytes.Buffer

	if err := WriteCSV(&buf, rows); err != nil {
		t.Fatalf("WriteCSV: %v", err)
	}
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("len(lines) = %d, want 3 (header + 2 rows): %q", len(lines), buf.String())
	}
	if lines[0] != "bucket,key,size,last_modified" {
		t.Errorf("header = %q, want bucket,key,size,last_modified", lines[0])
	}
	if !strings.HasPrefix(lines[1], "alpha,a/1.txt,10,") {
		t.Errorf("row[1] = %q, want prefix alpha,a/1.txt,10,", lines[1])
	}
	if !strings.Contains(lines[1], "2026-05-01T12:00:00Z") {
		t.Errorf("row[1] = %q, want RFC3339 timestamp 2026-05-01T12:00:00Z", lines[1])
	}
}
