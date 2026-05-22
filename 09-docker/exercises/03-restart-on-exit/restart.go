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
	// TODO: check msg.Type, msg.Action.
	// TODO: read exit code from msg.Actor.Attributes["exitCode"]; if "0" or empty, return false.
	// TODO: return true otherwise.
	return false
}

// Run subscribes to the daemon's event stream, filters to container "die"
// events, and calls ContainerStart on any that exited non-zero.
//
// Returns nil on clean ctx cancellation; the first transport error otherwise.
// ContainerStart errors are NOT fatal — they're logged elsewhere and Run continues.
func Run(ctx context.Context, api DockerAPI) error {
	// TODO: f := filters.NewArgs(filters.Arg("type", "container"), filters.Arg("event", "die"))
	// TODO: msgCh, errCh := api.Events(ctx, events.ListOptions{Filters: f})
	// TODO: for {
	// TODO:     select {
	// TODO:     case <-ctx.Done(): return nil
	// TODO:     case err := <-errCh:
	// TODO:         if err != nil && ctx.Err() == nil { return err }
	// TODO:         return nil
	// TODO:     case msg := <-msgCh:
	// TODO:         if !ShouldRestart(msg) { continue }
	// TODO:         _ = api.ContainerStart(ctx, msg.Actor.ID, container.StartOptions{})  // ignore start errors; continue
	// TODO:     }
	// TODO: }
	_ = errors.New
	return errors.New("Run not implemented")
}
