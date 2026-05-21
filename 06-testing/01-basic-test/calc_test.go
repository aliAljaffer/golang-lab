package calc

import "testing"

// TestAdd_HappyPath shows the simplest shape: name starts with `Test`, takes
// `*testing.T`. `t.Errorf` reports a failure and continues; `t.Fatalf` reports
// and aborts THIS test (other tests in the file keep running either way).
func TestAdd_HappyPath(t *testing.T) {
	got := Add(2, 3)
	if got != 5 {
		t.Errorf("Add(2, 3) = %d, want 5", got)
	}
}

// TestAdd_Negatives — t.Errorf instead of t.Fatalf because subsequent
// assertions in this test would still make sense to run.
func TestAdd_Negatives(t *testing.T) {
	if got := Add(-2, -3); got != -5 {
		t.Errorf("Add(-2, -3) = %d, want -5", got)
	}
	if got := Add(-2, 3); got != 1 {
		t.Errorf("Add(-2, 3) = %d, want 1", got)
	}
}

// TODO: write TestSub_HappyPath and TestSub_Negatives following the shape above.
// TODO: try `go test -v ./06-testing/01-basic-test/...` to see each test name.
// TODO: introduce a bug in Sub (e.g. return a + b) and watch the test fail.
