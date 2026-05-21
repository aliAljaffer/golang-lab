//go:build exercise

package classify

import "testing"

// TestClassify_Table is the target shape. Fill in `tests` with cases that
// cover every branch in Classify (including the boundaries — 89 vs 90, 0,
// 100, negative, 101). Use t.Run so each case is its own named subtest.
//
// Goal: cover every case from classify_separate_test.go AND each boundary,
// in ONE function. Then delete classify_separate_test.go.
func TestClassify_Table(t *testing.T) {
	tests := []struct {
		name  string
		score int
		want  string
	}{
		// TODO: fill in cases. At minimum cover:
		//   - the A range (90, 100)
		//   - the B range (80, 89)
		//   - the C range (70, 79)
		//   - the D range (60, 69)
		//   - the F range (0, 59)
		//   - out-of-range (-1, 101)
	}

	if len(tests) == 0 {
		t.Fatal("TODO: add cases — start with the A-grade boundary (90 vs 89)")
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := Classify(tc.score); got != tc.want {
				t.Errorf("Classify(%d) = %q, want %q", tc.score, got, tc.want)
			}
		})
	}
}
