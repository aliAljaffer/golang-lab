// logrotate — rotate a log file: rename → gzip → prune.
//
// Spec (from ../PLAN.md):
//   logrotate --file PATH
//   logrotate --file PATH --keep-days 7
//
// Exit codes:
//   0  success
//   1  I/O error
//   2  flag misuse (handled by cobra)
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// rotateOnce performs one rotation cycle for `path`:
//   1. if path+".1" exists: gzip it to path+".2.gz", then remove path+".1".
//   2. rename path -> path+".1"
//   3. create a fresh empty file at path with the original mode.
//
// Returns the first error encountered.
func rotateOnce(path string) error {
	// TODO: walk the three-step rotation described above. Two non-obvious
	//   bits the test pins:
	//     - the fresh file at `path` must keep the original mode (otherwise
	//       a 0600 log file rotates into a world-readable one).
	//     - the previous rotated copy (path+".1") MUST be gzipped to .2.gz
	//       BEFORE the rename, or you'll lose its content when .1 is
	//       overwritten.
	return fmt.Errorf("rotateOnce: not implemented")
}

// gzipFile reads src and writes a gzip-compressed copy to dst.
// dst must not exist (we don't want to clobber a previous archive silently).
func gzipFile(src, dst string) error {
	// TODO: stream src into a new gzip writer pointed at dst. Use the
	//   exclusive-create flag (O_EXCL) so an existing dst is an error, not
	//   a silent overwrite. The gzip writer MUST be Close'd to flush its
	//   trailer — a missing trailer makes the file unreadable.
	return fmt.Errorf("gzipFile: not implemented")
}

// pruneOld deletes files in `dir` whose name starts with `prefix` and ends
// with ".gz" and whose ModTime is older than `now.Add(-keepDays * 24h)`.
// Returns the list of paths it removed (for logging/testing).
//
// `now` is injected so tests can pin the clock.
func pruneOld(dir, prefix string, keepDays int, now time.Time) ([]string, error) {
	// TODO: keepDays <= 0 is the "keep everything" sentinel — early return
	//   with no error. Otherwise scan dir, filter by prefix + ".gz" suffix,
	//   and delete anything whose mtime is older than the cutoff.
	return nil, fmt.Errorf("pruneOld: not implemented")
}

func newRootCmd() *cobra.Command {
	var file string
	var keepDays int

	cmd := &cobra.Command{
		Use:   "logrotate",
		Short: "rotate a log file: rename, gzip, prune",
		RunE: func(cmd *cobra.Command, args []string) error {
			if file == "" {
				return fmt.Errorf("--file is required")
			}
			// TODO: rotateOnce, then prune. The prefix passed to pruneOld is
			//   filepath.Base(file) + "." so it only matches THIS log's
			//   rotated copies, not anything else in the directory.
			return fmt.Errorf("logrotate: not implemented")
		},
	}
	cmd.Flags().StringVar(&file, "file", "", "log file to rotate (required)")
	cmd.Flags().IntVar(&keepDays, "keep-days", 0, "delete rotated *.gz older than N days (0 = keep all)")
	return cmd
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
