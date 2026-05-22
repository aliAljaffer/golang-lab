// 06-events — subscribe to the daemon's event stream and print events as they happen.
//
// What this example proves:
//   - `cli.Events(ctx, events.ListOptions{Filters: ...})` returns two channels:
//     a `<-chan events.Message` and a `<-chan error`. The same two-channel
//     pattern as `ContainerWait` (03-pull-and-run).
//   - The `filters.Args` lets you narrow to specific event types/actions —
//     same DSL as 02-list-containers and the same as `docker events --filter`.
//   - The stream stays open as long as ctx is alive. Cancellation closes both
//     channels.
//
// What an event looks like:
//
//	events.Message{
//	  Type:   "container",        // container | image | volume | network | daemon | plugin | service | node ...
//	  Action: "start",            // create, start, die, kill, pull, untag, ...
//	  Actor: events.Actor{
//	    ID:         "abcd1234...",
//	    Attributes: map[string]string{"image": "alpine:3", "name": "talker", "exitCode": "0"},
//	  },
//	  Time:     1737550000,       // seconds since epoch
//	  TimeNano: 1737550000123456,
//	}
//
// Run:
//
//	go run .
//	# in another terminal:
//	docker run --rm alpine echo hi
//	# observe: container create -> attach -> start -> die -> destroy
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Fprintln(os.Stderr, "build client:", err)
		os.Exit(1)
	}
	defer cli.Close()

	// Narrow to container events only. Try removing this to see image pulls,
	// network attachments, volume mounts, etc.
	// TODO: f := filters.NewArgs(filters.Arg("type", "container"))

	// TODO: msgCh, errCh := cli.Events(ctx, events.ListOptions{Filters: f})

	// TODO: for {
	// TODO:     select {
	// TODO:     case <-ctx.Done():
	// TODO:         return
	// TODO:     case err, ok := <-errCh:
	// TODO:         if !ok { return }
	// TODO:         if err != nil && ctx.Err() == nil {
	// TODO:             fmt.Fprintln(os.Stderr, "events:", err); os.Exit(1)
	// TODO:         }
	// TODO:         return
	// TODO:     case msg := <-msgCh:
	// TODO:         name := msg.Actor.Attributes["name"]
	// TODO:         fmt.Printf("%-10s %-10s %s  %s\n", msg.Type, msg.Action, msg.Actor.ID[:12], name)
	// TODO:     }
	// TODO: }

	_ = ctx
	_ = cli
	_ = events.ListOptions{}
	_ = filters.NewArgs
}
