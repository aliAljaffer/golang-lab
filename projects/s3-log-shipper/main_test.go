//go:build exercise

package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aws/smithy-go"
)

// fixedNow at 2026-05-22T12:00:00Z — every clock-sensitive Batcher test uses
// this as the base "now" so age durations are exact.
var fixedNow = time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)

// ---- Tailer ---------------------------------------------------------------

// writeFile writes content to path, truncating any existing file.
func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// appendFile appends content to an existing file.
func appendFile(t *testing.T, path string, content string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open append %s: %v", path, err)
	}
	defer f.Close()
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("append %s: %v", path, err)
	}
}

// collect drains up to n Lines from ch or fails after timeout.
func collect(t *testing.T, ch <-chan Line, n int, timeout time.Duration) []Line {
	t.Helper()
	out := make([]Line, 0, n)
	deadline := time.After(timeout)
	for len(out) < n {
		select {
		case l := <-ch:
			out = append(out, l)
		case <-deadline:
			t.Fatalf("collect: got %d/%d lines before timeout", len(out), n)
		}
	}
	return out
}

func TestTailer_ReadsFromOffset(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "log")
	writeFile(t, path, "one\ntwo\nthree\n")

	store := &FileOffsetStore{Dir: dir}
	if err := store.Save(path, int64(len("one\n"))); err != nil {
		t.Fatalf("seed offset: %v", err)
	}

	tail := &Tailer{Path: path, Store: store, PollInterval: 10 * time.Millisecond}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ch := make(chan Line, 8)
	go func() { _ = tail.Run(ctx, ch) }()

	got := collect(t, ch, 2, 500*time.Millisecond)
	if string(got[0].Body) != "two" || string(got[1].Body) != "three" {
		t.Errorf("resumed lines = %q, %q; want \"two\", \"three\"", got[0].Body, got[1].Body)
	}
}

func TestTailer_PersistsOffsetOnEmit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "log")
	writeFile(t, path, "alpha\nbeta\n")

	store := &FileOffsetStore{Dir: dir}
	tail := &Tailer{Path: path, Store: store, PollInterval: 10 * time.Millisecond}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ch := make(chan Line, 8)
	go func() { _ = tail.Run(ctx, ch) }()

	_ = collect(t, ch, 2, 500*time.Millisecond)

	// give the tailer a moment to fsync the post-emit offset
	time.Sleep(50 * time.Millisecond)
	got, err := store.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if want := int64(len("alpha\nbeta\n")); got != want {
		t.Errorf("persisted offset = %d, want %d (== file size)", got, want)
	}
}

func TestTailer_HandlesTruncation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "log")
	writeFile(t, path, "old line\n")

	store := &FileOffsetStore{Dir: dir}
	// pretend we'd already read past the end of the (soon-to-be-truncated) file.
	if err := store.Save(path, 9999); err != nil {
		t.Fatalf("seed offset: %v", err)
	}

	tail := &Tailer{Path: path, Store: store, PollInterval: 10 * time.Millisecond}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ch := make(chan Line, 4)
	go func() { _ = tail.Run(ctx, ch) }()

	got := collect(t, ch, 1, 500*time.Millisecond)
	if string(got[0].Body) != "old line" {
		t.Errorf("after truncation, got %q; want \"old line\" (offset reset to 0)", got[0].Body)
	}
}

func TestTailer_HandlesEOFThenAppend(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "log")
	writeFile(t, path, "first\n")

	store := &FileOffsetStore{Dir: dir}
	tail := &Tailer{Path: path, Store: store, PollInterval: 10 * time.Millisecond}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ch := make(chan Line, 4)
	go func() { _ = tail.Run(ctx, ch) }()

	if got := collect(t, ch, 1, 500*time.Millisecond); string(got[0].Body) != "first" {
		t.Fatalf("first line = %q, want \"first\"", got[0].Body)
	}

	// Append AFTER EOF; the tailer's poll loop should pick it up.
	time.Sleep(30 * time.Millisecond)
	appendFile(t, path, "second\n")

	if got := collect(t, ch, 1, 500*time.Millisecond); string(got[0].Body) != "second" {
		t.Errorf("post-append line = %q, want \"second\"", got[0].Body)
	}
}

func TestTailer_HoldsPartialLineAtEOF(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "log")
	// no trailing newline — this MUST NOT be emitted yet.
	writeFile(t, path, "partial-no-newline")

	store := &FileOffsetStore{Dir: dir}
	tail := &Tailer{Path: path, Store: store, PollInterval: 10 * time.Millisecond}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ch := make(chan Line, 4)
	go func() { _ = tail.Run(ctx, ch) }()

	select {
	case l := <-ch:
		t.Fatalf("got line %q from partial input; expected the tailer to hold the buffer", l.Body)
	case <-time.After(150 * time.Millisecond):
		// good — no emit
	}

	// the missing newline arrives; now the line should fire.
	appendFile(t, path, "\n")
	select {
	case l := <-ch:
		if string(l.Body) != "partial-no-newline" {
			t.Errorf("late-completed line = %q, want %q", l.Body, "partial-no-newline")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("did not emit the line after the trailing newline arrived")
	}
}

// ---- FileOffsetStore ------------------------------------------------------

func TestFileOffsetStore_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := &FileOffsetStore{Dir: dir}
	path := filepath.Join(dir, "fake-input.log")

	if got, err := store.Load(path); err != nil || got != 0 {
		t.Fatalf("Load(missing) = (%d, %v); want (0, nil)", got, err)
	}
	if err := store.Save(path, 4242); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if got, err := store.Load(path); err != nil || got != 4242 {
		t.Errorf("Load after Save = (%d, %v); want (4242, nil)", got, err)
	}
}

// ---- Batcher --------------------------------------------------------------

// newBatcher builds a Batcher with an injectable clock the test controls.
func newBatcher(maxBytes int, maxAge time.Duration, now *time.Time) *Batcher {
	return &Batcher{
		MaxBytes: maxBytes,
		MaxAge:   maxAge,
		Hostname: "test-host",
		Now:      func() time.Time { return *now },
	}
}

func TestBatcher_SizeTrigger(t *testing.T) {
	now := fixedNow
	b := newBatcher(10, time.Hour, &now) // size trigger; age effectively disabled

	if _, ok := b.Add([]byte("hello")); ok {
		t.Fatal("size threshold should not fire after 6 raw bytes (\"hello\\n\")")
	}
	batch, ok := b.Add([]byte("world"))
	if !ok {
		t.Fatal("size threshold should fire after 12 raw bytes (\"hello\\nworld\\n\"); did not")
	}
	if batch.Count != 2 {
		t.Errorf("Count = %d, want 2", batch.Count)
	}
}

func TestBatcher_AgeTrigger(t *testing.T) {
	now := fixedNow
	b := newBatcher(1<<20, 5*time.Second, &now)

	if _, ok := b.Add([]byte("alpha")); ok {
		t.Fatal("Add returned (_, true) below size threshold")
	}
	if _, ok := b.MaybeFlushByAge(); ok {
		t.Fatal("MaybeFlushByAge fired before MaxAge elapsed")
	}
	now = now.Add(10 * time.Second)
	batch, ok := b.MaybeFlushByAge()
	if !ok {
		t.Fatal("MaybeFlushByAge did not fire after 10s with MaxAge=5s")
	}
	if batch.Count != 1 {
		t.Errorf("Count = %d, want 1", batch.Count)
	}
}

func TestBatcher_SizeBeatsTime(t *testing.T) {
	now := fixedNow
	b := newBatcher(10, time.Nanosecond, &now)

	// One Add that hits the size threshold must return ok=true on Add itself,
	// not punt to MaybeFlushByAge.
	if _, ok := b.Add([]byte("hello")); ok {
		t.Fatal("Add fired at 6 raw bytes with MaxBytes=10")
	}
	if batch, ok := b.Add([]byte("world")); !ok || batch.Count != 2 {
		t.Errorf("Add at threshold returned (%+v, %v); want size-trigger fire with Count=2", batch, ok)
	}
}

func TestBatcher_EmptyFlushIsNoop(t *testing.T) {
	now := fixedNow
	b := newBatcher(100, time.Hour, &now)

	if _, ok := b.Flush(); ok {
		t.Error("Flush on empty Batcher returned ok=true; want false")
	}
}

func TestBatcher_GzipRoundTrip(t *testing.T) {
	now := fixedNow
	b := newBatcher(8, time.Hour, &now)

	_, _ = b.Add([]byte("ab"))
	_, _ = b.Add([]byte("cd"))
	batch, ok := b.Add([]byte("ef")) // 9 bytes raw -> over 8 -> fires
	if !ok {
		t.Fatalf("expected size flush at \"ab\\ncd\\nef\\n\" (9 bytes), got ok=false")
	}

	zr, err := gzip.NewReader(bytes.NewReader(batch.Body))
	if err != nil {
		t.Fatalf("gzip.NewReader: %v", err)
	}
	plain, err := io.ReadAll(zr)
	if err != nil {
		t.Fatalf("decompress: %v", err)
	}
	if want := "ab\ncd\nef\n"; string(plain) != want {
		t.Errorf("decompressed = %q, want %q", plain, want)
	}
}

// TestBatcher_MD5MatchesGzippedBody is the load-bearing S3-specific
// regression. S3's ETag for a single-part PutObject is the md5 of the bytes
// it received — the gzipped Body — so the client-side MD5 we record must
// match an independent md5 of Body. A common bug is to hash the *raw lines*
// instead, which would not match S3's ETag and would silently break any
// integrity check downstream.
func TestBatcher_MD5MatchesGzippedBody(t *testing.T) {
	now := fixedNow
	b := newBatcher(5, time.Hour, &now)
	batch, ok := b.Add([]byte("payload"))
	if !ok {
		t.Fatalf("expected immediate flush (payload\\n is 8 bytes >= 5)")
	}

	sum := md5.Sum(batch.Body)
	want := hex.EncodeToString(sum[:])

	if batch.MD5 != want {
		t.Errorf("Batch.MD5 = %q; want %q (md5 of gzipped Body, S3 ETag parity)", batch.MD5, want)
	}

	// Negative check: md5 of the plaintext is NOT what we want here.
	plain := md5.Sum([]byte("payload\n"))
	plainHex := hex.EncodeToString(plain[:])
	if batch.MD5 == plainHex {
		t.Errorf("Batch.MD5 matches md5 of plaintext %q; should hash the gzipped Body (S3 hashes what it received)", plainHex)
	}
}

// ---- S3Uploader (retry/backoff) ------------------------------------------

// apiErr builds a smithy.APIError with the given ErrorCode for the retry tests.
func apiErr(code string) error {
	return &smithy.GenericAPIError{Code: code, Message: code}
}

// programmedPutter is a fake putter that returns the next programmed error
// on each call, tracking how many times it was invoked.
type programmedPutter struct {
	mu     sync.Mutex
	errs   []error
	calls  int
	bucket string
	key    string
	body   []byte
}

func (p *programmedPutter) putObject(_ context.Context, bucket, key string, body []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls++
	p.bucket, p.key, p.body = bucket, key, body
	if p.calls > len(p.errs) {
		return nil // exhausted programmed errors -> success
	}
	return p.errs[p.calls-1]
}

// fastSleep ignores the duration so tests don't actually wait. ctx-aware.
func fastSleep(ctx context.Context, _ time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

func TestS3Uploader_TransientThenSuccess(t *testing.T) {
	transient := apiErr("ServiceUnavailable")
	pp := &programmedPutter{errs: []error{transient, transient}} // 3rd call -> nil
	u := &S3Uploader{
		Inner:       pp,
		Bucket:      "b",
		MaxRetries:  3,
		BaseBackoff: time.Millisecond,
		Now:         time.Now,
		Sleep:       fastSleep,
	}
	if err := u.Put(context.Background(), "k", []byte("hi")); err != nil {
		t.Fatalf("Put = %v, want nil after 2 transient then success", err)
	}
	if pp.calls != 3 {
		t.Errorf("calls = %d, want 3 (2 retries + success)", pp.calls)
	}
}

func TestS3Uploader_PermanentNoRetry(t *testing.T) {
	permanent := apiErr("AccessDenied")
	pp := &programmedPutter{errs: []error{permanent}}
	u := &S3Uploader{Inner: pp, Bucket: "b", MaxRetries: 5, BaseBackoff: time.Millisecond, Now: time.Now, Sleep: fastSleep}

	err := u.Put(context.Background(), "k", []byte("hi"))
	if err == nil {
		t.Fatal("Put returned nil for a permanent error")
	}
	if pp.calls != 1 {
		t.Errorf("calls = %d on permanent error; want 1 (no retry)", pp.calls)
	}
}

func TestS3Uploader_BackoffBounded(t *testing.T) {
	transient := apiErr("SlowDown")
	// always-transient: 1 initial + MaxRetries retries == MaxRetries+1 calls total, then give up.
	pp := &programmedPutter{errs: []error{transient, transient, transient, transient, transient, transient}}
	u := &S3Uploader{Inner: pp, Bucket: "b", MaxRetries: 2, BaseBackoff: time.Millisecond, Now: time.Now, Sleep: fastSleep}

	err := u.Put(context.Background(), "k", []byte("hi"))
	if err == nil {
		t.Fatal("Put returned nil after MaxRetries+1 transient failures")
	}
	if pp.calls != 3 {
		t.Errorf("calls = %d, want 3 (1 attempt + MaxRetries=2 retries)", pp.calls)
	}
}

func TestIsTransient_Classification(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil is not transient", nil, false},
		{"ServiceUnavailable is transient", apiErr("ServiceUnavailable"), true},
		{"InternalError is transient", apiErr("InternalError"), true},
		{"RequestTimeout is transient", apiErr("RequestTimeout"), true},
		{"SlowDown is transient", apiErr("SlowDown"), true},
		{"Throttling is transient", apiErr("Throttling"), true},
		{"ThrottlingException is transient", apiErr("ThrottlingException"), true},
		{"AccessDenied is permanent", apiErr("AccessDenied"), false},
		{"NoSuchBucket is permanent", apiErr("NoSuchBucket"), false},
		{"NoSuchKey is permanent", apiErr("NoSuchKey"), false},
		{"InvalidAccessKeyId is permanent", apiErr("InvalidAccessKeyId"), false},
		{"SignatureDoesNotMatch is permanent", apiErr("SignatureDoesNotMatch"), false},
		{"context.Canceled is not transient", context.Canceled, false},
		{"plain error (unknown shape) is transient", errors.New("transport hiccup"), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsTransient(tc.err); got != tc.want {
				t.Errorf("IsTransient(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

// ---- End-to-end ----------------------------------------------------------

// captureUploader records every (key, body) Put it sees. body is decompressed
// into lines so the test can assert no line loss across batch boundaries.
type captureUploader struct {
	mu    sync.Mutex
	keys  []string
	lines []string
	puts  atomic.Int64
}

func (c *captureUploader) Put(_ context.Context, key string, body []byte) error {
	zr, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("gzip.NewReader: %w", err)
	}
	defer zr.Close()
	plain, err := io.ReadAll(zr)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.keys = append(c.keys, key)
	for _, line := range strings.Split(strings.TrimRight(string(plain), "\n"), "\n") {
		if line == "" {
			continue
		}
		c.lines = append(c.lines, line)
	}
	c.puts.Add(1)
	return nil
}

func (c *captureUploader) snapshot() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]string, len(c.lines))
	copy(out, c.lines)
	return out
}

// TestRun_EndToEnd_NoLoss — the load-bearing pipeline test. Writes N lines
// to a tmp file; Run tails, batches, and "uploads" via captureUploader. The
// test asserts every line shows up exactly once across the captured bodies,
// no matter where the batch boundaries fall.
func TestRun_EndToEnd_NoLoss(t *testing.T) {
	const n = 20
	dir := t.TempDir()
	path := filepath.Join(dir, "log")
	writeFile(t, path, "") // start empty

	store := &FileOffsetStore{Dir: dir}
	tail := &Tailer{Path: path, Store: store, PollInterval: 5 * time.Millisecond}
	batcher := &Batcher{
		MaxBytes: 30, // small enough that boundaries definitely fall between lines
		MaxAge:   30 * time.Millisecond,
		Hostname: "test-host",
		Now:      time.Now,
	}
	cap := &captureUploader{}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	done := make(chan error, 1)
	go func() {
		done <- Run(ctx, []*Tailer{tail}, batcher, cap, "prefix/", io.Discard)
	}()

	// Append lines one at a time so the tailer + batcher have to make
	// boundary decisions.
	for i := 0; i < n; i++ {
		appendFile(t, path, fmt.Sprintf("line-%02d\n", i))
		time.Sleep(3 * time.Millisecond)
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if len(cap.snapshot()) >= n {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	cancel()
	if err := <-done; err != nil {
		t.Fatalf("Run returned: %v", err)
	}

	got := cap.snapshot()
	if len(got) != n {
		t.Fatalf("uploaded line count = %d, want %d. got=%v", len(got), n, got)
	}
	for i, line := range got {
		want := fmt.Sprintf("line-%02d", i)
		if line != want {
			t.Errorf("line %d = %q, want %q", i, line, want)
		}
	}
	if cap.puts.Load() < 2 {
		t.Errorf("only %d Put call(s); want ≥ 2 (the test config forces multiple batches)", cap.puts.Load())
	}
}
