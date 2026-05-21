// 08-race-detector — a program that races; show `-race` catches it.
//
// Goal: see what a real race looks like, see what the detector reports,
// and learn to make `-race` part of your CI.
//
// Run:
//   go run .            # may "work" — output is undefined
//   go run -race .      # prints a WARNING: DATA RACE with stack traces
package main

import (
	"fmt"
	"sync"
)

func main() {
	// PART 1 — racy: many goroutines increment the same int without a lock.
	// The compiled program may "work" if the scheduler is kind to you, but the
	// behavior is undefined and shifts on different machines / loads / Go versions.
	var n int
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			n++ // RACE: concurrent read-modify-write of `n`
		}()
	}
	wg.Wait()
	fmt.Println("racy n =", n) // usually < 1000

	// PART 2 — fix the race with a mutex. Uncomment to make `-race` happy.
	// TODO: var mu sync.Mutex
	// TODO: var safe int
	// TODO: for i := 0; i < 1000; i++ {
	// TODO:     wg.Add(1)
	// TODO:     go func() {
	// TODO:         defer wg.Done()
	// TODO:         mu.Lock()
	// TODO:         safe++
	// TODO:         mu.Unlock()
	// TODO:     }()
	// TODO: }
	// TODO: wg.Wait()
	// TODO: fmt.Println("safe n =", safe)
}
