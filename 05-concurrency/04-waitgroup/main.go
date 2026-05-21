// 04-waitgroup — wait for N goroutines to finish.
//
// Goal: fan out N pieces of work, block until all are done.
//
// Run:
//   go run .
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	urls := []string{
		"https://example.com",
		"https://example.org",
		"https://example.net",
	}

	// Pattern: wg.Add(1) BEFORE the `go` statement, wg.Done() (via defer) inside.
	// Adding inside the goroutine is a race — wg.Wait() may run before Add is called.
	var wg sync.WaitGroup

	// TODO: for _, u := range urls {
	// TODO:     wg.Add(1)
	// TODO:     go func(url string) {
	// TODO:         defer wg.Done()
	// TODO:         time.Sleep(50 * time.Millisecond) // simulate work
	// TODO:         fmt.Println("checked", url)
	// TODO:     }(u)
	// TODO: }
	// TODO: wg.Wait()
	// TODO: fmt.Println("all done")

	_ = wg.Wait
	_ = urls
	_ = time.Sleep
	_ = fmt.Println

	// Why pass `u` as an argument? In Go versions before 1.22, the loop variable was
	// shared across iterations — every goroutine would see the same final value.
	// In 1.22+ the loop variable is per-iteration, so plain capture works too. But
	// passing as an argument is still the clearest, version-independent style.
}
