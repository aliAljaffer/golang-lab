// 01-os-args — the rawest form of CLI input.
//
// Goal: print os.Args, then implement a tiny "echo" that joins args[1:] with spaces.
// Bonus: if the first arg is "--upper", uppercase the rest.
//
// Run:
//   go run . hello world
//   go run . --upper hello world
package main

import (
	"fmt"
	"os"
)

func main() {
	// TODO: print len(os.Args) and each element with its index.
	// TODO: if no extra args, print "usage: os-args [--upper] WORDS..." to stderr and exit 2.
	// TODO: handle the --upper flag manually (no flag package yet).
	// TODO: print the joined result to stdout.

	_ = os.Args
	_ = fmt.Println
}
