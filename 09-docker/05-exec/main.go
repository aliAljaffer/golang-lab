// 05-exec — run a command in a running container, capture stdout/stderr, exit code.
//
// What this example proves:
//   - `cli.ContainerExecCreate(ctx, container, container.ExecOptions{...})`
//     reserves an exec slot in the daemon and returns an exec ID. Nothing is
//     running yet.
//   - `cli.ContainerExecAttach(ctx, execID, container.ExecAttachOptions{})`
//     starts the exec AND attaches your TCP connection to its stdin+stdout+stderr.
//     The returned `HijackedResponse` has a `Reader` (multiplexed, same as logs)
//     and a `Conn` (writable for stdin).
//   - `cli.ContainerExecInspect(ctx, execID)` is how you read the exit code
//     AFTER the exec finishes (when `Running` flips to false). The HijackedResponse
//     itself doesn't give you the exit code.
//
// The two-phase API (Create then Attach) exists because you can also call
// `ContainerExecStart` to launch without attaching — useful for fire-and-forget
// healthchecks. Most tools want Attach.
//
// Run:
//
//	# in another terminal:
//	docker run -d --name worker alpine sleep 600
//	# then:
//	go run . worker -- sh -c 'echo hi; echo oops >&2; exit 7'
//	# clean up:
//	docker rm -f worker
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

func main() {
	// Expect: <prog> <container> -- <cmd> [args...]
	if len(os.Args) < 4 || os.Args[2] != "--" {
		fmt.Fprintln(os.Stderr, "usage: 05-exec <container> -- <cmd> [args...]")
		os.Exit(2)
	}
	target := os.Args[1]
	cmd := os.Args[3:]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Fprintln(os.Stderr, "build client:", err)
		os.Exit(1)
	}
	defer cli.Close()

	// 1. CREATE the exec
	// TODO: created, err := cli.ContainerExecCreate(ctx, target, container.ExecOptions{
	// TODO:     Cmd:          cmd,
	// TODO:     AttachStdout: true,
	// TODO:     AttachStderr: true,
	// TODO: })
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "exec create:", err); os.Exit(1) }

	// 2. ATTACH (this starts the exec)
	// TODO: resp, err := cli.ContainerExecAttach(ctx, created.ID, container.ExecAttachOptions{})
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "exec attach:", err); os.Exit(1) }
	// TODO: defer resp.Close()

	// 3. STREAM the multiplexed output (same format as `04-logs-stream`)
	// TODO: if _, err := stdcopy.StdCopy(os.Stdout, os.Stderr, resp.Reader); err != nil {
	// TODO:     fmt.Fprintln(os.Stderr, "stdcopy:", err); os.Exit(1)
	// TODO: }

	// 4. INSPECT to harvest the exit code (HijackedResponse doesn't carry it)
	// TODO: inspect, err := cli.ContainerExecInspect(ctx, created.ID)
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "exec inspect:", err); os.Exit(1) }
	// TODO: os.Exit(inspect.ExitCode)

	_ = ctx
	_ = cli
	_ = target
	_ = cmd
	_ = container.ExecOptions{}
	_ = stdcopy.StdCopy
}
