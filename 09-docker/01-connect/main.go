// 01-connect — build a Docker client and print the daemon's version.
//
// What this example proves:
//   - `client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())`
//     reads DOCKER_HOST / DOCKER_TLS_VERIFY / DOCKER_CERT_PATH from the env
//     (just like the docker CLI), then handshakes API versions with the daemon.
//     Without `WithAPIVersionNegotiation`, you'll pin to whatever version the
//     SDK was built against — and any older daemon will reject your calls with
//     "client version X is too new for daemon API version Y".
//   - `cli.ServerVersion(ctx)` is the cheapest "is the daemon reachable?" probe.
//     Equivalent to `docker version` (the server half).
//   - Always `defer cli.Close()` — leaking the underlying HTTP connections will
//     bite you in a long-running tool that builds a fresh client per request.
//
// Run:
//
//	go run .
//	DOCKER_HOST=tcp://remote:2375 go run .
//
// Requires a reachable Docker daemon (`docker info` should succeed first).
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/docker/docker/client"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// TODO: cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "build client:", err); os.Exit(1) }
	// TODO: defer cli.Close()

	// TODO: ver, err := cli.ServerVersion(ctx)
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "server version:", err); os.Exit(1) }
	// TODO: fmt.Printf("daemon: %s (API %s, OS %s/%s)\n", ver.Version, ver.APIVersion, ver.Os, ver.Arch)

	_ = ctx
	_ = client.FromEnv
	_ = os.Exit
	_ = fmt.Println
}
