package strs

import "testing"

// TestRepeat shows the upgrade from a plain table loop to `t.Run` subtests.
//
// Each call to `t.Run(name, fn)` creates a subtest. With `go test -v` you see
// PASS/FAIL per case. With `-run`, you can target a single case across the
// whole test binary:
//
//   go test -v -run TestRepeat/single ./06-testing/03-subtests/...
//
// `t.Run` also enables parallelism: `tt.Parallel()` inside the subtest body
// lets the runner schedule cases concurrently within the same parent test.
func TestRepeat(t *testing.T) {
	tests := []struct {
		name string
		s    string
		n    int
		want string
	}{
		{"zero", "ab", 0, ""},
		{"single", "ab", 1, "ab"},
		{"twice", "ab", 2, "abab"},
		{"negative clamps to zero", "ab", -3, ""},
		{"empty input", "", 5, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Repeat(tc.s, tc.n)
			if got != tc.want {
				t.Errorf("Repeat(%q, %d) = %q, want %q", tc.s, tc.n, got, tc.want)
			}
		})
	}
}

// TODO: try `go test -run TestRepeat/zero ./06-testing/03-subtests/...` — only the "zero" subtest should run.
// TODO: add `t.Parallel()` as the first line of each subtest and re-run with `-v`. Output order will scramble — that's the giveaway they're concurrent.
