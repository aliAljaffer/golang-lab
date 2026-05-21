// 03-walk — find files matching a suffix under a directory tree.
//
// Goal: walk the current directory and print every path ending in ".go".
//
// Run:
//   go run . .
//   go run . ..
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	// TODO: filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error { ... })
	//   - if err != nil, return it (propagates and stops the walk)
	//   - if d.IsDir(), return nil
	//   - if strings.HasSuffix(path, ".go"), print it
	//   - skip directories named "node_modules" or ".git" with: return filepath.SkipDir

	// TODO: handle the error returned by WalkDir.

	_ = root
	_ = fs.DirEntry(nil)
	_ = filepath.WalkDir
	_ = strings.HasSuffix
	_ = fmt.Println
}
