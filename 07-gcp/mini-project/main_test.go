//go:build exercise

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

// fakeGCS records every call and lets tests prime responses + observe peak
// in-flight upload count. Same hand-rolled pattern as 07-aws/mini-project's
// fakeS3 and 07-gcp/07-mocking-gcs/fakeGCS.
type fakeGCS struct {
	mu sync.Mutex

	// primed: what List returns.
	objects map[string]RemoteObject // name → RemoteObject

	// recorded calls.
	uploads []string // keys uploaded
	deletes []string // keys deleted
	lists   int      // List calls

	// concurrency tracking.
	inflight atomic.Int32
	peak     atomic.Int32

	// behaviour knobs.
	holdUpload chan struct{} // if non-nil, Upload blocks until closed
	primedErr  error         // if set, ALL ops return this
}

func (f *fakeGCS) List(_ context.Context, bucket, _ string) ([]RemoteObject, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lists++
	if f.primedErr != nil {
		return nil, f.primedErr
	}
	out := make([]RemoteObject, 0, len(f.objects))
	keys := make([]string, 0, len(f.objects))
	for k := range f.objects {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		out = append(out, f.objects[k])
	}
	_ = bucket
	return out, nil
}

func (f *fakeGCS) Upload(_ context.Context, _ string, key string, body io.Reader) error {
	n := f.inflight.Add(1)
	defer f.inflight.Add(-1)
	for {
		p := f.peak.Load()
		if n <= p || f.peak.CompareAndSwap(p, n) {
			break
		}
	}
	if f.holdUpload != nil {
		<-f.holdUpload
	}
	// Drain the body so the caller's file-handle close is observable.
	_, _ = io.Copy(io.Discard, body)
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.primedErr != nil {
		return f.primedErr
	}
	f.uploads = append(f.uploads, key)
	return nil
}

func (f *fakeGCS) Delete(_ context.Context, _ string, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.primedErr != nil {
		return f.primedErr
	}
	f.deletes = append(f.deletes, key)
	return nil
}

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

// crc returns the Castagnoli CRC32 of s — same algorithm GCS uses.
func crc(s string) uint32 {
	return crc32.Checksum([]byte(s), castagnoli)
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
	crcs := map[string]uint32{}
	for _, f := range got {
		keys = append(keys, f.Key)
		crcs[f.Key] = f.CRC32C
	}
	sort.Strings(keys)
	want := []string{"a.txt", "sub/b.txt", "sub/c/d.md"}
	if strings.Join(keys, ",") != strings.Join(want, ",") {
		t.Fatalf("keys = %v, want %v", keys, want)
	}
	if crcs["a.txt"] != crc("alpha") {
		t.Errorf("crc(a.txt) = %d, want %d (Castagnoli, not IEEE)", crcs["a.txt"], crc("alpha"))
	}
}

// ---- ListRemote ------------------------------------------------------------

func TestListRemote_BuildsKeyedMap(t *testing.T) {
	f := &fakeGCS{objects: map[string]RemoteObject{
		"k1": {Name: "k1", Size: 100, CRC32C: 0xAA},
		"k2": {Name: "k2", Size: 200, CRC32C: 0xBB},
	}}

	got, err := ListRemote(context.Background(), f, "b")
	if err != nil {
		t.Fatalf("ListRemote: %v", err)
	}
	if got["k1"].CRC32C != 0xAA || got["k2"].CRC32C != 0xBB {
		t.Errorf("ListRemote returned %v, want {k1:0xAA, k2:0xBB}", got)
	}
}

// ---- Plan ------------------------------------------------------------------

func TestPlan_UploadsNewAndChanged_SkipsIdentical_OmitsExtrasWithoutDelete(t *testing.T) {
	locals := []LocalFile{
		{Key: "same.txt", CRC32C: 0xAAA},
		{Key: "changed.txt", CRC32C: 0xFFF},
		{Key: "new.txt", CRC32C: 0xBEEF},
	}
	remotes := map[string]RemoteObject{
		"same.txt":    {Name: "same.txt", CRC32C: 0xAAA},
		"changed.txt": {Name: "changed.txt", CRC32C: 0x111},
		"extra.txt":   {Name: "extra.txt", CRC32C: 0x999},
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
	locals := []LocalFile{{Key: "a.txt", CRC32C: 0x1}}
	remotes := map[string]RemoteObject{
		"a.txt":     {Name: "a.txt", CRC32C: 0x1},
		"stale.txt": {Name: "stale.txt", CRC32C: 0x9},
	}

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

func TestPlan_DeterministicOrder(t *testing.T) {
	locals := []LocalFile{
		{Key: "zulu.txt", CRC32C: 0x1},
		{Key: "alpha.txt", CRC32C: 0x2},
		{Key: "mike.txt", CRC32C: 0x3},
	}

	plan := Plan(locals, map[string]RemoteObject{}, Options{})

	keys := make([]string, len(plan))
	for i, a := range plan {
		keys[i] = a.Key
	}
	want := []string{"alpha.txt", "mike.txt", "zulu.txt"}
	if strings.Join(keys, ",") != strings.Join(want, ",") {
		t.Errorf("plan order = %v, want %v (sorted)", keys, want)
	}
}

// ---- Sync ------------------------------------------------------------------

func TestSync_HappyPath_UploadsAndSkips(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"same.txt": "alpha",
		"new.txt":  "bravo",
	})
	f := &fakeGCS{objects: map[string]RemoteObject{
		"same.txt": {Name: "same.txt", CRC32C: crc("alpha")},
	}}

	up, del, skip, err := Sync(context.Background(), f, Options{
		Bucket: "b", Dir: dir, Concurrency: 2,
	})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if up != 1 || del != 0 || skip != 1 {
		t.Errorf("counts = up:%d del:%d skip:%d, want 1/0/1", up, del, skip)
	}
	if len(f.uploads) != 1 || f.uploads[0] != "new.txt" {
		t.Errorf("Upload calls = %v, want [new.txt]", f.uploads)
	}
}

func TestSync_DryRunMakesNoMutatingCalls(t *testing.T) {
	dir := writeFiles(t, map[string]string{"new.txt": "x"})
	f := &fakeGCS{objects: map[string]RemoteObject{"stale.txt": {Name: "stale.txt", CRC32C: 0x9}}}

	_, _, _, err := Sync(context.Background(), f, Options{
		Bucket: "b", Dir: dir, Concurrency: 2, DryRun: true, Delete: true,
	})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if len(f.uploads) != 0 {
		t.Errorf("dry-run made %d Upload call(s), want 0: %v", len(f.uploads), f.uploads)
	}
	if len(f.deletes) != 0 {
		t.Errorf("dry-run made %d Delete call(s), want 0: %v", len(f.deletes), f.deletes)
	}
}

func TestSync_RespectsConcurrencyLimit(t *testing.T) {
	files := map[string]string{}
	for i := 0; i < 12; i++ {
		files[fmt.Sprintf("f%02d.txt", i)] = fmt.Sprintf("content-%d", i)
	}
	dir := writeFiles(t, files)

	hold := make(chan struct{})
	f := &fakeGCS{holdUpload: hold} // every Upload blocks on hold

	done := make(chan error, 1)
	go func() {
		_, _, _, err := Sync(context.Background(), f, Options{
			Bucket: "b", Dir: dir, Concurrency: 3,
		})
		done <- err
	}()

	if !waitForInflight(f, 3, 2_000) {
		t.Fatalf("inflight never reached >=3 (peak=%d)", f.peak.Load())
	}
	close(hold)

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

// waitForInflight busy-waits up to maxMs for f.peak to reach target.
func waitForInflight(f *fakeGCS, target int32, maxMs int) bool {
	for i := 0; i < maxMs; i++ {
		if f.peak.Load() >= target {
			return true
		}
		ch := make(chan struct{}, 1)
		go func() { ch <- struct{}{} }()
		<-ch
	}
	return false
}

func TestSync_DeleteRemovesStaleKeys(t *testing.T) {
	dir := writeFiles(t, map[string]string{"keep.txt": "x"})
	f := &fakeGCS{objects: map[string]RemoteObject{
		"keep.txt":  {Name: "keep.txt", CRC32C: crc("x")},
		"stale.txt": {Name: "stale.txt", CRC32C: 0x9},
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
		t.Errorf("Delete calls = %v, want [stale.txt]", f.deletes)
	}
}

func TestSync_PropagatesError(t *testing.T) {
	dir := writeFiles(t, map[string]string{"f.txt": "x"})
	stub := errors.New("permission denied")
	f := &fakeGCS{primedErr: stub}

	_, _, _, err := Sync(context.Background(), f, Options{
		Bucket: "b", Dir: dir, Concurrency: 1,
	})
	if !errors.Is(err, stub) {
		t.Errorf("err = %v, want %v", err, stub)
	}
}

// CRC32C match means no re-upload — this is the load-bearing GCS-specific test.
// A naive implementation using the IEEE polynomial (the default if you forget
// to pass castagnoli) will compute a different value than the server's and
// re-upload every file every run.
func TestSync_CRC32CMatchSkipsUpload(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"a.txt": "hello",
		"b.txt": "world",
	})
	f := &fakeGCS{objects: map[string]RemoteObject{
		"a.txt": {Name: "a.txt", CRC32C: crc("hello")},
		"b.txt": {Name: "b.txt", CRC32C: crc("world")},
	}}

	up, del, skip, err := Sync(context.Background(), f, Options{
		Bucket: "b", Dir: dir, Concurrency: 2,
	})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if up != 0 || del != 0 || skip != 2 {
		t.Errorf("counts = up:%d del:%d skip:%d, want 0/0/2 (CRC matches → skip everything)", up, del, skip)
	}
	if len(f.uploads) != 0 {
		t.Errorf("Upload calls = %v, want 0 (CRC matches)", f.uploads)
	}
}

// Sanity check the helper we use everywhere else in the suite.
func TestCRCHelperUsesCastagnoli(t *testing.T) {
	const s = "hello"
	want := crc32.Checksum([]byte(s), castagnoli)
	got := crc(s)
	if got != want {
		t.Fatalf("crc(%q) = %d, want %d (Castagnoli polynomial)", s, got, want)
	}
	if got == crc32.ChecksumIEEE([]byte(s)) {
		t.Fatalf("crc(%q) matched the IEEE polynomial — should use Castagnoli", s)
	}
}

// computeCRC32C round-trips bytes via io.Reader (used by WalkLocal under
// the hood). Tests that the implementation hashes the right polynomial when
// it pulls from a stream.
func TestComputeCRC32C_StreamMatchesChecksum(t *testing.T) {
	body := []byte("the quick brown fox")
	want := crc32.Checksum(body, castagnoli)
	got, err := computeCRC32C(bytes.NewReader(body))
	if err != nil {
		t.Fatalf("computeCRC32C: %v", err)
	}
	if got != want {
		t.Errorf("computeCRC32C = %d, want %d", got, want)
	}
}
