// 04-s3-presigned — generate a 5-minute presigned GET URL.
//
// A presigned URL embeds an SDK-signed authentication so a third party (a
// browser, a webhook handler, a teammate without IAM) can fetch (or upload)
// a single object for a bounded window. The server doing the signing must
// hold credentials with permission to perform the action.
//
// Run:
//
//	go run . <bucket> <key>
//	# prints a URL. curl it within 5 minutes; after that, it 403s.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	_ "github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: go run . <bucket> <key>")
		os.Exit(2)
	}
	bucket, key := os.Args[1], os.Args[2]

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load config:", err)
		os.Exit(1)
	}

	// Presigning needs a separate client wrapper. The PresignClient signs
	// requests locally — it never calls AWS, so no creds are spent here.
	// TODO: presigner := s3.NewPresignClient(s3.NewFromConfig(cfg))

	// TODO: req, err := presigner.PresignGetObject(ctx,
	// TODO:     &s3.GetObjectInput{Bucket: &bucket, Key: &key},
	// TODO:     s3.WithPresignExpires(5*time.Minute),
	// TODO: )
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "presign:", err); os.Exit(1) }
	// TODO: fmt.Println(req.URL)

	_ = cfg
	_ = bucket
	_ = key
}
