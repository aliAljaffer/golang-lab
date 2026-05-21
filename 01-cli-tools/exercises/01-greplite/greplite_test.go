//go:build exercise

package greplite

import (
	"strings"
	"testing"
)

const sample = `apple
Banana
cherry
APPLE pie
`

func TestGrep_BasicMatch(t *testing.T) {
	got, err := Grep(strings.NewReader(sample), "apple", Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Text != "apple" {
		t.Errorf("got %+v, want one match for 'apple'", got)
	}
}

func TestGrep_IgnoreCase(t *testing.T) {
	got, err := Grep(strings.NewReader(sample), "apple", Options{IgnoreCase: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Errorf("got %d matches, want 2 (apple + APPLE pie)", len(got))
	}
}

func TestGrep_LineNumbers(t *testing.T) {
	got, err := Grep(strings.NewReader(sample), "cherry", Options{LineNumbers: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d matches, want 1", len(got))
	}
	if got[0].LineNumber != 3 {
		t.Errorf("line number = %d, want 3", got[0].LineNumber)
	}
}

func TestGrep_EmptyPatternMatchesAll(t *testing.T) {
	got, err := Grep(strings.NewReader(sample), "", Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 {
		t.Errorf("got %d, want 4 (all lines)", len(got))
	}
}

func TestGrep_NoMatch(t *testing.T) {
	got, err := Grep(strings.NewReader(sample), "zzz", Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("got %d matches, want 0", len(got))
	}
}
