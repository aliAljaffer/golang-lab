//go:build exercise

package dirdiff

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func mkfile(t *testing.T, root, rel string, data []byte) {
	t.Helper()
	full := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDiff_IdenticalTreesEmpty(t *testing.T) {
	l := t.TempDir()
	r := t.TempDir()
	mkfile(t, l, "a.txt", []byte("same"))
	mkfile(t, r, "a.txt", []byte("same"))

	got, err := Diff(l, r)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty diff, got %+v", got)
	}
}

func TestDiff_OnlyLeftAndOnlyRight(t *testing.T) {
	l := t.TempDir()
	r := t.TempDir()
	mkfile(t, l, "left-only.txt", []byte("L"))
	mkfile(t, r, "right-only.txt", []byte("R"))

	got, err := Diff(l, r)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	sort.Slice(got, func(i, j int) bool { return got[i].Path < got[j].Path })

	if len(got) != 2 {
		t.Fatalf("len=%d, want 2: %+v", len(got), got)
	}
	if got[0].Path != "left-only.txt" || got[0].Kind != OnlyLeft {
		t.Errorf("got[0] = %+v, want {left-only.txt, OnlyLeft}", got[0])
	}
	if got[1].Path != "right-only.txt" || got[1].Kind != OnlyRight {
		t.Errorf("got[1] = %+v, want {right-only.txt, OnlyRight}", got[1])
	}
}

func TestDiff_ModifiedDetected(t *testing.T) {
	l := t.TempDir()
	r := t.TempDir()
	mkfile(t, l, "shared.txt", []byte("hello"))
	mkfile(t, r, "shared.txt", []byte("HELLO"))

	got, err := Diff(l, r)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len=%d, want 1: %+v", len(got), got)
	}
	if got[0].Path != "shared.txt" || got[0].Kind != Modified {
		t.Errorf("got[0] = %+v, want {shared.txt, Modified}", got[0])
	}
}

func TestDiff_NestedPathsRelativeToRoot(t *testing.T) {
	l := t.TempDir()
	r := t.TempDir()
	mkfile(t, l, "sub/dir/file.txt", []byte("x"))

	got, err := Diff(l, r)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len=%d, want 1: %+v", len(got), got)
	}
	if got[0].Path != filepath.Join("sub", "dir", "file.txt") {
		t.Errorf("path = %q, want relative %q", got[0].Path, filepath.Join("sub", "dir", "file.txt"))
	}
	if got[0].Kind != OnlyLeft {
		t.Errorf("kind = %d, want OnlyLeft (%d)", got[0].Kind, OnlyLeft)
	}
}

func TestDiff_MissingRootErrors(t *testing.T) {
	if _, err := Diff(filepath.Join(t.TempDir(), "does-not-exist"), t.TempDir()); err == nil {
		t.Fatal("expected error for missing left root, got nil")
	}
}
