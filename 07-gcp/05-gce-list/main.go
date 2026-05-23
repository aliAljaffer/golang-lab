// 05-gce-list — list GCE instances across all zones via AggregatedList, then
// optionally narrow by label.
//
// Demonstrates:
//   - `compute.NewInstancesRESTClient(ctx)` — one of the new generated REST/gRPC
//     clients in `cloud.google.com/go/compute/apiv1`. Note the `REST` suffix:
//     there's also a NewInstancesClient for gRPC, but the REST flavor is the
//     one most code uses (lighter, no protobuf streaming).
//   - `AggregatedList(ctx, req)` returns instances grouped by zone — the GCP
//     model where compute is regional/zonal and you always have to flatten.
//   - The Iterator pattern again: `it.Next()` → `(key, scopedList, err)` →
//     `iterator.Done` to stop.
//   - `Filter: "labels.env=prod"` — GCP's server-side filter DSL.
//
// Run:
//
//	go run . <project>
//	go run . <project> env prod         # only instances with label env=prod
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run . <project> [label-key label-value]")
		os.Exit(2)
	}
	project := os.Args[1]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "new instances client:", err)
		os.Exit(1)
	}
	defer client.Close()

	req := &computepb.AggregatedListInstancesRequest{
		Project: project,
	}

	// Optional label filter — GCP-server-side. Push the work to GCE, not
	// "list everything and filter in Go".
	if len(os.Args) == 4 {
		// TODO: filter := fmt.Sprintf("labels.%s=%s", os.Args[2], os.Args[3])
		// TODO: req.Filter = &filter
	}

	// AggregatedList: iterate by zone. Each item is
	// (key="zones/<zone>", value=*InstancesScopedList containing []*Instance).
	// Most zones will be empty for a small project — skip them.
	// TODO: it := client.AggregatedList(ctx, req)
	// TODO: for {
	// TODO:     pair, err := it.Next()
	// TODO:     if errors.Is(err, iterator.Done) { break }
	// TODO:     if err != nil { fmt.Fprintln(os.Stderr, "next zone:", err); os.Exit(1) }
	// TODO:     if pair.Value == nil || len(pair.Value.Instances) == 0 {
	// TODO:         continue // empty zone — skip silently
	// TODO:     }
	// TODO:     for _, inst := range pair.Value.Instances {
	// TODO:         fmt.Printf("%s  %-30s  %s\n",
	// TODO:             pair.Key,                  // "zones/us-central1-a"
	// TODO:             inst.GetName(),
	// TODO:             inst.GetStatus(),          // RUNNING, TERMINATED, STOPPED, ...
	// TODO:         )
	// TODO:     }
	// TODO: }

	_ = req
	_ = errors.Is
	_ = iterator.Done
}
