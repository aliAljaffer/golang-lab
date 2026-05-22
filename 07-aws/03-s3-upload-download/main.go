// 03-s3-upload-download — PutObject (upload) then GetObject (download) round-trip.
//
// The minimal path: PutObject takes a Body that is an io.Reader, GetObject
// returns a Body that is an io.ReadCloser. No manager, no multipart. Good
// enough for files < 5 MB.
//
// (For >5 MB or unknown size, prefer `feature/s3/manager.Uploader` — it does
// concurrent multipart automatically. Not used here to keep the example small.)
//
// Run:
//
//	go run . <bucket> <key> <local-file>
//	# uploads <local-file> to s3://<bucket>/<key>, then downloads it to /tmp/<basename>.dl
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "usage: go run . <bucket> <key> <local-file>")
		os.Exit(2)
	}
	bucket, key, src := os.Args[1], os.Args[2], os.Args[3]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load config:", err)
		os.Exit(1)
	}
	client := s3.NewFromConfig(cfg)

	// PART 1 — upload.
	// TODO: f, err := os.Open(src)
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "open:", err); os.Exit(1) }
	// TODO: defer f.Close()
	// TODO: _, err = client.PutObject(ctx, &s3.PutObjectInput{
	// TODO:     Bucket: &bucket,
	// TODO:     Key:    &key,
	// TODO:     Body:   f,
	// TODO: })
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "put:", err); os.Exit(1) }
	// TODO: fmt.Printf("uploaded s3://%s/%s\n", bucket, key)

	// PART 2 — download to /tmp/<basename>.dl.
	dst := filepath.Join("/tmp", filepath.Base(src)+".dl")
	// TODO: out, err := client.GetObject(ctx, &s3.GetObjectInput{Bucket: &bucket, Key: &key})
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "get:", err); os.Exit(1) }
	// TODO: defer out.Body.Close()
	// TODO: w, err := os.Create(dst)
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "create:", err); os.Exit(1) }
	// TODO: defer w.Close()
	// TODO: n, err := io.Copy(w, out.Body)
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "copy:", err); os.Exit(1) }
	// TODO: fmt.Printf("downloaded %d bytes to %s\n", n, dst)

	_ = client
	_ = dst
	_ = bucket
	_ = key
	_ = io.Copy
}
