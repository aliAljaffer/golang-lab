package main

import (
	"fmt"
	"os"
)

type info struct {
	Module string `json:"module"`
	Go     string `json:"go"`
	Deps   int    `json:"deps"`
	GOOS   string `json:"goos"`
	GOARCH string `json:"goarch"`
}

// parseGoMod extracts module path, go version, and direct dependency count from go.mod content.
// TODO: implement.
//
// Hints:
//   - the "module" line starts with `module ` followed by the path
//   - the "go" line starts with `go ` followed by a version like 1.22
//   - dependencies appear inside `require ( ... )` blocks (count non-comment, non-blank lines)
//     or as single-line `require pkg v1.2.3` declarations
func parseGoMod(content []byte) (info, error) {
	return info{}, fmt.Errorf("not implemented")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: gostat [--json] <path-to-go.mod>")
		os.Exit(2)
	}
	// TODO: parse --json flag, read file, call parseGoMod, print (table or JSON)
	fmt.Fprintln(os.Stderr, "not implemented yet — see README.md and main.go TODOs")
	os.Exit(1)
}
