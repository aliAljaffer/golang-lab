// Package restart watches Docker container events and restarts any container
// that exits non-zero.
//
// Exercise surface:
//
//	type DockerAPI interface {
//	    Events(ctx, opts) (<-chan events.Message, <-chan error)
//	    ContainerStart(ctx, id, opts) error
//	}
//
//	func ShouldRestart(msg events.Message) bool
//	func Run(ctx context.Context, api DockerAPI) error
//
// Why split these?
//   - `ShouldRestart` is a pure function on a single event — easy to unit-test
//     with hand-built `events.Message` values.
//   - `Run` is the loop — tested via a stub `DockerAPI` that feeds canned events
//     through a channel.
//
// Production-grade hardening you'd add on top (NOT in scope):
//   - Exponential backoff on repeated crash loops.
//   - Limits ("max 5 restarts per container per hour").
//   - Label-based opt-in: only restart containers with `auto-restart=true`.
//   - The daemon already has `--restart=on-failure` — this exercise replicates
//     it for the pedagogical exercise, not to ship.
package restart

import (
	"context"
	"errors"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
)

// DockerAPI is the slice of *client.Client this package uses. A real
// `*client.Client` satisfies it; tests pass a fake.
type DockerAPI interface {
	Events(ctx context.Context, opts events.ListOptions) (<-chan events.Message, <-chan error)
	ContainerStart(ctx context.Context, id string, opts container.StartOptions) error
}

// ShouldRestart decides whether a single event represents a non-zero exit
// that should trigger a restart.
//
// Contract pinned by tests:
//   - msg.Type == "container" AND msg.Action == "die" AND
//     msg.Actor.Attributes["exitCode"] is set AND != "0"  → true
//   - Anything else → false.
//
// Why not just "die" by itself? Because `docker stop` also fires a "die"
// event (containers stopped by a user count as exited). The daemon includes
// `exitCode` in the actor attributes; "0" means clean shutdown.
func ShouldRestart(msg events.Message) bool {
	// TODO: implement the truth table from the doc above. The trap is that
	//   "die" alone isn't enough — `docker stop` also fires die, with exitCode
	//   "0". The exitCode attribute is the discriminator.
	return false
}

// Run subscribes to the daemon's event stream, filters to container "die"
// events, and calls ContainerStart on any that exited non-zero.
//
// Returns nil on clean ctx cancellation; the first transport error otherwise.
// ContainerStart errors are NOT fatal — they're logged elsewhere and Run continues.
func Run(ctx context.Context, api DockerAPI) error {
	// TODO: subscribe to the daemon's event stream and dispatch on every
	//   incoming message. Decisions the tests pin:
	//     - filter server-side to type=container, event=die — saves work
	//       on the client.
	//     - errCh after ctx.Done() is expected (transport closes); only
	//       treat it as an error when ctx is still alive.
	//     - ContainerStart errors are NON-fatal — log if you want, but the
	//       loop must keep running. A single bad start shouldn't kill the
	//       watcher.
	_ = errors.New
	return errors.New("Run not implemented")
}
