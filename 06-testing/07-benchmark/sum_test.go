package sum

import "testing"

// First, a correctness test. Benchmarks measure speed; they say nothing
// about correctness. Always have a TestXxx covering the function too.
func TestSum_BothAgree(t *testing.T) {
	ns := []int{1, 2, 3, 4, 5}
	want := 15
	if got := SumLoop(ns); got != want {
		t.Errorf("SumLoop = %d, want %d", got, want)
	}
	if got := SumRange(ns); got != want {
		t.Errorf("SumRange = %d, want %d", got, want)
	}
}

// makeInput is shared by both benchmarks so they hit the same data shape.
func makeInput(n int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = i
	}
	return out
}

// BenchmarkSumLoop_10k — `b.N` is set by the runner; you loop B.N times.
// Reset the timer AFTER one-shot setup so allocation cost doesn't leak in.
func BenchmarkSumLoop_10k(b *testing.B) {
	ns := makeInput(10_000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SumLoop(ns)
	}
}

func BenchmarkSumRange_10k(b *testing.B) {
	ns := makeInput(10_000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SumRange(ns)
	}
}

// TODO: add a sub-benchmark over multiple sizes using b.Run, e.g.
//   for _, n := range []int{100, 10_000, 1_000_000} { b.Run(strconv.Itoa(n), ...) }
// to see how the gap (if any) scales with input size.
//
// TODO: try `go test -bench=. -benchmem ./06-testing/07-benchmark/...` — the
// `-benchmem` flag adds allocs/op + bytes/op columns. For SumLoop/SumRange
// both should be zero (no heap allocs in the loop).
