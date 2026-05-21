package classify

import "testing"

// These five separate tests are the "before" state. They all pass.
// The exercise is to collapse them into a single table-driven test in
// classify_table_test.go (gated by //go:build exercise). Once your table
// test covers everything below, delete this file.

func TestClassify_A(t *testing.T) {
	if got := Classify(95); got != "A" {
		t.Errorf("Classify(95) = %q, want A", got)
	}
}

func TestClassify_B(t *testing.T) {
	if got := Classify(85); got != "B" {
		t.Errorf("Classify(85) = %q, want B", got)
	}
}

func TestClassify_C(t *testing.T) {
	if got := Classify(75); got != "C" {
		t.Errorf("Classify(75) = %q, want C", got)
	}
}

func TestClassify_F(t *testing.T) {
	if got := Classify(40); got != "F" {
		t.Errorf("Classify(40) = %q, want F", got)
	}
}

func TestClassify_OutOfRange(t *testing.T) {
	if got := Classify(150); got != "?" {
		t.Errorf("Classify(150) = %q, want ?", got)
	}
}
