package parse

import (
	"strconv"
	"testing"
)

// Standard table test covering the obvious cases.
func TestParseInt(t *testing.T) {
	tests := []struct {
		in      string
		want    int
		wantErr bool
	}{
		{"0", 0, false},
		{"123", 123, false},
		{"-42", -42, false},
		{"", 0, true},
		{"abc", 0, true},
	}
	for _, tc := range tests {
		got, err := ParseInt(tc.in)
		if (err != nil) != tc.wantErr {
			t.Errorf("ParseInt(%q) err=%v, wantErr=%v", tc.in, err, tc.wantErr)
			continue
		}
		if !tc.wantErr && got != tc.want {
			t.Errorf("ParseInt(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

// FuzzParseInt — the fuzzer mutates the seed corpus and feeds inputs to the
// fuzz target. Invariant we assert: if our ParseInt returns no error, the
// stdlib's strconv.Atoi must agree.
//
// Run interactively to actually fuzz:
//   go test -fuzz=FuzzParseInt ./06-testing/08-fuzz/...
//
// CI / `go test` (without -fuzz) only runs the seed corpus as plain test
// cases — fast, deterministic, no random search.
func FuzzParseInt(f *testing.F) {
	// Seed corpus: inputs that should already pass. The fuzzer uses these as
	// starting points for mutation.
	for _, seed := range []string{"0", "1", "-1", "42", "999"} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, in string) {
		got, err := ParseInt(in)
		if err != nil {
			return // our parser rejected it — fine, no invariant to check
		}
		want, stdErr := strconv.Atoi(in)
		if stdErr != nil {
			// We accepted what stdlib rejected. That's the bug.
			t.Errorf("ParseInt(%q) = %d (nil err) but strconv.Atoi rejects it: %v",
				in, got, stdErr)
			return
		}
		if got != want {
			t.Errorf("ParseInt(%q) = %d, strconv.Atoi = %d (disagreement)",
				in, got, want)
		}
	})
}

// TODO: run `go test -fuzz=FuzzParseInt ./06-testing/08-fuzz/...` for a few
// seconds. The fuzzer will write the failing input to
// `testdata/fuzz/FuzzParseInt/<hash>` and from then on `go test` (without
// -fuzz) will replay it as a regression test forever.
//
// TODO: fix the bug in parse.go so "-" returns ErrEmpty (or a similar error).
// Re-run the fuzzer — it should run without finding anything new.
