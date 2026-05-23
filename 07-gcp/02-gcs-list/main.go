// 02-gcs-list — list buckets in a project, then page through one bucket's
// objects using the GCS Iterator pattern.
//
// Demonstrates:
//   - One service client per service: `storage.NewClient(ctx)`. No region/zone
//     wiring — bucket names are globally unique.
//   - `client.Buckets(ctx, projectID)` returns a `*BucketIterator`.
//   - `bucket.Objects(ctx, &storage.Query{Prefix: ...})` returns an
//     `*ObjectIterator`.
//   - Both follow the same `it.Next()` → `(T, error)` shape, with
//     `iterator.Done` as the terminating sentinel. This is GCP's answer to
//     pagination — no manual page-token juggling.
//
// Run:
//
//	go run . <project-id>
//	go run . <project-id> <bucket-name>
//
// If <bucket-name> is omitted, only the bucket list runs.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run . <project-id> [bucket-name]")
		os.Exit(2)
	}
	project := os.Args[1]

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// storage.NewClient implicitly uses ADC. No project ID needed here — that
	// goes on per-call inputs (Buckets) since one client can talk to any project
	// you have access to.
	client, err := storage.NewClient(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "new client:", err)
		os.Exit(1)
	}
	defer client.Close()

	// PART 1 — list buckets in the project. The iterator pattern in 6 lines.
	// TODO: it := client.Buckets(ctx, project)
	// TODO: for {
	// TODO:     attrs, err := it.Next()
	// TODO:     if errors.Is(err, iterator.Done) { break }
	// TODO:     if err != nil { fmt.Fprintln(os.Stderr, "next bucket:", err); os.Exit(1) }
	// TODO:     fmt.Printf("%s  %s\n", attrs.Created.Format(time.RFC3339), attrs.Name)
	// TODO: }

	if len(os.Args) < 3 {
		return
	}
	bucket := os.Args[2]

	// PART 2 — page through every object in <bucket>. Same shape — different
	// iterator type. Query{} (empty) is "everything"; set Prefix to narrow.
	// TODO: oit := client.Bucket(bucket).Objects(ctx, &storage.Query{})
	// TODO: for {
	// TODO:     attrs, err := oit.Next()
	// TODO:     if errors.Is(err, iterator.Done) { break }
	// TODO:     if err != nil { fmt.Fprintln(os.Stderr, "next object:", err); os.Exit(1) }
	// TODO:     fmt.Printf("%10d  %s  %s\n", attrs.Size, attrs.Updated.Format(time.RFC3339), attrs.Name)
	// TODO: }

	_ = project
	_ = bucket
	_ = errors.Is
	_ = iterator.Done
}
