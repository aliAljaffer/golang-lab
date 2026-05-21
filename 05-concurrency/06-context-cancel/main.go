// 06-context-cancel — propagate cancellation across goroutines.
//
// Goal: a worker that does periodic work and stops cleanly when its parent
// cancels the context.
//
// Run:
//   go run .
package main

import (
	"context"
	"fmt"
	"time"
)

// worker prints a tick once a second until ctx is cancelled.
// The select-on-ctx.Done() pattern is the canonical cancellable loop in Go.
func worker(ctx context.Context, name string) {
	t := time.NewTicker(200 * time.Millisecond)
	defer t.Stop()

	for {
		// TODO: select {
		// TODO: case <-ctx.Done():
		// TODO:     fmt.Printf("%s: stopping (%v)\n", name, ctx.Err())
		// TODO:     return
		// TODO: case now := <-t.C:
		// TODO:     fmt.Printf("%s: tick at %s\n", name, now.Format("15:04:05.000"))
		// TODO: }

		_ = ctx.Done
		_ = t.C
		return
	}
}

func main() {
	// WithCancel returns a derived context and a `cancel` func. Calling cancel
	// closes ctx.Done(); every goroutine listening on it wakes up at once.
	ctx, cancel := context.WithCancel(context.Background())

	go worker(ctx, "w1")
	go worker(ctx, "w2")

	time.Sleep(600 * time.Millisecond)
	fmt.Println("main: cancelling")
	cancel()

	// Give workers a moment to print their "stopping" line before main exits.
	time.Sleep(100 * time.Millisecond)
	fmt.Println("main: done")

	// Bonus: WithTimeout / WithDeadline are sugar over WithCancel — they call
	// cancel() automatically after the deadline. Pattern:
	//   ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	//   defer cancel() // always defer, even if the timeout already fired
}
