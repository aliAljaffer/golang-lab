//go:build exercise

package tailf

import (
	"os"
	"path/filepath"
	"testing"
)

func openSeed(t *testing.T, content string) (*os.File, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "log.txt")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { f.Close() })
	return f, path
}

func TestReadAppend_FirstReadReturnsAll(t *testing.T) {
	f, _ := openSeed(t, "hello\n")
	got, n, err := ReadAppend(f, 0)
	if err != nil {
		t.Fatalf("ReadAppend: %v", err)
	}
	if string(got) != "hello\n" {
		t.Errorf("bytes = %q, want %q", got, "hello\n")
	}
	if n != 6 {
		t.Errorf("size = %d, want 6", n)
	}
}

func TestReadAppend_NoGrowthReturnsEmpty(t *testing.T) {
	f, _ := openSeed(t, "hello\n")
	got, n, err := ReadAppend(f, 6)
	if err != nil {
		t.Fatalf("ReadAppend: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("bytes = %q, want empty", got)
	}
	if n != 6 {
		t.Errorf("size = %d, want 6", n)
	}
}

func TestReadAppend_AppendedSinceLastCall(t *testing.T) {
	f, path := openSeed(t, "first\n")
	_, n1, err := ReadAppend(f, 0)
	if err != nil {
		t.Fatalf("first read: %v", err)
	}

	// Append to the file out-of-band.
	w, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("second\n")); err != nil {
		t.Fatal(err)
	}
	w.Close()

	got, n2, err := ReadAppend(f, n1)
	if err != nil {
		t.Fatalf("second read: %v", err)
	}
	if string(got) != "second\n" {
		t.Errorf("delta = %q, want %q", got, "second\n")
	}
	if n2 != n1+int64(len("second\n")) {
		t.Errorf("new size = %d, want %d", n2, n1+int64(len("second\n")))
	}
}

func TestReadAppend_TruncationErrors(t *testing.T) {
	f, path := openSeed(t, "hello world\n")
	if err := os.Truncate(path, 4); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ReadAppend(f, 12); err == nil {
		t.Fatal("expected error on shrunk file, got nil")
	}
}
