// 05-signals — graceful shutdown on SIGINT / SIGTERM.
//
// Goal: print "working..." once a second. On Ctrl-C, print "cleaning up..."
// then exit 0.
//
// Run:
//   go run .
//   (press Ctrl-C after a few seconds)
package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// TODO: ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	// TODO: defer stop()

	// TODO: ticker := time.NewTicker(1 * time.Second). defer ticker.Stop().

	// TODO: for { select { case <-ctx.Done(): ... ; case <-ticker.C: print "working..." } }

	// TODO: print "cleaning up..." once the loop exits. Optionally sleep a bit
	//       to simulate flushing buffers, then return cleanly.

	_ = context.Background
	_ = signal.NotifyContext
	_ = syscall.SIGINT
	_ = time.Second
	_ = fmt.Println
}
