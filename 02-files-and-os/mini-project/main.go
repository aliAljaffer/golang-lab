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
	// TODO: stat path — if it doesn't exist, return an error.
	// TODO: handle the path+".1" -> path+".2.gz" step (call gzipFile, then os.Remove).
	// TODO: os.Rename(path, path+".1")
	// TODO: create fresh empty file at path with the same perm as the original.
	return fmt.Errorf("rotateOnce: not implemented")
}

// gzipFile reads src and writes a gzip-compressed copy to dst.
// dst must not exist (we don't want to clobber a previous archive silently).
func gzipFile(src, dst string) error {
	// TODO: os.Open(src) — defer close.
	// TODO: os.OpenFile(dst, O_WRONLY|O_CREATE|O_EXCL, 0o644) — defer close.
	// TODO: gzip.NewWriter(dstFile) — defer close.
	// TODO: io.Copy(gzWriter, srcFile)
	return fmt.Errorf("gzipFile: not implemented")
}

// pruneOld deletes files in `dir` whose name starts with `prefix` and ends
// with ".gz" and whose ModTime is older than `now.Add(-keepDays * 24h)`.
// Returns the list of paths it removed (for logging/testing).
//
// `now` is injected so tests can pin the clock.
func pruneOld(dir, prefix string, keepDays int, now time.Time) ([]string, error) {
	// TODO: if keepDays <= 0, return (nil, nil) — "0 = keep everything".
	// TODO: cutoff := now.AddDate(0, 0, -keepDays)
	// TODO: os.ReadDir(dir) — for each entry:
	//         - skip dirs
	//         - skip names not matching prefix + "*.gz"
	//         - entry.Info() for ModTime; if ModTime is before cutoff, os.Remove and record the path
	// TODO: return removed list + nil (or the first error).
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
			// TODO: call rotateOnce(file)
			// TODO: if keepDays > 0:
			//         dir := filepath.Dir(file)
			//         prefix := filepath.Base(file) + "."
			//         _, err := pruneOld(dir, prefix, keepDays, time.Now())
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
