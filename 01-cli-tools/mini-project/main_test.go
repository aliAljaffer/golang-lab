//go:build exercise

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fixture builds:
//   root/
//     a/file1 (10 bytes)
//     a/file2 (20 bytes)        -> a totals 30
//     b/sub/file3 (100 bytes)   -> b totals 100
//     c/                         -> c totals 0
func fixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	mkfile := func(rel string, n int) {
		full := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, make([]byte, n), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mkfile("a/file1", 10)
	mkfile("a/file2", 20)
	mkfile("b/sub/file3", 100)
	if err := os.MkdirAll(filepath.Join(root, "c"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestScan_SumsRecursively(t *testing.T) {
	root := fixture(t)
	entries, err := scan(root)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	got := map[string]int64{}
	for _, e := range entries {
		got[filepath.Base(e.Path)] = e.Bytes
	}
	want := map[string]int64{"a": 30, "b": 100, "c": 0}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("dir %q: got %d bytes, want %d", k, got[k], v)
		}
	}
}

func TestScan_MissingPathErrors(t *testing.T) {
	if _, err := scan(filepath.Join(t.TempDir(), "does-not-exist")); err == nil {
		t.Fatal("expected error for missing path, got nil")
	}
}

func TestSortAndTrim_DescendingThenTop(t *testing.T) {
	in := []Entry{
		{Path: "a", Bytes: 30},
		{Path: "b", Bytes: 100},
		{Path: "c", Bytes: 0},
	}
	out := sortAndTrim(in, 2)

	if len(out) != 2 {
		t.Fatalf("len=%d, want 2", len(out))
	}
	if out[0].Path != "b" || out[1].Path != "a" {
		t.Errorf("order = %v, want [b a]", out)
	}
}

func TestSortAndTrim_TopZeroReturnsAll(t *testing.T) {
	in := []Entry{{Path: "a", Bytes: 1}, {Path: "b", Bytes: 2}}
	out := sortAndTrim(in, 0)
	if len(out) != 2 {
		t.Errorf("len=%d, want 2", len(out))
	}
}

func TestRenderJSON_Parseable(t *testing.T) {
	entries := []Entry{{Path: "a", Bytes: 30}, {Path: "b", Bytes: 100}}
	s, err := renderJSON(entries)
	if err != nil {
		t.Fatalf("renderJSON: %v", err)
	}

	var back []Entry
	if err := json.Unmarshal([]byte(s), &back); err != nil {
		t.Fatalf("output is not valid JSON: %v\ngot: %s", err, s)
	}
	if len(back) != 2 || back[0].Path != "a" || back[1].Bytes != 100 {
		t.Errorf("roundtrip mismatch: %+v", back)
	}
}

func TestRenderText_ContainsAllPaths(t *testing.T) {
	entries := []Entry{{Path: "alpha", Bytes: 30}, {Path: "beta", Bytes: 100}}
	s := renderText(entries)
	for _, want := range []string{"alpha", "beta"} {
		if !strings.Contains(s, want) {
			t.Errorf("renderText output missing %q\ngot: %s", want, s)
		}
	}
}
