// dirsize — walk a directory and report sizes per immediate subdirectory.
//
// Spec (from ../PLAN.md):
//   dirsize PATH            # human-readable, sorted descending by size
//   dirsize PATH --top 3    # show only the 3 largest
//   dirsize PATH --json     # emit JSON: [{"path":"...","bytes":1234}, ...]
//
// Exit codes:
//   0  success
//   1  PATH missing or not a directory
//   2  usage error (bad flags)
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Entry is the per-subdirectory result. Exported so encoding/json can marshal it
// and so tests in this package can construct expected values.
type Entry struct {
	Path  string `json:"path"`
	Bytes int64  `json:"bytes"`
}

// scan returns one Entry per immediate child directory of root, with Bytes
// equal to the recursive sum of regular-file sizes under that child.
//
// Implementation tips:
//   - os.ReadDir(root) gives you immediate children.
//   - filepath.WalkDir(child) walks recursively; skip non-regular files.
//   - Return ([]Entry, error). Caller decides exit code.
func scan(root string) ([]Entry, error) {
	// TODO: implement.
	return nil, fmt.Errorf("scan: not implemented")
}

// sortAndTrim sorts entries descending by Bytes (stable on Path for ties) and,
// if top > 0, truncates to the first `top` entries.
func sortAndTrim(entries []Entry, top int) []Entry {
	// TODO: implement using sort.SliceStable.
	return entries
}

// renderText prints entries in human-readable form to w.
//   1.2 MB  ./node_modules
//   340 KB  ./src
func renderText(entries []Entry) string {
	// TODO: implement. humanize the byte count (1024-based: KB, MB, GB...).
	return ""
}

// renderJSON marshals entries as a JSON array (no trailing newline).
func renderJSON(entries []Entry) (string, error) {
	// TODO: implement using encoding/json.
	return "", nil
}

func newRootCmd() *cobra.Command {
	var top int
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "dirsize PATH",
		Short: "report per-subdirectory sizes",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO:
			//   1. stat args[0] — if missing or not a dir, return an error (-> exit 1).
			//   2. entries, err := scan(args[0])
			//   3. entries = sortAndTrim(entries, top)
			//   4. if asJSON: print renderJSON; else print renderText.
			return fmt.Errorf("dirsize: not implemented")
		},
	}

	cmd.Flags().IntVar(&top, "top", 0, "show only the N largest (0 = all)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "emit JSON instead of human text")
	return cmd
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
