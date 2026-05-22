// 06-assume-role — use STS to assume a cross-account role, then use the
// derived credentials to make a call.
//
// The pattern, end-to-end:
//  1. Load default config (the "source" creds — usually your dev/CI principal).
//  2. Build an STS client from those.
//  3. Wrap STS in `stscreds.NewAssumeRoleProvider(stsClient, roleARN)`.
//  4. Build a NEW config whose credentials are that provider.
//  5. Build service clients from the new config — every call is now scoped to
//     the assumed role.
//
// Run:
//
//	go run . arn:aws:iam::123456789012:role/AssumableRole
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: go run . <role-arn>")
		os.Exit(2)
	}
	roleARN := os.Args[1]

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// PART 1 — load source creds (whatever you'd normally run as).
	srcCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load source config:", err)
		os.Exit(1)
	}

	// PART 2 — build an STS client from source creds and an AssumeRole provider.
	// TODO: stsClient := sts.NewFromConfig(srcCfg)
	// TODO: provider := stscreds.NewAssumeRoleProvider(stsClient, roleARN, func(o *stscreds.AssumeRoleOptions) {
	// TODO:     o.RoleSessionName = "go-learning-session"
	// TODO:     o.Duration = 15 * time.Minute
	// TODO: })

	// PART 3 — derive a new config that signs with the assumed-role creds.
	// aws.NewCredentialsCache wraps the provider so repeated calls don't
	// re-issue AssumeRole every time.
	// TODO: assumedCfg := srcCfg.Copy()
	// TODO: assumedCfg.Credentials = aws.NewCredentialsCache(provider)

	// PART 4 — make a call as the assumed role. ListBuckets is a fine smoke test.
	// TODO: s3c := s3.NewFromConfig(assumedCfg)
	// TODO: out, err := s3c.ListBuckets(ctx, &s3.ListBucketsInput{})
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "list buckets as assumed:", err); os.Exit(1) }
	// TODO: fmt.Printf("assumed-role sees %d bucket(s)\n", len(out.Buckets))

	_ = srcCfg
	_ = roleARN
	_ = sts.NewFromConfig
	_ = stscreds.NewAssumeRoleProvider
	_ = s3.NewFromConfig
	_ = aws.NewCredentialsCache
}
