// 05-mutex — protect shared state.
//
// Goal: show that a naive concurrent map write races, and that sync.Mutex
// fixes it. Also: when to reach for RWMutex.
//
// Run:
//   go run .
//   go run -race .   # the racy version trips the detector
package main

import (
	"fmt"
	"sync"
)

// Counter is a concurrency-safe counter using a Mutex.
type Counter struct {
	mu sync.Mutex
	n  int
}

func (c *Counter) Inc() {
	// TODO: c.mu.Lock()
	// TODO: defer c.mu.Unlock()
	// TODO: c.n++
}

func (c *Counter) Value() int {
	// TODO: c.mu.Lock()
	// TODO: defer c.mu.Unlock()
	// TODO: return c.n
	return 0
}

func main() {
	// PART 1 — racy version (uncomment + run with `go run -race .` to see the race).
	// var n int
	// var wg sync.WaitGroup
	// for i := 0; i < 1000; i++ {
	//     wg.Add(1)
	//     go func() { defer wg.Done(); n++ }()
	// }
	// wg.Wait()
	// fmt.Println("racy n =", n) // usually < 1000

	// PART 2 — safe version using Counter.
	var c Counter
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); c.Inc() }()
	}
	wg.Wait()
	fmt.Println("safe n =", c.Value()) // exactly 1000 once Inc/Value are implemented
}
