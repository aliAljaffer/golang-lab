package strs

import "testing"

// TestReverse is one test that covers many cases via a table.
//
// The struct shape is local — `tests := []struct{...}{...}` — because it is
// only used here. No need to name the type unless multiple tests share it.
func TestReverse(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"ascii", "hello", "olleh"},
		{"single char", "x", "x"},
		{"palindrome", "racecar", "racecar"},
		{"unicode", "héllo", "olléh"},
		// TODO: add a case for the empty string.
		// TODO: add a case for a string with a multi-byte emoji like "✨a✨".
	}

	for _, tc := range tests {
		got := Reverse(tc.input)
		if got != tc.want {
			t.Errorf("%s: Reverse(%q) = %q, want %q", tc.name, tc.input, got, tc.want)
		}
	}
}
