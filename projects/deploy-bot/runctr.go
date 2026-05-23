package main

import (
	"context"
	"errors"

	"github.com/docker/docker/client"
)

// RunOpts is the request shape passed to Runner.Run.
type RunOpts struct {
	ImageID       string
	HostPort      int
	ContainerPort int
	Env           []string
	RemoveOnExit  bool // -> docker HostConfig.AutoRemove
}

// Runner creates + starts a container, and can remove it later. Production
// impl is *DockerRunner; tests use a fake recording calls.
type Runner interface {
	Run(ctx context.Context, opts RunOpts) (containerID string, err error)
	Remove(ctx context.Context, containerID string) error
}

// dockerSDK is the low-level slice DockerRunner uses. Production passes a
// *dockerRunAdapter wrapping a real *client.Client; tests pass a fake that
// captures what AutoRemove the caller asked for. Exactly the same shape as
// the s3-log-shipper / gcs-log-shipper uploader-adapter pattern.
type dockerSDK interface {
	containerCreate(ctx context.Context, image string, env []string, hostPort, containerPort int, autoRemove bool) (id string, err error)
	containerStart(ctx context.Context, id string) error
	containerRemove(ctx context.Context, id string) error
}

// DockerRunner wraps a dockerSDK with the high-level Run/Remove semantics.
//
// The whole point of this type vs. calling dockerSDK directly: tests can
// pin "DockerRunner forwards opts.RemoveOnExit to the SDK as AutoRemove"
// without needing a real docker daemon.
type DockerRunner struct {
	Inner dockerSDK
}

// Run creates a container with autoRemove=opts.RemoveOnExit, starts it,
// and returns its ID.
//
// Behavior contract:
//   - create error -> return ("", err); container does not exist.
//   - start error  -> the container WAS created but failed to start.
//                     Best-effort Inner.containerRemove on the orphan;
//                     ignore that remove's error so the caller sees the
//                     real start failure. Then return ("", startErr).
//   - success      -> return (id, nil).
//
// AutoRemove note: opts.RemoveOnExit MUST be propagated as the autoRemove
// argument to Inner.containerCreate. This is the load-bearing test for
// this type (TestDockerRunner_PassesAutoRemove).
func (r *DockerRunner) Run(ctx context.Context, opts RunOpts) (string, error) {
	// TODO: create + start. The two non-obvious decisions:
	//   - opts.RemoveOnExit must be forwarded as the autoRemove argument to
	//     containerCreate (TestDockerRunner_PassesAutoRemove pins this).
	//   - if start fails, the container already exists — best-effort remove
	//     the orphan, but return the original start error, not the remove's.
	return "", errors.New("DockerRunner.Run not implemented")
}

// Remove deletes the container by ID. Wraps the low-level remove so the
// pipeline's failure-cleanup branch (in run.go) can call it via the
// high-level Runner interface.
func (r *DockerRunner) Remove(ctx context.Context, containerID string) error {
	// TODO: pass through to the inner SDK.
	return errors.New("DockerRunner.Remove not implemented")
}

// dockerRunAdapter is the production dockerSDK: drives a real *client.Client.
// Tests do NOT use this — they wire a fake dockerSDK into DockerRunner.
type dockerRunAdapter struct {
	Client *client.Client
}

// containerCreate creates a container with the given image, env, port mapping,
// and AutoRemove setting. Returns the new container ID.
//
// HostConfig:
//   - PortBindings: containerPort/tcp -> hostPort
//   - AutoRemove: passed in
//
// Use `nat.Port(fmt.Sprintf("%d/tcp", containerPort))` for the port spec.
func (a *dockerRunAdapter) containerCreate(ctx context.Context, image string, env []string, hostPort, containerPort int, autoRemove bool) (string, error) {
	// TODO: build a container.Config + container.HostConfig and call
	//   a.Client.ContainerCreate. The fiddly bit is the port mapping —
	//   PortBindings keys are nat.Port (a typed "<port>/tcp" string) and
	//   the host side is also a string (strconv.Itoa(hostPort)).
	//   Don't forget HostConfig.AutoRemove = autoRemove.
	return "", errors.New("dockerRunAdapter.containerCreate not implemented")
}

// containerStart starts a created container by ID.
func (a *dockerRunAdapter) containerStart(ctx context.Context, id string) error {
	// TODO: pass through to a.Client.ContainerStart.
	return errors.New("dockerRunAdapter.containerStart not implemented")
}

// containerRemove force-removes the container by ID.
func (a *dockerRunAdapter) containerRemove(ctx context.Context, id string) error {
	// TODO: pass through to a.Client.ContainerRemove. Force=true so an
	//   already-running container is killed rather than refused.
	return errors.New("dockerRunAdapter.containerRemove not implemented")
}
