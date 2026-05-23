// image-pruner — finds and removes Docker images by policy:
//
//	1. Untagged ("dangling") images.
//	2. Images older than --max-age.
//	3. Images with no running OR stopped containers referencing them.
//
// --dry-run prints what would be removed without calling RemoveImage.
// --force passes Force=true on removal (deletes even if referenced by stopped containers).
//
// Testable surface lives at the top of this file. CLI wiring at the bottom.
//
// Run:
//
//	go run . --dry-run
//	go run . --untagged --max-age 168h
//	go run . --no-containers --force
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"

	"github.com/spf13/cobra"
)

// DockerAPI is the slice of *client.Client this tool needs. Tests pass a fake;
// a real `*client.Client` satisfies the interface naturally.
type DockerAPI interface {
	ImageList(ctx context.Context, opts image.ListOptions) ([]image.Summary, error)
	ImageRemove(ctx context.Context, id string, opts image.RemoveOptions) ([]image.DeleteResponse, error)
	ContainerList(ctx context.Context, opts container.ListOptions) ([]container.Summary, error)
}

// Policy controls which images Plan flags for removal.
//
// At least one of the three Removable flags must be true; otherwise Plan
// returns an empty plan with no error (a no-op is still a successful run).
//
// When --max-age is non-zero, RemoveOlderThan is the cutoff (older = remove).
// When --no-containers is set, RemoveUnused excludes images referenced by ANY
// container (running OR stopped); the daemon will refuse to remove them
// without --force anyway, so this is the conservative pre-filter.
type Policy struct {
	RemoveUntagged     bool
	RemoveOlderThan    time.Duration // zero = ignore age
	RemoveUnused       bool          // unused = no container (running OR stopped) references it
	DryRun             bool
	Force              bool
}

// Plan returns the IDs that would be removed under p, sorted lexicographically
// for deterministic output. now is the wall clock (injected for tests).
//
// Plan does NOT mutate the daemon. It's pure given the inputs returned by
// the API. Sync calls Plan then issues ImageRemove for each ID, unless DryRun.
func Plan(images []image.Summary, containers []container.Summary, p Policy, now time.Time) []string {
	// TODO: produce the list of IDs to remove. The three policy rules are OR'd
	//   per image — any one match flags it. The "unused" check is the only one
	//   that needs the container list (build a set of referenced ImageIDs).
	//   Sort the result lexicographically — the tests assert deterministic
	//   ordering and so does any reasonable CLI user.
	_ = sort.Strings
	return nil
}

// isUntagged returns true if img has no non-dangling tags. The Docker SDK
// represents dangling images as having `RepoTags` of `[]string{"<none>:<none>"}`
// (or sometimes nil). Either counts as untagged.
func isUntagged(img image.Summary) bool {
	// TODO: detect dangling. The daemon represents that as either an empty
	//   RepoTags or RepoTags == ["<none>:<none>"]. Both shapes must count.
	return false
}

// Sync executes the plan against api. Returns the IDs that were (or would be)
// removed. In dry-run mode, no ImageRemove calls happen.
func Sync(ctx context.Context, api DockerAPI, p Policy, now time.Time, out io.Writer) ([]string, error) {
	// TODO: list images + containers (both with All:true so stopped containers
	//   count for the unused-reference check), feed Plan, then either print
	//   or remove per p.DryRun. The doc comment above pins the output prefix
	//   ("would remove" vs "removed") via actionPrefix. Wrap remove errors
	//   with the ID so the user knows which one failed.
	return nil, errors.New("Sync not implemented")
}

func actionPrefix(dryRun bool) string {
	if dryRun {
		return "would remove"
	}
	return "removed"
}

// ---- cobra wiring (not unit-tested) ----------------------------------------

type runOpts struct {
	Untagged     bool
	MaxAge       time.Duration
	NoContainers bool
	DryRun       bool
	Force        bool
}

func newRootCmd() *cobra.Command {
	var opts runOpts
	cmd := &cobra.Command{
		Use:   "image-pruner",
		Short: "Prune Docker images by policy (untagged / older-than / unused)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
			if err != nil {
				return fmt.Errorf("build client: %w", err)
			}
			defer cli.Close()

			p := Policy{
				RemoveUntagged:  opts.Untagged,
				RemoveOlderThan: opts.MaxAge,
				RemoveUnused:    opts.NoContainers,
				DryRun:          opts.DryRun,
				Force:           opts.Force,
			}
			_, err = Sync(cmd.Context(), cli, p, time.Now(), cmd.OutOrStdout())
			return err
		},
	}
	cmd.Flags().BoolVar(&opts.Untagged, "untagged", false, "remove dangling/untagged images")
	cmd.Flags().DurationVar(&opts.MaxAge, "max-age", 0, "remove images older than this (e.g. 168h)")
	cmd.Flags().BoolVar(&opts.NoContainers, "no-containers", false, "remove images with no container references")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "print what would be removed; do not mutate")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "pass Force=true to RemoveImage")
	return cmd
}

func main() {
	if err := newRootCmd().ExecuteContext(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
