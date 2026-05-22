// 03-pull-and-run — pull an image, run a one-shot container, capture output, clean up.
//
// What this example proves:
//   - `cli.ImagePull(ctx, ref, image.PullOptions{})` returns an `io.ReadCloser`
//     of newline-delimited JSON progress events. You MUST drain it (read until
//     EOF) or the daemon won't consider the pull complete — even if you don't
//     care about the bytes, `io.Copy(io.Discard, rc)` is mandatory.
//   - `cli.ContainerCreate` returns just an ID; nothing is running yet.
//   - `cli.ContainerStart` flips it to running.
//   - `cli.ContainerWait` returns two channels (status + err) — select on both.
//     The `condition` arg ("not-running" vs "next-exit" vs "removed") controls
//     what counts as "done."
//   - `cli.ContainerLogs` returns the multiplexed stdout+stderr stream;
//     `stdcopy.StdCopy` is the demuxer (see 04-logs-stream for why).
//   - `cli.ContainerRemove` cleans up. Pass `Force: true` if it might still be
//     running; in this example we Wait first so a plain Remove is enough.
//
// Run:
//
//	go run .
//
// Equivalent to: docker run --rm alpine:3 echo "hello from a goroutine"
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Fprintln(os.Stderr, "build client:", err)
		os.Exit(1)
	}
	defer cli.Close()

	const ref = "alpine:3"

	// 1. PULL
	// TODO: rc, err := cli.ImagePull(ctx, ref, image.PullOptions{})
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "pull:", err); os.Exit(1) }
	// TODO: _, _ = io.Copy(io.Discard, rc)  // MUST drain to EOF
	// TODO: rc.Close()

	// 2. CREATE
	// TODO: created, err := cli.ContainerCreate(ctx, &container.Config{
	// TODO:     Image: ref,
	// TODO:     Cmd:   []string{"echo", "hello from a goroutine"},
	// TODO: }, nil, nil, nil, "")
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "create:", err); os.Exit(1) }

	// 3. START
	// TODO: if err := cli.ContainerStart(ctx, created.ID, container.StartOptions{}); err != nil {
	// TODO:     fmt.Fprintln(os.Stderr, "start:", err); os.Exit(1)
	// TODO: }

	// 4. WAIT for the process to exit
	// TODO: statusCh, errCh := cli.ContainerWait(ctx, created.ID, container.WaitConditionNotRunning)
	// TODO: select {
	// TODO: case err := <-errCh:
	// TODO:     if err != nil { fmt.Fprintln(os.Stderr, "wait:", err); os.Exit(1) }
	// TODO: case status := <-statusCh:
	// TODO:     fmt.Fprintf(os.Stderr, "container exited code=%d\n", status.StatusCode)
	// TODO: }

	// 5. LOGS — multiplexed stdout+stderr; demux into our own writers.
	// TODO: logs, err := cli.ContainerLogs(ctx, created.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "logs:", err); os.Exit(1) }
	// TODO: defer logs.Close()
	// TODO: _, _ = stdcopy.StdCopy(os.Stdout, os.Stderr, logs)

	// 6. REMOVE
	// TODO: if err := cli.ContainerRemove(ctx, created.ID, container.RemoveOptions{}); err != nil {
	// TODO:     fmt.Fprintln(os.Stderr, "remove:", err); os.Exit(1)
	// TODO: }

	_ = ctx
	_ = cli
	_ = ref
	_ = io.Copy
	_ = image.PullOptions{}
	_ = container.Config{}
	_ = stdcopy.StdCopy
}
