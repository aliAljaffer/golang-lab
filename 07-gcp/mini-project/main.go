// gcssync — mirror a local directory to a GCS bucket.
//
// Spec (from ../PLAN.md):
//   - Walks <local-dir>; for each file computes CRC32C (Castagnoli table)
//     and compares to the remote object's CRC32C. Skips files whose CRC matches.
//   - --dry-run prints the actions but doesn't touch GCS.
//   - --delete removes GCS objects whose name is not present locally.
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
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spf13/cobra"
)

// GCSAPI is the slice of GCS the tool uses. Production wires it to a real
// *storage.Client (see ../07-mocking-gcs/ for the wrapper pattern); tests
// pass a fakeGCS (see main_test.go). Three methods — same shape as
// `gcsutil.GCSAPI` in 07-mocking-gcs.
type GCSAPI interface {
	List(ctx context.Context, bucket, prefix string) ([]RemoteObject, error)
	Upload(ctx context.Context, bucket, key string, body io.Reader) error
	Delete(ctx context.Context, bucket, key string) error
}

// RemoteObject is the minimal projection ListObjects returns.
type RemoteObject struct {
	Name   string
	Size   int64
	CRC32C uint32
}

// LocalFile is one file under the source directory.
type LocalFile struct {
	Path   string // absolute on disk
	Key    string // forward-slash relative to root (the GCS object name)
	CRC32C uint32 // Castagnoli CRC32 of the file contents
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

// castagnoli is the polynomial GCS uses. *Always* use this — the default
// IEEE table will compute a different hash and you'll re-upload every
// object every run.
var castagnoli = crc32.MakeTable(crc32.Castagnoli)

// WalkLocal walks <root>, returning one LocalFile per regular file. Keys are
// relative to root, with backslashes converted to forward slashes (Windows
// safety even on macOS, since GCS object names must use forward slashes).
//
// The CRC32C is computed by hashing the file contents with the Castagnoli
// table — same one GCS uses server-side.
func WalkLocal(root string) ([]LocalFile, error) {
	// TODO: walk <root> with filepath.WalkDir; skip directories.
	// TODO: for each regular file: open it, compute crc32.Update over its
	// TODO: bytes using the `castagnoli` table, build a forward-slash key
	// TODO: with filepath.Rel(root, path), append a LocalFile.
	return nil, errors.New("WalkLocal not implemented")
}

// computeCRC32C returns the Castagnoli CRC32 of r's contents.
func computeCRC32C(r io.Reader) (uint32, error) {
	h := crc32.New(castagnoli)
	if _, err := io.Copy(h, r); err != nil {
		return 0, err
	}
	return h.Sum32(), nil
}

// ListRemote drains the bucket listing into a {name: RemoteObject} map for
// O(1) lookups in Plan. Uses the GCSAPI's List under the hood.
func ListRemote(ctx context.Context, api GCSAPI, bucket string) (map[string]RemoteObject, error) {
	// TODO: call api.List(ctx, bucket, ""). For each returned RemoteObject,
	// TODO: store it under its Name in the map. Return the map.
	return nil, errors.New("ListRemote not implemented")
}

// Plan diffs locals against remotes and returns the action list. Order is
// deterministic (sorted by key) so callers / tests can read it.
//
// Rules:
//   - local key not in remote                       → upload
//   - local key in remote with different CRC32C     → upload
//   - local key in remote with same CRC32C          → skip
//   - remote key not in local + opts.Delete         → delete
//   - remote key not in local + !opts.Delete        → omitted
func Plan(locals []LocalFile, remotes map[string]RemoteObject, opts Options) []Action {
	// TODO: implement. Return actions sorted by Key for determinism.
	return nil
}

// Sync executes the plan with at most opts.Concurrency operations in flight.
// Returns the count of uploaded + deleted + skipped actions and the first
// error encountered (other actions still drain — callers see the count).
//
// If opts.DryRun is true, no GCS calls are made — actions are returned as-is.
func Sync(ctx context.Context, api GCSAPI, opts Options) (uploaded, deleted, skipped int, err error) {
	// TODO: locals, err := WalkLocal(opts.Dir); return on err.
	// TODO: remotes, err := ListRemote(ctx, api, opts.Bucket); return on err.
	// TODO: plan := Plan(locals, remotes, opts).

	// TODO: if opts.DryRun, count actions by op and return.

	// TODO: otherwise: bounded worker pool — semaphore-channel (buffered chan
	// TODO: of size opts.Concurrency) OR a fixed pool reading from a job chan.
	// TODO: For "upload", os.Open the file, api.Upload(...). For "delete",
	// TODO: api.Delete(...). For "skip", increment skipped.
	// TODO: collect first error via sync.Once or guarded var; other workers
	// TODO: continue so counts are accurate.

	_ = sync.WaitGroup{}
	_ = atomic.Int32{}
	return 0, 0, 0, errors.New("Sync not implemented")
}

// helper: avoid lint about unused imports during scaffolding.
var (
	_ = strings.TrimPrefix
	_ = filepath.Walk
	_ = time.Second
	_ = fmt.Println
)

// ---- cobra wiring (not unit-tested) ----------------------------------------

func newRootCmd() *cobra.Command {
	opts := Options{Concurrency: 4}

	cmd := &cobra.Command{
		Use:   "gcssync",
		Short: "Mirror a local directory to a GCS bucket",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if opts.Bucket == "" {
				return errors.New("--bucket is required")
			}
			if opts.Dir == "" {
				return errors.New("--dir is required")
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
			defer cancel()

			// In production wire a RealGCS from ../07-mocking-gcs/ here.
			// For now the binary is a scaffold — the real GCSAPI wiring is
			// the last thing you'd add once Sync/Plan/etc. pass tests.
			// TODO: api, err := gcsutil.NewReal(ctx); if err != nil { return err }
			// TODO: defer api.Client.Close()
			// TODO: up, del, skip, err := Sync(ctx, api, opts)
			var (
				up, del, skip int
				err           error
			)
			err = errors.New("wire GCSAPI in main.go: see TODO above")

			fmt.Fprintf(cmd.OutOrStdout(), "uploaded=%d deleted=%d skipped=%d\n", up, del, skip)
			_ = ctx
			return err
		},
	}
	cmd.Flags().StringVar(&opts.Bucket, "bucket", "", "destination GCS bucket")
	cmd.Flags().StringVar(&opts.Dir, "dir", "", "local directory to mirror")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "plan only, don't modify the bucket")
	cmd.Flags().BoolVar(&opts.Delete, "delete", false, "remove GCS objects not present locally")
	cmd.Flags().IntVar(&opts.Concurrency, "concurrency", 4, "max parallel operations")
	return cmd
}

func main() {
	if err := newRootCmd().ExecuteContext(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
