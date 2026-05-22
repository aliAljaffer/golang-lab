//go:build exercise

package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// fakeS3 records every call and lets tests prime responses + observe peak
// in-flight count. Same hand-rolled-fake pattern as example 07-mocking-sdk.
type fakeS3 struct {
	mu sync.Mutex

	// primed state — what ListObjectsV2 returns.
	objects map[string]string // key → unquoted ETag

	// recorded calls. tests assert against these.
	puts    []string // keys uploaded
	deletes []string // keys deleted
	lists   int      // ListObjectsV2 calls

	// concurrency tracking.
	inflight   atomic.Int32
	peak       atomic.Int32
	holdPut    chan struct{} // if non-nil, PutObject blocks until closed/sent on
	primedErr  error         // if set, ALL ops return this
}

func (f *fakeS3) ListObjectsV2(_ context.Context, in *s3.ListObjectsV2Input, _ ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lists++
	if f.primedErr != nil {
		return nil, f.primedErr
	}
	out := &s3.ListObjectsV2Output{}
	keys := make([]string, 0, len(f.objects))
	for k := range f.objects {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		etag := "\"" + f.objects[k] + "\""
		out.Contents = append(out.Contents, s3types.Object{Key: ptr(k), ETag: ptr(etag)})
	}
	return out, nil
}

func (f *fakeS3) PutObject(_ context.Context, in *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	n := f.inflight.Add(1)
	defer f.inflight.Add(-1)
	for {
		p := f.peak.Load()
		if n <= p || f.peak.CompareAndSwap(p, n) {
			break
		}
	}
	if f.holdPut != nil {
		<-f.holdPut
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.primedErr != nil {
		return nil, f.primedErr
	}
	f.puts = append(f.puts, *in.Key)
	return &s3.PutObjectOutput{}, nil
}

func (f *fakeS3) DeleteObject(_ context.Context, in *s3.DeleteObjectInput, _ ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.primedErr != nil {
		return nil, f.primedErr
	}
	f.deletes = append(f.deletes, *in.Key)
	return &s3.DeleteObjectOutput{}, nil
}

func ptr(s string) *string { return &s }

// writeFiles drops {relpath: content} under dir and returns dir.
func writeFiles(t *testing.T, m map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for rel, content := range m {
		full := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func mustMD5(s string) string {
	h := md5.New()
	io.WriteString(h, s)
	return hex.EncodeToString(h.Sum(nil))
}

// ---- WalkLocal -------------------------------------------------------------

func TestWalkLocal_FlattensTreeWithForwardSlashKeys(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"a.txt":      "alpha",
		"sub/b.txt":  "bravo",
		"sub/c/d.md": "delta",
	})

	got, err := WalkLocal(dir)
	if err != nil {
		t.Fatalf("WalkLocal: %v", err)
	}
	keys := make([]string, 0, len(got))
	md5s := map[string]string{}
	for _, f := range got {
		keys = append(keys, f.Key)
		md5s[f.Key] = f.MD5
	}
	sort.Strings(keys)
	want := []string{"a.txt", "sub/b.txt", "sub/c/d.md"}
	if strings.Join(keys, ",") != strings.Join(want, ",") {
		t.Fatalf("keys = %v, want %v", keys, want)
	}
	if md5s["a.txt"] != mustMD5("alpha") {
		t.Errorf("md5(a.txt) = %s, want %s", md5s["a.txt"], mustMD5("alpha"))
	}
}

// ---- ListRemote ------------------------------------------------------------

func TestListRemote_UnwrapsETagQuotes(t *testing.T) {
	f := &fakeS3{objects: map[string]string{"k1": "abc123", "k2": "def456"}}

	got, err := ListRemote(context.Background(), f, "b")
	if err != nil {
		t.Fatalf("ListRemote: %v", err)
	}
	if got["k1"] != "abc123" || got["k2"] != "def456" {
		t.Errorf("ListRemote returned %v, want {k1: abc123, k2: def456}", got)
	}
}

// ---- Plan ------------------------------------------------------------------

func TestPlan_UploadsNewAndChanged_SkipsIdentical_OmitsExtrasWithoutDelete(t *testing.T) {
	locals := []LocalFile{
		{Key: "same.txt", MD5: "aaa"},
		{Key: "changed.txt", MD5: "new"},
		{Key: "new.txt", MD5: "fresh"},
	}
	remotes := map[string]string{
		"same.txt":    "aaa",
		"changed.txt": "old",
		"extra.txt":   "zzz",
	}

	plan := Plan(locals, remotes, Options{})

	ops := map[string]string{}
	for _, a := range plan {
		ops[a.Key] = a.Op
	}
	if ops["same.txt"] != "skip" {
		t.Errorf("same.txt op = %q, want skip", ops["same.txt"])
	}
	if ops["changed.txt"] != "upload" {
		t.Errorf("changed.txt op = %q, want upload", ops["changed.txt"])
	}
	if ops["new.txt"] != "upload" {
		t.Errorf("new.txt op = %q, want upload", ops["new.txt"])
	}
	if _, present := ops["extra.txt"]; present {
		t.Errorf("extra.txt was included without --delete (op=%q)", ops["extra.txt"])
	}
}

func TestPlan_DeleteFlagIncludesExtras(t *testing.T) {
	locals := []LocalFile{{Key: "a.txt", MD5: "x"}}
	remotes := map[string]string{"a.txt": "x", "stale.txt": "y"}

	plan := Plan(locals, remotes, Options{Delete: true})

	var sawDelete bool
	for _, a := range plan {
		if a.Op == "delete" && a.Key == "stale.txt" {
			sawDelete = true
		}
	}
	if !sawDelete {
		t.Errorf("plan = %+v, want a {delete stale.txt} entry", plan)
	}
}

// ---- Sync ------------------------------------------------------------------

func TestSync_HappyPath_UploadsAndSkips(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"same.txt": "alpha",
		"new.txt":  "bravo",
	})
	f := &fakeS3{objects: map[string]string{"same.txt": mustMD5("alpha")}}

	up, del, skip, err := Sync(context.Background(), f, Options{
		Bucket: "b", Dir: dir, Concurrency: 2,
	})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if up != 1 || del != 0 || skip != 1 {
		t.Errorf("counts = up:%d del:%d skip:%d, want 1/0/1", up, del, skip)
	}
	if len(f.puts) != 1 || f.puts[0] != "new.txt" {
		t.Errorf("PutObject calls = %v, want [new.txt]", f.puts)
	}
}

func TestSync_DryRunDoesNotCallPutOrDelete(t *testing.T) {
	dir := writeFiles(t, map[string]string{"new.txt": "x"})
	f := &fakeS3{objects: map[string]string{"stale.txt": "yyy"}}

	_, _, _, err := Sync(context.Background(), f, Options{
		Bucket: "b", Dir: dir, Concurrency: 2, DryRun: true, Delete: true,
	})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if len(f.puts) != 0 {
		t.Errorf("dry-run made %d PutObject call(s), want 0: %v", len(f.puts), f.puts)
	}
	if len(f.deletes) != 0 {
		t.Errorf("dry-run made %d DeleteObject call(s), want 0: %v", len(f.deletes), f.deletes)
	}
}

func TestSync_RespectsConcurrencyLimit(t *testing.T) {
	files := map[string]string{}
	for i := 0; i < 12; i++ {
		files[fmt.Sprintf("f%02d.txt", i)] = fmt.Sprintf("content-%d", i)
	}
	dir := writeFiles(t, files)

	hold := make(chan struct{})
	f := &fakeS3{holdPut: hold} // every Put blocks on hold

	done := make(chan error, 1)
	go func() {
		_, _, _, err := Sync(context.Background(), f, Options{
			Bucket: "b", Dir: dir, Concurrency: 3,
		})
		done <- err
	}()

	// give the workers a moment to fill the semaphore.
	deadline := waitForInflight(f, 3, 2_000)
	if !deadline {
		t.Fatalf("inflight never reached >=3 (peak=%d)", f.peak.Load())
	}
	close(hold) // release every blocked put

	if err := <-done; err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if peak := f.peak.Load(); peak > 3 {
		t.Errorf("peak in-flight = %d, want ≤ 3", peak)
	}
	if peak := f.peak.Load(); peak < 2 {
		t.Errorf("peak in-flight = %d, want ≥ 2 (worker pool isn't doing parallel work)", peak)
	}
}

// waitForInflight busy-waits up to maxMs for f.peak to reach target. Returns true on hit.
func waitForInflight(f *fakeS3, target int32, maxMs int) bool {
	for i := 0; i < maxMs; i++ {
		if f.peak.Load() >= target {
			return true
		}
		// 1ms sleep without time.Sleep — yield in a goroutine.
		ch := make(chan struct{}, 1)
		go func() { ch <- struct{}{} }()
		<-ch
	}
	return false
}

func TestSync_DeleteRemovesStaleKeys(t *testing.T) {
	dir := writeFiles(t, map[string]string{"keep.txt": "x"})
	f := &fakeS3{objects: map[string]string{
		"keep.txt":  mustMD5("x"),
		"stale.txt": "yyy",
	}}

	up, del, _, err := Sync(context.Background(), f, Options{
		Bucket: "b", Dir: dir, Concurrency: 2, Delete: true,
	})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if up != 0 || del != 1 {
		t.Errorf("counts = up:%d del:%d, want 0/1", up, del)
	}
	if len(f.deletes) != 1 || f.deletes[0] != "stale.txt" {
		t.Errorf("DeleteObject calls = %v, want [stale.txt]", f.deletes)
	}
}

func TestSync_PropagatesError(t *testing.T) {
	dir := writeFiles(t, map[string]string{"f.txt": "x"})
	stub := errors.New("boom")
	f := &fakeS3{primedErr: stub}

	_, _, _, err := Sync(context.Background(), f, Options{
		Bucket: "b", Dir: dir, Concurrency: 1,
	})
	if !errors.Is(err, stub) {
		t.Errorf("err = %v, want %v", err, stub)
	}
}
