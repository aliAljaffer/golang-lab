// 07-benchmark — two implementations of the same function, side by side, to
// answer "is the rewrite actually faster?"
//
// `go test -bench=.` runs anything named `BenchmarkXxx`. The runner re-runs
// each benchmark with growing `b.N` until it has a stable per-op number.
package sum

// SumLoop uses a classic indexed for loop.
func SumLoop(ns []int) int {
	s := 0
	for i := 0; i < len(ns); i++ {
		s += ns[i]
	}
	return s
}

// SumRange uses for-range. Compilers can sometimes optimize this differently;
// the benchmark will tell us.
func SumRange(ns []int) int {
	s := 0
	for _, n := range ns {
		s += n
	}
	return s
}
