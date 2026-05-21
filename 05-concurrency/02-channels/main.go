// 02-channels — unbuffered vs buffered channels.
//
// Goal: feel the difference between a synchronous handoff (unbuffered)
// and a small queue (buffered). Both are typed pipes.
//
// Run:
//   go run .
package main

import (
	"fmt"
	"time"
)

func main() {
	// PART 1 — unbuffered: send blocks until a receiver is ready (and vice versa).
	// Try swapping the order of the two operations and watch it deadlock.
	unbuffered := make(chan string)
	go func() {
		time.Sleep(100 * time.Millisecond)
		unbuffered <- "hello" // blocks until main receives
	}()
	msg := <-unbuffered // blocks until the goroutine sends
	fmt.Println("unbuffered:", msg)

	// PART 2 — buffered: send returns immediately if there's room in the buffer.
	// TODO: ch := make(chan int, 3)
	// TODO: ch <- 1
	// TODO: ch <- 2
	// TODO: ch <- 3
	// TODO: // ch <- 4  // this would block — buffer full and no receiver yet
	// TODO: close(ch)
	// TODO: for v := range ch { fmt.Println("buffered:", v) }
	//
	// Notice: `for v := range ch` exits when the channel is closed *and* drained.
	// If you forget close(ch), `range` blocks forever — a classic goroutine leak.

	// PART 3 — comma-ok receive distinguishes "got a value" from "channel closed".
	// TODO: done := make(chan struct{})
	// TODO: close(done)
	// TODO: v, ok := <-done
	// TODO: fmt.Printf("done: v=%v ok=%v (ok=false means channel closed and empty)\n", v, ok)

	_ = fmt.Println
}
