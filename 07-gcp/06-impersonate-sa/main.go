// 06-impersonate-sa — impersonate a target service account, then use the
// short-lived token to build a GCS client that acts as that SA.
//
// The pattern, end-to-end:
//  1. Caller principal (your gcloud user, your CI SA, your GKE WI SA — whatever
//     ADC resolves to) has `roles/iam.serviceAccountTokenCreator` on the target SA.
//  2. `impersonate.CredentialsTokenSource(ctx, CredentialsConfig{TargetPrincipal: <sa-email>, Scopes: [...]})`
//     mints short-lived tokens for the target. The token source caches and
//     refreshes them transparently.
//  3. Pass the token source as `option.WithTokenSource(ts)` to the service
//     client — every API call now signs as the target SA, not the caller.
//
// This is the GCP equivalent of AWS's STS AssumeRole.
//
// Run:
//
//	go run . <target-sa-email> <project>
//	# e.g. go run . readonly@my-project.iam.gserviceaccount.com my-project
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/impersonate"
	"google.golang.org/api/option"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: go run . <target-sa-email> <project>")
		os.Exit(2)
	}
	targetSA, project := os.Args[1], os.Args[2]

	_, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// PART 1 — build an impersonation token source. Lifetime is bounded; the
	// source refreshes transparently before expiry.
	// TODO: ts, err := impersonate.CredentialsTokenSource(ctx, impersonate.CredentialsConfig{
	// TODO:     TargetPrincipal: targetSA,
	// TODO:     Scopes:          []string{"https://www.googleapis.com/auth/cloud-platform"},
	// TODO:     Lifetime:        15 * time.Minute,
	// TODO: })
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "impersonate token source:", err); os.Exit(1) }

	// PART 2 — build a GCS client signed by the target SA's tokens.
	// TODO: client, err := storage.NewClient(ctx, option.WithTokenSource(ts))
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "new storage client (impersonated):", err); os.Exit(1) }
	// TODO: defer client.Close()

	// PART 3 — smoke test: list buckets the TARGET SA can see. If the target
	// can't see them, this errors with a Permission Denied — which proves
	// you're actually signing as the target, not as the caller.
	// TODO: it := client.Buckets(ctx, project)
	// TODO: count := 0
	// TODO: for {
	// TODO:     attrs, err := it.Next()
	// TODO:     if errors.Is(err, iterator.Done) { break }
	// TODO:     if err != nil { fmt.Fprintln(os.Stderr, "list buckets (impersonated):", err); os.Exit(1) }
	// TODO:     count++
	// TODO:     _ = attrs
	// TODO: }
	// TODO: fmt.Printf("target SA %s sees %d bucket(s) in %s\n", targetSA, count, project)

	_ = targetSA
	_ = project
	_ = impersonate.CredentialsConfig{}
	_ = option.WithTokenSource
	_ = storage.NewClient
}
