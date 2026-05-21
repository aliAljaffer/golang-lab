// 02-line-scanner — bufio.Scanner over a file.
//
// Goal: write a small multi-line file, then read it back line-by-line and
// print "<n>: <line>" for each.
//
// Run:
//   go run .
package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	const path = "/tmp/go-learning-02-line-scanner.txt"

	// Seed a file with a few lines.
	if err := os.WriteFile(path, []byte("alpha\nbeta\ngamma\n"), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "seed:", err)
		os.Exit(1)
	}
	defer os.Remove(path)

	// TODO: os.Open(path), defer f.Close(). Handle the error.

	// TODO: scanner := bufio.NewScanner(f). Loop with scanner.Scan().
	//       Track a 1-based line number and print "<n>: <line>".

	// TODO: check scanner.Err() after the loop.

	_ = bufio.NewScanner
}
