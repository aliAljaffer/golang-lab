// 03-select — multiplexing channels + timeout pattern.
//
// Goal: feel `select` as Go's `switch` for channel ops.
//
// Run:
//   go run .
package main

import (
	"fmt"
	"time"
)

func main() {
	a := make(chan string, 1)
	b := make(chan string, 1)

	// Two producers running at different speeds.
	go func() { time.Sleep(50 * time.Millisecond); a <- "from a" }()
	go func() { time.Sleep(20 * time.Millisecond); b <- "from b" }()

	// PART 1 — receive from whichever channel is ready first.
	// If multiple are ready, Go picks one at random — do not rely on order.
	// TODO: select {
	// TODO: case msg := <-a:
	// TODO:     fmt.Println("got", msg)
	// TODO: case msg := <-b:
	// TODO:     fmt.Println("got", msg)
	// TODO: }

	// PART 2 — timeout pattern. time.After returns a channel that fires once.
	slow := make(chan string)
	go func() { time.Sleep(200 * time.Millisecond); slow <- "eventually" }()

	// TODO: select {
	// TODO: case msg := <-slow:
	// TODO:     fmt.Println("slow finished:", msg)
	// TODO: case <-time.After(50 * time.Millisecond):
	// TODO:     fmt.Println("timed out waiting for slow")
	// TODO: }

	// PART 3 — non-blocking try-receive with `default`.
	empty := make(chan int)
	// TODO: select {
	// TODO: case v := <-empty:
	// TODO:     fmt.Println("got", v)
	// TODO: default:
	// TODO:     fmt.Println("nothing ready right now (default branch)")
	// TODO: }

	_ = empty
	_ = fmt.Println
}
