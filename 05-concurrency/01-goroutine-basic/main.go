// 01-goroutine-basic — `go f()` and why a missing WaitGroup loses output.
//
// Goal: prove that goroutines run concurrently with main, and that main exiting
// kills them. Then fix it with sync.WaitGroup.
//
// Run:
//   go run .
//
// Notes for veterans of Python/Node/Java:
//   - `go f()` is fire-and-forget — there is no Promise/Future returned.
//   - When `main` returns, the whole program exits, killing every goroutine.
//   - That's why naive examples "miss" their output: main wins the race.
package main

import (
	"fmt"
	"sync"
)

func main() {
	// PART 1 — uncomment to see the bug: nothing prints (most of the time).
	// for i := 0; i < 5; i++ {
	// 	go func(n int) { fmt.Println("hi from", n) }(i)
	// }
	// // main returns immediately — goroutines never get to run.

	// PART 2 — fix with sync.WaitGroup. wg.Add(N) before launching, wg.Done() inside,
	// wg.Wait() blocks until all N have called Done.
	var wg sync.WaitGroup

	// TODO: for i := 0; i < 5; i++ {
	// TODO:     wg.Add(1)
	// TODO:     go func(n int) {
	// TODO:         defer wg.Done()
	// TODO:         fmt.Println("hi from", n)
	// TODO:     }(i)
	// TODO: }
	// TODO: wg.Wait()

	_ = wg.Wait
	_ = fmt.Println
}
