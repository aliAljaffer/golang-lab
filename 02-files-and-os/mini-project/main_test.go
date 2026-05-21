//go:build exercise

package main

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"
)

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRotateOnce_NoPriorRotation(t *testing.T) {
	dir := t.TempDir()
	log := filepath.Join(dir, "app.log")
	writeFile(t, log, []byte("line one\nline two\n"))

	if err := rotateOnce(log); err != nil {
		t.Fatalf("rotateOnce: %v", err)
	}

	// app.log should exist and be empty.
	got, err := os.ReadFile(log)
	if err != nil {
		t.Fatalf("read fresh log: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("fresh log not empty: %q", got)
	}

	// app.log.1 should hold the old content.
	old, err := os.ReadFile(log + ".1")
	if err != nil {
		t.Fatalf("read rotated: %v", err)
	}
	if string(old) != "line one\nline two\n" {
		t.Errorf("rotated content mismatch: %q", old)
	}
}

func TestRotateOnce_SecondRotationGzips(t *testing.T) {
	dir := t.TempDir()
	log := filepath.Join(dir, "app.log")
	writeFile(t, log, []byte("current\n"))
	writeFile(t, log+".1", []byte("previous\n"))

	if err := rotateOnce(log); err != nil {
		t.Fatalf("rotateOnce: %v", err)
	}

	// app.log.1 should now hold "current".
	if b, _ := os.ReadFile(log + ".1"); string(b) != "current\n" {
		t.Errorf(".1 content = %q, want %q", b, "current\n")
	}

	// app.log.2.gz should exist and decompress to "previous".
	f, err := os.Open(log + ".2.gz")
	if err != nil {
		t.Fatalf("open .2.gz: %v", err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatalf("gzip.NewReader: %v", err)
	}
	defer gz.Close()
	out, err := io.ReadAll(gz)
	if err != nil {
		t.Fatalf("read gz: %v", err)
	}
	if string(out) != "previous\n" {
		t.Errorf("gz content = %q, want %q", out, "previous\n")
	}
}

func TestGzipFile_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "src.txt.gz")
	want := []byte("hello world, this compresses\n")
	writeFile(t, src, want)

	if err := gzipFile(src, dst); err != nil {
		t.Fatalf("gzipFile: %v", err)
	}

	f, err := os.Open(dst)
	if err != nil {
		t.Fatalf("open dst: %v", err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatalf("gzip.NewReader: %v", err)
	}
	defer gz.Close()
	got, err := io.ReadAll(gz)
	if err != nil {
		t.Fatalf("read gz: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("round-trip mismatch: %q vs %q", got, want)
	}
}

func TestPruneOld_RemovesOnlyOldGz(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)

	mk := func(name string, mtime time.Time) {
		p := filepath.Join(dir, name)
		writeFile(t, p, []byte("x"))
		if err := os.Chtimes(p, mtime, mtime); err != nil {
			t.Fatal(err)
		}
	}

	mk("app.log.2.gz", now.AddDate(0, 0, -10)) // old, should die
	mk("app.log.3.gz", now.AddDate(0, 0, -2))  // recent, keep
	mk("app.log.4.gz", now.AddDate(0, 0, -20)) // old, should die
	mk("other.log.5.gz", now.AddDate(0, 0, -30)) // wrong prefix, keep
	mk("app.log.1", now.AddDate(0, 0, -30))    // wrong suffix, keep

	removed, err := pruneOld(dir, "app.log.", 7, now)
	if err != nil {
		t.Fatalf("pruneOld: %v", err)
	}
	sort.Strings(removed)
	if len(removed) != 2 {
		t.Fatalf("removed %d files, want 2: %v", len(removed), removed)
	}

	// Verify the survivors.
	for _, keep := range []string{"app.log.3.gz", "other.log.5.gz", "app.log.1"} {
		if _, err := os.Stat(filepath.Join(dir, keep)); err != nil {
			t.Errorf("%s should still exist: %v", keep, err)
		}
	}
	for _, gone := range []string{"app.log.2.gz", "app.log.4.gz"} {
		if _, err := os.Stat(filepath.Join(dir, gone)); !os.IsNotExist(err) {
			t.Errorf("%s should be gone, err=%v", gone, err)
		}
	}
}

func TestPruneOld_KeepDaysZeroIsNoOp(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	writeFile(t, filepath.Join(dir, "app.log.2.gz"), []byte("x"))

	removed, err := pruneOld(dir, "app.log.", 0, now)
	if err != nil {
		t.Fatalf("pruneOld: %v", err)
	}
	if len(removed) != 0 {
		t.Errorf("keep-days=0 should remove nothing, got %v", removed)
	}
}
