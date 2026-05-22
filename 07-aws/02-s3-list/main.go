// 02-s3-list — list S3 buckets, then page through one bucket's objects.
//
// Demonstrates:
//   - One service client per service: `s3.NewFromConfig(cfg)`.
//   - Every operation method takes `context.Context` + an input struct.
//   - Paginators are first-class in SDK v2: `s3.NewListObjectsV2Paginator`.
//     You loop with `for p.HasMorePages() { p.NextPage(ctx) }` — no manual
//     ContinuationToken juggling.
//
// Run:
//
//	go run . <bucket-name>
//
// If <bucket-name> is omitted, only ListBuckets runs.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load config:", err)
		os.Exit(1)
	}

	client := s3.NewFromConfig(cfg)

	// PART 1 — list buckets in the account.
	// TODO: out, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "list buckets:", err); os.Exit(1) }
	// TODO: for _, b := range out.Buckets {
	// TODO:     fmt.Printf("%s  %s\n", b.CreationDate.Format(time.RFC3339), *b.Name)
	// TODO: }

	if len(os.Args) < 2 {
		return
	}
	bucket := os.Args[1]

	// PART 2 — page through every object in <bucket>.
	// TODO: pager := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{Bucket: &bucket})
	// TODO: for pager.HasMorePages() {
	// TODO:     page, err := pager.NextPage(ctx)
	// TODO:     if err != nil { fmt.Fprintln(os.Stderr, "next page:", err); os.Exit(1) }
	// TODO:     for _, obj := range page.Contents {
	// TODO:         fmt.Printf("%10d  %s  %s\n", *obj.Size, obj.LastModified.Format(time.RFC3339), *obj.Key)
	// TODO:     }
	// TODO: }

	_ = client
	_ = bucket
}
