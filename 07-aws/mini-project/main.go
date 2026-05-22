// s3sync — mirror a local directory to an S3 bucket.
//
// Spec (from ../PLAN.md):
//   - Walks <local-dir>; for each file computes md5 (hex), compares to the
//     remote object's ETag (which equals md5 for single-PUT uploads).
//     Skips files whose md5 matches.
//   - --dry-run prints the actions but doesn't touch S3.
//   - --delete removes S3 objects whose key is not present locally.
//   - --concurrency N caps in-flight upload/delete operations.
//
// Testable surface (top of the file). cobra wiring is at the bottom.
//
// Run:
//
//	go run . --bucket my-b --dir ./pages --concurrency 4
//	go run . --bucket my-b --dir ./pages --dry-run
//	go run . --bucket my-b --dir ./pages --delete
package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/spf13/cobra"
)

// S3API is the slice of *s3.Client the tool uses. Production passes a real
// client; tests pass a fake (see main_test.go for the pattern).
type S3API interface {
	ListObjectsV2(ctx context.Context, in *s3.ListObjectsV2Input, opts ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
	PutObject(ctx context.Context, in *s3.PutObjectInput, opts ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	DeleteObject(ctx context.Context, in *s3.DeleteObjectInput, opts ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

// LocalFile is one file under the source directory.
type LocalFile struct {
	Path string // absolute on disk
	Key  string // forward-slash relative to root (the S3 key)
	MD5  string // hex
}

// Action is one planned change. Op is "upload", "delete", or "skip".
type Action struct {
	Op  string
	Key string
}

// Options bundle inputs for Sync. Keeping them in a struct keeps the call
// site honest as flags accumulate.
type Options struct {
	Bucket      string
	Dir         string
	DryRun      bool
	Delete      bool
	Concurrency int
}

// WalkLocal walks <root>, returning one LocalFile per regular file. Keys are
// relative to root, with backslashes converted to forward slashes (Windows
// safety even on macOS, since S3 keys must be forward-slash).
func WalkLocal(root string) ([]LocalFile, error) {
	// TODO: implement. Use filepath.Walk or filepath.WalkDir. For each regular
	// TODO: file, compute its md5 by hashing the file contents (see computeMD5
	// TODO: below), and build a key as filepath.Rel(root, path) with backslashes
	// TODO: replaced by forward slashes.
	return nil, errors.New("WalkLocal not implemented")
}

// computeMD5 returns the hex md5 of f's contents. Closes nothing — caller owns
// the file handle.
func computeMD5(f io.Reader) (string, error) {
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// ListRemote returns a map of key → unquoted ETag (md5) for every object in
// <bucket>. Pages internally.
func ListRemote(ctx context.Context, api S3API, bucket string) (map[string]string, error) {
	// TODO: page with s3.NewListObjectsV2Paginator(api, &s3.ListObjectsV2Input{Bucket: &bucket}).
	// TODO: for each object, store {*obj.Key: strings.Trim(*obj.ETag, "\"")}.
	// TODO: return the map.
	return nil, errors.New("ListRemote not implemented")
}

// Plan diffs locals against remotes and returns the action list. Order is
// deterministic (sorted by key) so callers / tests can read it.
//
// Rules:
//   - local key not in remote  → upload
//   - local key in remote with different md5 → upload
//   - local key in remote with same md5 → skip
//   - remote key not in local + opts.Delete → delete
//   - remote key not in local + !opts.Delete → omitted
func Plan(locals []LocalFile, remotes map[string]string, opts Options) []Action {
	// TODO: implement.
	return nil
}

// Sync executes the plan with at most opts.Concurrency operations in flight.
// Returns the count of uploaded + deleted + skipped actions and the first
// error encountered (other actions still drain — callers see the count).
//
// If opts.DryRun is true, no S3 calls are made — actions are returned as-is.
func Sync(ctx context.Context, api S3API, opts Options) (uploaded, deleted, skipped int, err error) {
	// TODO: locals, _ := WalkLocal(opts.Dir)
	// TODO: remotes, _ := ListRemote(ctx, api, opts.Bucket)
	// TODO: plan := Plan(locals, remotes, opts)

	// TODO: if opts.DryRun, count actions by op and return.

	// TODO: otherwise spin up opts.Concurrency goroutines reading from a chan Action.
	// TODO: For "upload", os.Open the file, PutObject. For "delete", DeleteObject.
	// TODO: For "skip", increment skipped.
	// TODO: collect first error via sync.Once or a guarded var.

	_ = sync.WaitGroup{}
	_ = atomic.Int32{}
	return 0, 0, 0, errors.New("Sync not implemented")
}

// helper: avoid lint about unused imports during scaffolding.
var (
	_ = s3types.Object{}
	_ = strings.TrimPrefix
	_ = filepath.Walk
	_ = time.Second
	_ = fmt.Println
)

// ---- cobra wiring (not unit-tested) ----------------------------------------

func newRootCmd() *cobra.Command {
	opts := Options{Concurrency: 4}

	cmd := &cobra.Command{
		Use:   "s3sync",
		Short: "Mirror a local directory to an S3 bucket",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if opts.Bucket == "" {
				return errors.New("--bucket is required")
			}
			if opts.Dir == "" {
				return errors.New("--dir is required")
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
			defer cancel()

			cfg, err := config.LoadDefaultConfig(ctx)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			client := s3.NewFromConfig(cfg)

			up, del, skip, err := Sync(ctx, client, opts)
			fmt.Fprintf(cmd.OutOrStdout(), "uploaded=%d deleted=%d skipped=%d\n", up, del, skip)
			return err
		},
	}
	cmd.Flags().StringVar(&opts.Bucket, "bucket", "", "destination S3 bucket")
	cmd.Flags().StringVar(&opts.Dir, "dir", "", "local directory to mirror")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "plan only, don't modify the bucket")
	cmd.Flags().BoolVar(&opts.Delete, "delete", false, "remove S3 objects not present locally")
	cmd.Flags().IntVar(&opts.Concurrency, "concurrency", 4, "max parallel operations")
	return cmd
}

func main() {
	if err := newRootCmd().ExecuteContext(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
