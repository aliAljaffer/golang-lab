package csv

import (
	"os"
	"path/filepath"
	"testing"
)

// fixture opens a file in testdata/. `testdata` is a magic dir name — the Go
// build tool excludes it from package builds (so non-.go files don't trigger
// errors), but it's just a normal directory at test time.
func fixture(t *testing.T, name string) *os.File {
	t.Helper() // makes t.Fatal/t.Error blame the CALLER's line, not this line
	f, err := os.Open(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	t.Cleanup(func() { f.Close() }) // runs after the test, even on failure
	return f
}

func TestParse_GoodFixture(t *testing.T) {
	rows, err := Parse(fixture(t, "good.csv"))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("len(rows) = %d, want 3", len(rows))
	}
	if rows[0].Name != "ali" || rows[0].Score != 42 {
		t.Errorf("rows[0] = %+v", rows[0])
	}
	if rows[2].Name != "omar" || rows[2].Score != 7 {
		t.Errorf("rows[2] = %+v", rows[2])
	}
}

func TestParse_BadScoreFails(t *testing.T) {
	if _, err := Parse(fixture(t, "badscore.csv")); err == nil {
		t.Fatal("expected error for non-numeric score, got nil")
	}
}

// TODO: add a fixture testdata/empty.csv (just the header, no rows) and a test that asserts len(rows) == 0.
// TODO: add a fixture testdata/badheader.csv with wrong column names and a test that asserts an error.
