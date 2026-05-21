// 04-exec — run a subprocess, capture stdout and stderr separately.
//
// Goal: run `ls -la` (or `dir` on Windows; ignore Windows for now), print
// stdout, print stderr, and report the exit code on failure.
//
// Run:
//   go run .
//   go run . /no/such/path   # forces a non-zero exit
package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

func main() {
	target := "."
	if len(os.Args) > 1 {
		target = os.Args[1]
	}

	// TODO: cmd := exec.Command("ls", "-la", target)
	// TODO: var stdout, stderr bytes.Buffer
	// TODO: cmd.Stdout = &stdout; cmd.Stderr = &stderr
	// TODO: err := cmd.Run()
	// TODO: print stdout.String() to os.Stdout, stderr.String() to os.Stderr.
	// TODO: if err != nil, try a type assertion to *exec.ExitError and print
	//       its ExitCode(). Exit with the same code (or 1 for other errors).

	_ = target
	_ = bytes.Buffer{}
	_ = exec.Command
	_ = fmt.Println
}
