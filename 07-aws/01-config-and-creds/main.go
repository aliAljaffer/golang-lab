// 01-config-and-creds — load the SDK default config and inspect what it picked up.
//
// What this example proves:
//   - `config.LoadDefaultConfig(ctx)` walks the credentials chain:
//       env vars → ~/.aws/credentials → ~/.aws/config → IMDS (on EC2/ECS)
//   - The Region resolves from the same layered lookup (env AWS_REGION /
//     AWS_DEFAULT_REGION, then the active profile's `region = ...`).
//   - Credentials are NOT fetched until you call `cfg.Credentials.Retrieve(ctx)`
//     — that call is what actually hits the provider (env read, file read,
//     IMDS HTTP call, SSO refresh, etc.).
//
// Run:
//
//	go run .
//	AWS_PROFILE=my-prof go run .
//	AWS_REGION=eu-west-1 go run .
//
// Requires real credentials to print a non-empty AccessKeyID.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// TODO: cfg, err := config.LoadDefaultConfig(ctx)
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "load config:", err); os.Exit(1) }
	// TODO: fmt.Println("region:", cfg.Region)

	// TODO: creds, err := cfg.Credentials.Retrieve(ctx)
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "retrieve creds:", err); os.Exit(1) }
	// TODO: fmt.Println("provider:", creds.Source)
	// TODO: fmt.Println("access key id:", creds.AccessKeyID)
	// TODO: fmt.Println("expires:", creds.Expires) // zero value if static

	_ = ctx
	_ = config.LoadDefaultConfig
	_ = os.Exit
	_ = fmt.Println
}
