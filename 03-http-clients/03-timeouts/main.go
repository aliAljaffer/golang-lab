// 03-timeouts — the single most common Go networking bug.
//
// Goal: demonstrate that http.DefaultClient (used by http.Get) has NO timeout
// and will hang forever on a slow server, while a custom http.Client with a
// Timeout fails fast.
//
// We hit httpbin.org/delay/10 (server sleeps 10s before responding) with two
// clients: the default one and one with a 2s timeout. Watch the second fail.
//
// Run:
//   go run .
package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	const url = "https://httpbin.org/delay/10"

	// TODO: time how long http.DefaultClient takes (it will, eventually, return).
	// Use time.Now() / time.Since(). Print the elapsed duration and any error.
	//
	// (Comment this out in real runs — it really does hang for 10s.)
	//   started := time.Now()
	//   _, err := http.Get(url)
	//   fmt.Printf("default client: err=%v elapsed=%s\n", err, time.Since(started))

	// TODO: build a client with Timeout: 2 * time.Second.
	//   client := &http.Client{Timeout: 2 * time.Second}
	// TODO: client.Get(url), time it, print err + elapsed.
	// The error should mention "context deadline exceeded" or
	// "Client.Timeout exceeded while awaiting headers".

	_ = url
	_ = http.Client{}
	_ = time.Second
	_ = fmt.Println
}
