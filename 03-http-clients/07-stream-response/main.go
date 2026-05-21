// 07-stream-response — process a large response without buffering it in memory.
//
// Goal: GET a large file and count the number of lines, without ever reading
// the whole body into memory. We use bufio.Scanner on resp.Body.
//
// We hit https://norvig.com/big.txt — Peter Norvig's spell-checker corpus
// (~6.5MB, hundreds of thousands of lines). Small enough not to be rude,
// big enough that you can see the streaming pattern matter.
//
// Run:
//   go run .
package main

import (
	"bufio"
	"fmt"
	"net/http"
	"time"
)

func main() {
	const url = "https://norvig.com/big.txt"

	client := &http.Client{Timeout: 30 * time.Second}

	// TODO: resp, err := client.Get(url). Handle error. defer Body.Close().
	// TODO: guard non-200 status.
	// TODO: scanner := bufio.NewScanner(resp.Body)
	//       lines := 0
	//       for scanner.Scan() { lines++ }
	//       if err := scanner.Err(); err != nil { ... }
	//       fmt.Printf("lines: %d\n", lines)
	//
	// Why this matters: the alternative — io.ReadAll(resp.Body) — would pull
	// the entire 6.5MB into memory before you process a single byte. For
	// gigabyte-sized payloads that strategy OOMs your process.

	_ = client
	_ = url
	_ = bufio.NewScanner
	_ = fmt.Println
}
