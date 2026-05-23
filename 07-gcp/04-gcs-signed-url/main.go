// 04-gcs-signed-url — generate a 5-minute V4 signed GET URL.
//
// A V4 signed URL lets a third party (a browser, a webhook handler, a
// teammate) fetch an object for a bounded window without any GCP credentials
// of their own. The server doing the signing must hold credentials capable of
// signing — see "Signing identity" below.
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

	"cloud.google.com/go/storage"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: go run . <bucket> <key>")
		os.Exit(2)
	}
	bucket, key := os.Args[1], os.Args[2]

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := storage.NewClient(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "new client:", err)
		os.Exit(1)
	}
	defer client.Close()

	// Signed-URL signing requires a signing identity (private key OR an SA
	// with `iam.serviceAccounts.signBlob` permission). If you're running with a
	// JSON key, the storage client picks up GoogleAccessID + PrivateKey from it
	// automatically — leave SignedURLOptions empty and SignedURL works.
	//
	// If you're running with `gcloud auth application-default login` (user
	// creds), there's no private key, so signing falls back to IAM signBlob —
	// you must pass GoogleAccessID = "your-sa@project.iam.gserviceaccount.com"
	// and the SDK will call iam.signBlob under the hood.
	// TODO: opts := &storage.SignedURLOptions{
	// TODO:     Method:  "GET",
	// TODO:     Expires: time.Now().Add(5 * time.Minute),
	// TODO:     Scheme:  storage.SigningSchemeV4,
	// TODO:     // GoogleAccessID: "your-sa@project.iam.gserviceaccount.com", // uncomment if using user creds
	// TODO: }
	// TODO: url, err := client.Bucket(bucket).SignedURL(key, opts)
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "signed url:", err); os.Exit(1) }
	// TODO: fmt.Println(url)

	_ = client
	_ = bucket
	_ = key
}
