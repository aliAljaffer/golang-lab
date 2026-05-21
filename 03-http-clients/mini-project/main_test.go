//go:build exercise

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// newClient returns a client with a tight timeout so a hung handler in a test
// can't stall the whole suite.
func newClient() *http.Client {
	return &http.Client{Timeout: 5 * time.Second}
}

func TestFetchStats_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/golang/go" {
			t.Errorf("unexpected path: %q", r.URL.Path)
		}
		w.Header().Set("ETag", `W/"abc123"`)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"full_name":"golang/go","stargazers_count":120000,"forks_count":17000,"pushed_at":"2026-05-19T12:00:00Z"}`)
	}))
	defer srv.Close()

	stats, etag, fresh, err := fetchStats(newClient(), srv.URL, "golang/go", "")
	if err != nil {
		t.Fatalf("fetchStats: %v", err)
	}
	if !fresh {
		t.Errorf("fresh = false, want true (server returned 200)")
	}
	if etag != `W/"abc123"` {
		t.Errorf("etag = %q, want %q", etag, `W/"abc123"`)
	}
	if stats.Name != "golang/go" || stats.Stars != 120000 || stats.Forks != 17000 {
		t.Errorf("stats = %+v, fields decoded wrong", stats)
	}
}

func TestFetchStats_RetriesOn503(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("ETag", `"ok"`)
		_, _ = fmt.Fprint(w, `{"full_name":"x/y","stargazers_count":1,"forks_count":1,"pushed_at":"2026-01-01T00:00:00Z"}`)
	}))
	defer srv.Close()

	stats, _, fresh, err := fetchStats(newClient(), srv.URL, "x/y", "")
	if err != nil {
		t.Fatalf("fetchStats: %v (calls=%d)", err, calls)
	}
	if !fresh || stats.Name != "x/y" {
		t.Errorf("after retry: stats=%+v fresh=%v", stats, fresh)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Errorf("calls = %d, want 3 (two 503s then a 200)", got)
	}
}

func TestFetchStats_RetriesOn429(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = fmt.Fprint(w, `{"full_name":"x/y","stargazers_count":2,"forks_count":2,"pushed_at":"2026-01-01T00:00:00Z"}`)
	}))
	defer srv.Close()

	if _, _, _, err := fetchStats(newClient(), srv.URL, "x/y", ""); err != nil {
		t.Fatalf("fetchStats: %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("calls = %d, want 2 (one 429 then a 200)", got)
	}
}

func TestFetchStats_HonorsIfNoneMatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("If-None-Match") != `W/"cached"` {
			t.Errorf("If-None-Match header = %q, want %q",
				r.Header.Get("If-None-Match"), `W/"cached"`)
		}
		w.WriteHeader(http.StatusNotModified)
	}))
	defer srv.Close()

	stats, etag, fresh, err := fetchStats(newClient(), srv.URL, "x/y", `W/"cached"`)
	if err != nil {
		t.Fatalf("fetchStats: %v", err)
	}
	if fresh {
		t.Errorf("fresh = true on 304, want false")
	}
	if etag != `W/"cached"` {
		t.Errorf("etag should be preserved on 304, got %q", etag)
	}
	if stats != (Stats{}) {
		t.Errorf("stats on 304 should be zero value, got %+v", stats)
	}
}

func TestWriteCSV_Schema(t *testing.T) {
	var buf bytes.Buffer
	rows := []Stats{
		{Name: "golang/go", Stars: 100, Forks: 20, PushedAt: "2026-05-19T12:00:00Z"},
		{Name: "spf13/cobra", Stars: 50, Forks: 5, PushedAt: "2026-05-10T08:30:00Z"},
	}
	if err := writeCSV(&buf, rows); err != nil {
		t.Fatalf("writeCSV: %v", err)
	}
	got := buf.String()
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3 (header + 2 rows):\n%s", len(lines), got)
	}
	if lines[0] != "name,stars,forks,pushed_at" {
		t.Errorf("header = %q, want %q", lines[0], "name,stars,forks,pushed_at")
	}
	if lines[1] != "golang/go,100,20,2026-05-19T12:00:00Z" {
		t.Errorf("row 1 = %q", lines[1])
	}
	if lines[2] != "spf13/cobra,50,5,2026-05-10T08:30:00Z" {
		t.Errorf("row 2 = %q", lines[2])
	}
}

func TestCacheRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.json")

	// Missing file should not be an error.
	got, err := loadCache(path)
	if err != nil {
		t.Fatalf("loadCache on missing file: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("missing cache should be empty, got %+v", got)
	}

	want := map[string]CacheEntry{
		"golang/go": {
			ETag:  `W/"abc"`,
			Stats: Stats{Name: "golang/go", Stars: 1, Forks: 2, PushedAt: "2026-05-19T12:00:00Z"},
		},
	}
	if err := saveCache(path, want); err != nil {
		t.Fatalf("saveCache: %v", err)
	}

	got, err = loadCache(path)
	if err != nil {
		t.Fatalf("loadCache after save: %v", err)
	}
	gotJSON, _ := json.Marshal(got)
	wantJSON, _ := json.Marshal(want)
	if string(gotJSON) != string(wantJSON) {
		t.Errorf("round-trip mismatch:\ngot:  %s\nwant: %s", gotJSON, wantJSON)
	}
}
