//go:build exercise

package inventory

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// fakeGCS returns canned ListBuckets + per-bucket ListObjects data.
type fakeGCS struct {
	buckets []BucketAttrs
	// objects[bucket] is the canned slice that ListObjects returns.
	objects map[string][]ObjectAttrs
	// errPerBucket[bucket] makes ListObjects fail for that bucket.
	errPerBucket map[string]error
}

func (f *fakeGCS) ListBuckets(_ context.Context, _ string) ([]BucketAttrs, error) {
	return f.buckets, nil
}

func (f *fakeGCS) ListObjects(_ context.Context, bucket string) ([]ObjectAttrs, error) {
	if err, ok := f.errPerBucket[bucket]; ok {
		return nil, err
	}
	return f.objects[bucket], nil
}

func TestInventory_EmptyProject(t *testing.T) {
	f := &fakeGCS{}

	rows, err := Inventory(context.Background(), f, "p")
	if err != nil {
		t.Fatalf("Inventory: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("rows = %v, want empty", rows)
	}
}

func TestInventory_FlattensBucketsAndObjects(t *testing.T) {
	upd := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	f := &fakeGCS{
		buckets: []BucketAttrs{{Name: "alpha"}, {Name: "beta"}},
		objects: map[string][]ObjectAttrs{
			"alpha": {{Name: "a/1.txt", Size: 10, Updated: upd}},
			"beta":  {{Name: "b.txt", Size: 99, Updated: upd}},
		},
	}

	rows, err := Inventory(context.Background(), f, "p")
	if err != nil {
		t.Fatalf("Inventory: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("len(rows) = %d, want 2", len(rows))
	}
	if rows[0].Bucket != "alpha" || rows[0].Name != "a/1.txt" || rows[0].Size != 10 {
		t.Errorf("row[0] = %+v, want {alpha a/1.txt 10}", rows[0])
	}
	if rows[1].Bucket != "beta" || rows[1].Name != "b.txt" || rows[1].Size != 99 {
		t.Errorf("row[1] = %+v, want {beta b.txt 99}", rows[1])
	}
}

func TestInventory_PreservesBucketOrder(t *testing.T) {
	upd := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	f := &fakeGCS{
		buckets: []BucketAttrs{{Name: "zulu"}, {Name: "alpha"}, {Name: "mike"}}, // intentionally not sorted
		objects: map[string][]ObjectAttrs{
			"zulu":  {{Name: "z", Size: 1, Updated: upd}},
			"alpha": {{Name: "a", Size: 1, Updated: upd}},
			"mike":  {{Name: "m", Size: 1, Updated: upd}},
		},
	}

	rows, err := Inventory(context.Background(), f, "p")
	if err != nil {
		t.Fatalf("Inventory: %v", err)
	}
	got := []string{rows[0].Bucket, rows[1].Bucket, rows[2].Bucket}
	if got[0] != "zulu" || got[1] != "alpha" || got[2] != "mike" {
		t.Errorf("bucket order = %v, want [zulu alpha mike] (preserve ListBuckets order; no sorting)", got)
	}
}

func TestInventory_ErrorFromOneBucketBubblesUp(t *testing.T) {
	upd := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	stub := errors.New("permission denied")
	f := &fakeGCS{
		buckets: []BucketAttrs{{Name: "good"}, {Name: "bad"}},
		objects: map[string][]ObjectAttrs{
			"good": {{Name: "k", Size: 1, Updated: upd}},
		},
		errPerBucket: map[string]error{"bad": stub},
	}

	_, err := Inventory(context.Background(), f, "p")
	if !errors.Is(err, stub) {
		t.Errorf("err = %v, want %v", err, stub)
	}
}

func TestWriteCSV_HeaderAndRows(t *testing.T) {
	upd := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	rows := []Row{
		{Bucket: "alpha", Name: "a/1.txt", Size: 10, Updated: upd},
		{Bucket: "beta", Name: "b.txt", Size: 99, Updated: upd},
	}
	var buf bytes.Buffer

	if err := WriteCSV(&buf, rows); err != nil {
		t.Fatalf("WriteCSV: %v", err)
	}
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("len(lines) = %d, want 3 (header + 2 rows): %q", len(lines), buf.String())
	}
	if lines[0] != "bucket,name,size,updated" {
		t.Errorf("header = %q, want bucket,name,size,updated", lines[0])
	}
	if !strings.HasPrefix(lines[1], "alpha,a/1.txt,10,") {
		t.Errorf("row[1] = %q, want prefix alpha,a/1.txt,10,", lines[1])
	}
	if !strings.Contains(lines[1], "2026-05-01T12:00:00Z") {
		t.Errorf("row[1] = %q, want RFC3339 timestamp 2026-05-01T12:00:00Z", lines[1])
	}
}
