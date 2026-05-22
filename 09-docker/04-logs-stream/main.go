// 04-logs-stream — follow a running container's logs to stdout/stderr.
//
// What this example proves:
//   - `cli.ContainerLogs(ctx, id, opts)` returns an `io.ReadCloser` of the
//     daemon's multiplexed log stream. `Follow: true` makes it behave like
//     `docker logs -f` — the reader stays open and yields new bytes as they
//     arrive, until the container exits or you close it.
//   - `stdcopy.StdCopy(stdoutW, stderrW, r)` is the ONLY correct way to copy
//     that stream into separate writers. Each chunk is `[8-byte header][N bytes]`
//     where the header's first byte says "this is stdout (1) or stderr (2)"
//     and the last 4 bytes are the length. `io.Copy` would write the header
//     bytes into your output. Don't.
//   - Cancellation: when `ctx` is done OR the response body's `Close()` is
//     called, the underlying HTTP read returns; `StdCopy` returns. That's
//     your "stop tailing" signal.
//
// One footgun:
//   - When the container has `TTY: true`, the stream is NOT multiplexed —
//     there's no header, and `stdcopy.StdCopy` will produce garbage. In that
//     case use `io.Copy(stdout, logs)` directly. The example assumes no TTY.
//
// Run:
//
//	# in another terminal:
//	docker run -d --name talker alpine sh -c 'i=0; while :; do echo line $i; echo err $i >&2; i=$((i+1)); sleep 1; done'
//	# then:
//	go run . talker
//	# ctrl-c to stop. Don't forget: docker rm -f talker
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: 04-logs-stream <container-id-or-name>")
		os.Exit(2)
	}
	target := os.Args[1]

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Fprintln(os.Stderr, "build client:", err)
		os.Exit(1)
	}
	defer cli.Close()

	// TODO: rc, err := cli.ContainerLogs(ctx, target, container.LogsOptions{
	// TODO:     ShowStdout: true,
	// TODO:     ShowStderr: true,
	// TODO:     Follow:     true,
	// TODO:     Tail:       "10",  // last 10 lines before the follow starts; use "all" for everything.
	// TODO: })
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "logs:", err); os.Exit(1) }
	// TODO: defer rc.Close()

	// TODO: // StdCopy returns when rc returns EOF / error, i.e. when the
	// TODO: // container exits OR our ctx-driven Close() fires.
	// TODO: if _, err := stdcopy.StdCopy(os.Stdout, os.Stderr, rc); err != nil && ctx.Err() == nil {
	// TODO:     fmt.Fprintln(os.Stderr, "stdcopy:", err)
	// TODO:     os.Exit(1)
	// TODO: }

	_ = ctx
	_ = cli
	_ = target
	_ = container.LogsOptions{}
	_ = stdcopy.StdCopy
}
