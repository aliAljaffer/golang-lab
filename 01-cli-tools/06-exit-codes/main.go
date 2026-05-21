// 06-exit-codes — proper exit codes, stderr usage, log.Fatal pitfalls.
//
// Goal: a tiny `divide` CLI:
//   go run . 10 2   -> prints 5 to stdout, exits 0
//   go run . 10 0   -> prints "error: division by zero" to stderr, exits 1
//   go run . 10     -> prints usage to stderr, exits 2
//
// Demonstrates:
//   - stdout (data) vs stderr (diagnostics)
//   - meaningful exit codes (0 OK, 1 runtime error, 2 usage error — convention)
//   - why log.Fatal skips deferred cleanup
package main

import (
	"fmt"
	"os"
)

func main() {
	// TODO: if len(os.Args) != 3, print usage to os.Stderr and os.Exit(2).

	// TODO: parse args[1] and args[2] as ints (strconv.Atoi).
	//       On parse error: stderr + os.Exit(1).

	// TODO: if denominator == 0: stderr "error: division by zero" + os.Exit(1).

	// TODO: print the quotient to stdout. No explicit os.Exit(0) needed — main returning is exit 0.

	// PITFALL: log.Fatal calls os.Exit(1) immediately and does NOT run deferred funcs.
	// If you opened a file with `defer f.Close()` and then log.Fatal, the file leaks.
	// For real programs, prefer returning an error and letting main decide the exit code.

	_ = fmt.Println
	_ = os.Stderr
}
