//go:build exercise

package cleanup

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// fakeGCS returns a canned object list (honouring prefix) and records deletes.
type fakeGCS struct {
	objects []ObjectAttrs // primed: what ListObjects returns (full set; we filter by prefix below)

	deletes []string // recorded names
	delErr  error
	listErr error
}

func (f *fakeGCS) ListObjects(_ context.Context, _ string, prefix string) ([]ObjectAttrs, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	out := make([]ObjectAttrs, 0, len(f.objects))
	for _, o := range f.objects {
		if prefix == "" || strings.HasPrefix(o.Name, prefix) {
			out = append(out, o)
		}
	}
	return out, nil
}

func (f *fakeGCS) DeleteObject(_ context.Context, _ string, name string) error {
	if f.delErr != nil {
		return f.delErr
	}
	f.deletes = append(f.deletes, name)
	return nil
}

func TestCleanup_DeletesOnlyOldObjects(t *testing.T) {
	cutoff := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	f := &fakeGCS{objects: []ObjectAttrs{
		{Name: "logs/old.log", Updated: cutoff.Add(-48 * time.Hour)},
		{Name: "logs/fresh.log", Updated: cutoff.Add(48 * time.Hour)},
		{Name: "logs/ancient.log", Updated: cutoff.Add(-365 * 24 * time.Hour)},
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
	f := &fakeGCS{objects: []ObjectAttrs{
		{Name: "logs/old.log", Updated: cutoff.Add(-48 * time.Hour)},
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
	f := &fakeGCS{objects: []ObjectAttrs{
		{Name: "logs/old.log", Updated: cutoff.Add(-48 * time.Hour)},
		{Name: "data/old.csv", Updated: cutoff.Add(-48 * time.Hour)},
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
	f := &fakeGCS{listErr: stub}

	_, err := Cleanup(context.Background(), f, "b", "logs/", time.Now(), false)
	if !errors.Is(err, stub) {
		t.Errorf("err = %v, want %v", err, stub)
	}
}

func TestCleanup_DeleteErrorReturnsPartialList(t *testing.T) {
	cutoff := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	stub := errors.New("delete denied")
	f := &fakeGCS{
		objects: []ObjectAttrs{
			{Name: "logs/old.log", Updated: cutoff.Add(-48 * time.Hour)},
		},
		delErr: stub,
	}

	_, err := Cleanup(context.Background(), f, "b", "logs/", cutoff, false)
	if !errors.Is(err, stub) {
		t.Errorf("err = %v, want %v", err, stub)
	}
}

func TestCleanup_EmptyBucketRejected(t *testing.T) {
	f := &fakeGCS{}

	_, err := Cleanup(context.Background(), f, "", "logs/", time.Now(), false)
	if err == nil {
		t.Error("Cleanup(bucket=\"\"): err = nil, want validation error")
	}
}
