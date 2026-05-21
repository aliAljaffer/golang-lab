// 01-basic-get — the simplest possible HTTP client.
//
// Goal: GET a URL, read the whole body, print it.
//
// Run:
//   go run . https://httpbin.org/get
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: 01-basic-get <url>")
		os.Exit(2)
	}
	url := os.Args[1]

	// TODO: resp, err := http.Get(url). Handle the error.
	// TODO: defer resp.Body.Close() — even on non-2xx responses you still own the body.
	// TODO: print resp.Status and resp.Header.Get("Content-Type").
	// TODO: body, err := io.ReadAll(resp.Body). Print body to stdout.

	_ = url
	_ = http.Get
	_ = io.ReadAll
}
