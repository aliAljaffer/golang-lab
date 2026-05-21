//go:build exercise

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ---------- TestMain: package-level setup/teardown ----------

// TestMain runs once for the whole test binary. Use it for one-shot setup
// (e.g. opening a shared DB connection, creating a tmpdir).
// Here we just demonstrate the shape — record start time, log total duration.
//
// `m.Run()` runs all the tests; its return code is what we exit with.
func TestMain(m *testing.M) {
	start := time.Now()
	code := m.Run()
	fmt.Fprintf(os.Stderr, "[TestMain] suite ran in %v\n", time.Since(start).Round(time.Millisecond))
	os.Exit(code)
}

// ---------- 02 table-driven + 03 subtests: Parse ----------

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    Entry
		wantErr bool
	}{
		{"info", "[INFO] hello", Entry{Level: LevelInfo, Body: "hello"}, false},
		{"debug empty body", "[DEBUG]", Entry{Level: LevelDebug, Body: ""}, false},
		{"warn with extra spaces", "  [WARN] something", Entry{Level: LevelWarn, Body: "something"}, false},
		{"error long body", "[ERROR] db connection refused at 10.0.0.5", Entry{Level: LevelError, Body: "db connection refused at 10.0.0.5"}, false},
		{"unknown level", "[TRACE] hi", Entry{}, true},
		{"missing bracket", "INFO hello", Entry{}, true},
		{"empty string", "", Entry{}, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.in)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Parse(%q) err=%v, wantErr=%v", tc.in, err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}
			if got != tc.want {
				t.Errorf("Parse(%q) = %+v, want %+v", tc.in, got, tc.want)
			}
		})
	}
}

// ---------- 01 basic + 03 subtests: FormatRate ----------

func TestFormatRate(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		got := FormatRate(100, 2*time.Second)
		if got != "50.0 events/s" {
			t.Errorf("got %q, want %q", got, "50.0 events/s")
		}
	})

	t.Run("zero duration is safe", func(t *testing.T) {
		got := FormatRate(100, 0)
		if got != "0.0 events/s" {
			t.Errorf("got %q, want %q (must not divide by zero)", got, "0.0 events/s")
		}
	})

	t.Run("fractional rate", func(t *testing.T) {
		got := FormatRate(1, 4*time.Second)
		if got != "0.2 events/s" {
			t.Errorf("got %q, want %q", got, "0.2 events/s")
		}
	})
}

// ---------- Aggregator: stateful, uses t.Helper ----------

// addAll is a helper. t.Helper() makes assertion failures point at the test
// that called addAll, not at the addAll body.
func addAll(t *testing.T, a *Aggregator, entries ...Entry) {
	t.Helper()
	for _, e := range entries {
		a.Add(e)
	}
}

func TestAggregator_CountsPerLevel(t *testing.T) {
	var a Aggregator
	addAll(t, &a,
		Entry{Level: LevelInfo},
		Entry{Level: LevelInfo},
		Entry{Level: LevelError},
		Entry{Level: LevelWarn},
	)

	snap := a.Snapshot()
	if snap.Total != 4 {
		t.Errorf("Total = %d, want 4", snap.Total)
	}
	if snap.ByLevel[LevelInfo] != 2 {
		t.Errorf("ByLevel[INFO] = %d, want 2", snap.ByLevel[LevelInfo])
	}
	if snap.ByLevel[LevelError] != 1 {
		t.Errorf("ByLevel[ERROR] = %d, want 1", snap.ByLevel[LevelError])
	}
}

func TestAggregator_DurationGrows(t *testing.T) {
	var a Aggregator
	a.Add(Entry{Level: LevelInfo})
	time.Sleep(10 * time.Millisecond)
	a.Add(Entry{Level: LevelInfo})

	snap := a.Snapshot()
	if snap.Duration < 5*time.Millisecond {
		t.Errorf("Duration = %v, want >= 5ms", snap.Duration)
	}
}

// ---------- 04 mock-interface: recordingSource ----------

// recordingSource is a hand-rolled fake Source. The test injects canned bytes
// and asserts they get parsed + counted.
type recordingSource struct {
	body    string
	fetched int
	failErr error
}

func (r *recordingSource) Fetch(ctx context.Context) (io.ReadCloser, error) {
	r.fetched++
	if r.failErr != nil {
		return nil, r.failErr
	}
	return io.NopCloser(strings.NewReader(r.body)), nil
}

func TestSummarize_WithMockSource(t *testing.T) {
	src := &recordingSource{body: "[INFO] one\n[WARN] two\n[INFO] three\n"}

	stats, err := Summarize(context.Background(), src)
	if err != nil {
		t.Fatalf("Summarize: %v", err)
	}
	if src.fetched != 1 {
		t.Errorf("src.Fetch called %d times, want 1", src.fetched)
	}
	if stats.Total != 3 {
		t.Errorf("Total = %d, want 3", stats.Total)
	}
	if stats.ByLevel[LevelInfo] != 2 || stats.ByLevel[LevelWarn] != 1 {
		t.Errorf("ByLevel = %+v, want {INFO: 2, WARN: 1}", stats.ByLevel)
	}
}

func TestSummarize_PropagatesFetchError(t *testing.T) {
	want := errors.New("upstream down")
	src := &recordingSource{failErr: want}

	_, err := Summarize(context.Background(), src)
	if !errors.Is(err, want) {
		t.Errorf("err = %v, want %v", err, want)
	}
}

func TestSummarize_SkipsMalformedLines(t *testing.T) {
	src := &recordingSource{body: "[INFO] keep\nnot a valid line\n[ERROR] keep\n"}

	stats, err := Summarize(context.Background(), src)
	if err != nil {
		t.Fatalf("Summarize: %v", err)
	}
	if stats.Total != 2 {
		t.Errorf("Total = %d, want 2 (malformed line should be skipped)", stats.Total)
	}
}

// ---------- 06 testdata: FileSource ----------

func TestFileSource_Fetch(t *testing.T) {
	path := filepath.Join("testdata", "lines.log")
	src := FileSource{Path: path}

	stats, err := Summarize(context.Background(), src)
	if err != nil {
		t.Fatalf("Summarize: %v", err)
	}
	if stats.Total != 10 {
		t.Errorf("Total = %d, want 10 (testdata/lines.log has 10 lines)", stats.Total)
	}
	if stats.ByLevel[LevelInfo] != 5 {
		t.Errorf("ByLevel[INFO] = %d, want 5", stats.ByLevel[LevelInfo])
	}
	if stats.ByLevel[LevelError] != 2 {
		t.Errorf("ByLevel[ERROR] = %d, want 2", stats.ByLevel[LevelError])
	}
}

// ---------- 05 httptest: HTTPSource ----------

func TestHTTPSource_Fetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "[INFO] from server")
		fmt.Fprintln(w, "[DEBUG] from server")
		fmt.Fprintln(w, "[INFO] from server")
	}))
	t.Cleanup(srv.Close)

	src := HTTPSource{URL: srv.URL, Client: srv.Client()}
	stats, err := Summarize(context.Background(), src)
	if err != nil {
		t.Fatalf("Summarize: %v", err)
	}
	if stats.Total != 3 {
		t.Errorf("Total = %d, want 3", stats.Total)
	}
	if stats.ByLevel[LevelInfo] != 2 || stats.ByLevel[LevelDebug] != 1 {
		t.Errorf("ByLevel = %+v, want {INFO: 2, DEBUG: 1}", stats.ByLevel)
	}
}

func TestHTTPSource_Non200IsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(503)
	}))
	t.Cleanup(srv.Close)

	src := HTTPSource{URL: srv.URL, Client: srv.Client()}
	if _, err := Summarize(context.Background(), src); err == nil {
		t.Fatal("expected error on 503, got nil")
	}
}

// ---------- 07 benchmark: Parse ----------

func BenchmarkParse_1k(b *testing.B) {
	// Generate 1000 representative lines.
	lines := make([]string, 1000)
	for i := range lines {
		switch i % 4 {
		case 0:
			lines[i] = "[INFO] request handled in 12ms"
		case 1:
			lines[i] = "[DEBUG] config refresh"
		case 2:
			lines[i] = "[WARN] queue depth 90% full"
		case 3:
			lines[i] = "[ERROR] timeout calling upstream"
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, line := range lines {
			_, _ = Parse(line)
		}
	}
}

// ---------- 08 fuzz: Parse ----------

// FuzzParse seeds the fuzzer with valid lines and asserts two invariants:
//   1. Parse must not panic on any input.
//   2. If Parse returns no error, the Level must be one of KnownLevels.
//
// Run interactively:
//   go test -tags=exercise -fuzz=FuzzParse -fuzztime=10s ./06-testing/mini-project/...
func FuzzParse(f *testing.F) {
	for _, seed := range []string{
		"[INFO] hello",
		"[ERROR] db down",
		"[DEBUG]",
		"",
		"[TRACE] nope",
		"]ok",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, in string) {
		e, err := Parse(in)
		if err != nil {
			return
		}
		known := false
		for _, l := range KnownLevels {
			if e.Level == l {
				known = true
				break
			}
		}
		if !known {
			t.Errorf("Parse(%q) returned no error but Level=%q is not a known level",
				in, e.Level)
		}
	})
}
