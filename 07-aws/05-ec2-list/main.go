// 05-ec2-list — list EC2 instances, optionally filter by a tag.
//
// Demonstrates:
//   - Different service, same shape: `ec2.NewFromConfig(cfg)`.
//   - The Filter API for server-side narrowing — push the work to AWS, not
//     "list everything and filter in Go".
//   - DescribeInstances returns Reservations (each holds Instances) — a
//     historical wart, not a Go-SDK quirk.
//
// Run:
//
//	go run .                       # all instances
//	go run . Env prod              # only instances tagged Env=prod
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load config:", err)
		os.Exit(1)
	}
	client := ec2.NewFromConfig(cfg)

	in := &ec2.DescribeInstancesInput{}

	// Optional: filter by tag Key/Value passed on the CLI.
	if len(os.Args) == 3 {
		// TODO: in.Filters = []ec2types.Filter{{
		// TODO:     Name:   aws.String("tag:" + os.Args[1]),
		// TODO:     Values: []string{os.Args[2]},
		// TODO: }}
	}

	// Paginator pattern — same as S3.
	// TODO: p := ec2.NewDescribeInstancesPaginator(client, in)
	// TODO: for p.HasMorePages() {
	// TODO:     page, err := p.NextPage(ctx)
	// TODO:     if err != nil { fmt.Fprintln(os.Stderr, "next page:", err); os.Exit(1) }
	// TODO:     for _, r := range page.Reservations {
	// TODO:         for _, inst := range r.Instances {
	// TODO:             fmt.Printf("%s  %-12s  %s\n",
	// TODO:                 aws.ToString(inst.InstanceId),
	// TODO:                 inst.State.Name,
	// TODO:                 aws.ToString(inst.InstanceType.Values()[0]), // see README
	// TODO:             )
	// TODO:         }
	// TODO:     }
	// TODO: }

	_ = client
	_ = in
	_ = ec2types.Filter{}
	_ = aws.String
}
