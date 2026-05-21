// 04-headers-auth — custom request headers (Authorization, User-Agent).
//
// Goal: build an *http.Request with http.NewRequest, set Authorization and
// User-Agent headers, then send it through http.Client.Do.
//
// We hit https://httpbin.org/headers — it echoes back the headers it saw,
// so you can verify your headers landed.
//
// Run:
//
//	GH_TOKEN=ghp_xxx go run .       # token is optional, fake token works
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	const url = "https://httpbin.org/headers"

	token := os.Getenv("GH_TOKEN")
	if token == "" {
		token = "fake-token-for-demo"
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// TODO: req, err := http.NewRequest(http.MethodGet, url, nil). Handle error.
	// TODO: req.Header.Set("Authorization", "Bearer " + token)
	// TODO: req.Header.Set("User-Agent", "golang-lab/0.1 (https://github.com/alialjaffer/golang-lab)")
	// TODO: req.Header.Set("Accept", "application/json")
	// TODO: resp, err := client.Do(req). Handle error. defer Body.Close().
	// TODO: io.Copy(os.Stdout, resp.Body) — print what the server saw.

	_ = client
	_ = url
	_ = token
	_ = io.Copy
	_ = http.NewRequest
	_ = fmt.Println
}
