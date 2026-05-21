// 02-flag-basics — stdlib `flag` package.
//
// Goal: build a tiny "greet" CLI:
//   go run . --name Ali --shout --count 3
//   -> "HELLO, ALI!" printed 3 times
//
// Demonstrates: flag.String, flag.Bool, flag.Int, flag.Parse, default values, flag.Usage.
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	// TODO: declare flags with sensible defaults.
	//   name  := flag.String("name", "world", "who to greet")
	//   shout := flag.Bool("shout", false, "uppercase the greeting")
	//   count := flag.Int("count", 1, "how many times to print")

	// TODO: customize flag.Usage to print a friendlier help message to os.Stderr.

	// TODO: call flag.Parse().

	// TODO: validate — if *count < 1, print error to stderr and exit 2.

	// TODO: build the greeting, optionally uppercase, print *count times.

	_ = flag.Parse
	_ = fmt.Println
	_ = os.Exit
}
