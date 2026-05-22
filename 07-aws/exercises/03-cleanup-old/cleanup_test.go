//go:build exercise

package cleanup

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// fakeS3 returns canned ListObjectsV2 results and records DeleteObject calls.
type fakeS3 struct {
	objects []s3types.Object // primed: what ListObjectsV2 returns

	deletes []string // recorded keys
	delErr  error    // if non-nil, DeleteObject returns this
	listErr error
}

func (f *fakeS3) ListObjectsV2(_ context.Context, in *s3.ListObjectsV2Input, _ ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	// honour prefix filtering so tests can verify it gets passed.
	prefix := ""
	if in.Prefix != nil {
		prefix = *in.Prefix
	}
	out := &s3.ListObjectsV2Output{}
	for _, o := range f.objects {
		if prefix == "" || strings.HasPrefix(*o.Key, prefix) {
			out.Contents = append(out.Contents, o)
		}
	}
	return out, nil
}

func (f *fakeS3) DeleteObject(_ context.Context, in *s3.DeleteObjectInput, _ ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	if f.delErr != nil {
		return nil, f.delErr
	}
	f.deletes = append(f.deletes, *in.Key)
	return &s3.DeleteObjectOutput{}, nil
}

func ptr(s string) *string  { return &s }
func tptr(t time.Time) *time.Time { return &t }

func TestCleanup_DeletesOnlyOldObjects(t *testing.T) {
	cutoff := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	f := &fakeS3{objects: []s3types.Object{
		{Key: ptr("logs/old.log"), LastModified: tptr(cutoff.Add(-48 * time.Hour))},
		{Key: ptr("logs/fresh.log"), LastModified: tptr(cutoff.Add(48 * time.Hour))},
		{Key: ptr("logs/ancient.log"), LastModified: tptr(cutoff.Add(-365 * 24 * time.Hour))},
	}}

	deleted, err := Cleanup(context.Background(), f, "b", "logs/", cutoff, false)
	if err != nil {
		t.Fatalf("Cleanup: %v", err)
	}
	if strings.Join(deleted, ",") != "logs/old.log,logs/ancient.log" {
		t.Errorf("deleted = %v, want [logs/old.log logs/ancient.log]", deleted)
	}
	if strings.Join(f.deletes, ",") != "logs/old.log,logs/ancient.log" {
		t.Errorf("DeleteObject calls = %v, want both old keys", f.deletes)
	}
}

func TestCleanup_DryRunSkipsDelete(t *testing.T) {
	cutoff := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	f := &fakeS3{objects: []s3types.Object{
		{Key: ptr("logs/old.log"), LastModified: tptr(cutoff.Add(-48 * time.Hour))},
	}}

	deleted, err := Cleanup(context.Background(), f, "b", "logs/", cutoff, true)
	if err != nil {
		t.Fatalf("Cleanup: %v", err)
	}
	if len(deleted) != 1 || deleted[0] != "logs/old.log" {
		t.Errorf("dry-run plan = %v, want [logs/old.log]", deleted)
	}
	if len(f.deletes) != 0 {
		t.Errorf("dry-run made %d DeleteObject call(s), want 0: %v", len(f.deletes), f.deletes)
	}
}

func TestCleanup_PassesPrefixThrough(t *testing.T) {
	cutoff := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	f := &fakeS3{objects: []s3types.Object{
		{Key: ptr("logs/old.log"), LastModified: tptr(cutoff.Add(-48 * time.Hour))},
		{Key: ptr("data/old.csv"), LastModified: tptr(cutoff.Add(-48 * time.Hour))},
	}}

	deleted, err := Cleanup(context.Background(), f, "b", "logs/", cutoff, false)
	if err != nil {
		t.Fatalf("Cleanup: %v", err)
	}
	if len(deleted) != 1 || deleted[0] != "logs/old.log" {
		t.Errorf("deleted = %v, want only [logs/old.log] (prefix narrows server-side)", deleted)
	}
}

func TestCleanup_ListErrorAborts(t *testing.T) {
	stub := errors.New("list failed")
	f := &fakeS3{listErr: stub}

	_, err := Cleanup(context.Background(), f, "b", "logs/", time.Now(), false)
	if !errors.Is(err, stub) {
		t.Errorf("err = %v, want %v", err, stub)
	}
}

func TestCleanup_DeleteErrorReturnsPartialList(t *testing.T) {
	cutoff := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	stub := errors.New("delete denied")
	f := &fakeS3{
		objects: []s3types.Object{
			{Key: ptr("logs/old.log"), LastModified: tptr(cutoff.Add(-48 * time.Hour))},
		},
		delErr: stub,
	}

	_, err := Cleanup(context.Background(), f, "b", "logs/", cutoff, false)
	if !errors.Is(err, stub) {
		t.Errorf("err = %v, want %v", err, stub)
	}
}
