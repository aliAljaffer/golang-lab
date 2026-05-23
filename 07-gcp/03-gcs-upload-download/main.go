// 03-gcs-upload-download — write an object via the bucket Writer, read it back
// via the bucket Reader. The minimal round-trip.
//
// The asymmetry to internalize:
//   - Upload: `obj.NewWriter(ctx)` returns an `*io.Writer` you write to, then
//     `w.Close()` is what actually commits. Without Close, nothing is uploaded.
//   - Download: `obj.NewReader(ctx)` returns an `*io.ReadCloser` that streams.
//     `io.Copy(dst, r)` is fine for any size — GCS chunks under the hood.
//
// (For >5 MB uploads with resumability, the Writer transparently chunks +
// retries internally. Configurable via `Writer.ChunkSize`. Not toggled here.)
//
// Run:
//
//	go run . <bucket> <key> <local-file>
//	# uploads <local-file> to gs://<bucket>/<key>, then downloads it to /tmp/<basename>.dl
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "usage: go run . <bucket> <key> <local-file>")
		os.Exit(2)
	}
	bucket, key, src := os.Args[1], os.Args[2], os.Args[3]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := storage.NewClient(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "new client:", err)
		os.Exit(1)
	}
	defer client.Close()

	obj := client.Bucket(bucket).Object(key)

	// PART 1 — upload. Note: Close() is what commits; deferring Close hides
	// real upload errors. Call it explicitly and check the error.
	// TODO: f, err := os.Open(src)
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "open:", err); os.Exit(1) }
	// TODO: defer f.Close()
	// TODO: w := obj.NewWriter(ctx)
	// TODO: if _, err := io.Copy(w, f); err != nil {
	// TODO:     _ = w.Close() // best-effort
	// TODO:     fmt.Fprintln(os.Stderr, "copy to writer:", err); os.Exit(1)
	// TODO: }
	// TODO: if err := w.Close(); err != nil {
	// TODO:     fmt.Fprintln(os.Stderr, "writer close (commit):", err); os.Exit(1)
	// TODO: }
	// TODO: fmt.Printf("uploaded gs://%s/%s\n", bucket, key)

	// PART 2 — download to /tmp/<basename>.dl.
	dst := filepath.Join("/tmp", filepath.Base(src)+".dl")
	// TODO: r, err := obj.NewReader(ctx)
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "new reader:", err); os.Exit(1) }
	// TODO: defer r.Close()
	// TODO: out, err := os.Create(dst)
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "create:", err); os.Exit(1) }
	// TODO: defer out.Close()
	// TODO: n, err := io.Copy(out, r)
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "copy from reader:", err); os.Exit(1) }
	// TODO: fmt.Printf("downloaded %d bytes to %s\n", n, dst)

	_ = obj
	_ = dst
	_ = bucket
	_ = key
	_ = io.Copy
}
