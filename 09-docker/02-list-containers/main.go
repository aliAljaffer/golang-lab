// 02-list-containers — list running containers, then list everything.
//
// What this example proves:
//   - `cli.ContainerList(ctx, container.ListOptions{All: false})` returns only
//     running containers — same default as `docker ps`.
//   - `All: true` includes stopped/exited ones — same as `docker ps -a`.
//   - The returned `[]types.Container` has the fields you usually want without
//     a follow-up Inspect: ID, Image, Names ([]string), State, Status, Created.
//   - `filters.NewArgs(filters.Arg("status", "running"))` is the same DSL
//     `docker ps -f status=running` uses. The Filters key is open-ended; not
//     every key is valid for every endpoint — see the docker daemon API docs
//     for which `Filter` keys an endpoint accepts.
//
// One gotcha:
//   - `c.Names` is a slice (one container can have multiple names from network
//     aliases, etc.) and each entry comes prefixed with "/" — strip it before
//     displaying. `docker ps` does the same thing internally.
//
// Run:
//
//	go run .
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Fprintln(os.Stderr, "build client:", err)
		os.Exit(1)
	}
	defer cli.Close()

	// TODO: running, err := cli.ContainerList(ctx, container.ListOptions{All: false})
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "list running:", err); os.Exit(1) }
	// TODO: fmt.Printf("running (%d):\n", len(running))
	// TODO: for _, c := range running {
	// TODO:     name := strings.TrimPrefix(c.Names[0], "/")
	// TODO:     fmt.Printf("  %s  %-30s  %s  %s\n", c.ID[:12], name, c.Image, c.State)
	// TODO: }

	// TODO: all, err := cli.ContainerList(ctx, container.ListOptions{
	// TODO:     All:     true,
	// TODO:     Filters: filters.NewArgs(filters.Arg("status", "exited")),
	// TODO: })
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "list exited:", err); os.Exit(1) }
	// TODO: fmt.Printf("exited (%d):\n", len(all))

	_ = ctx
	_ = cli
	_ = container.ListOptions{}
	_ = filters.NewArgs
	_ = strings.TrimPrefix
}
