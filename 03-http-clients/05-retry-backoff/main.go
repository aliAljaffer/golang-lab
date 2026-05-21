// 05-retry-backoff — retry on 5xx / network errors with exponential backoff + jitter.
//
// Goal: write a small helper `getWithRetry(client, url, maxAttempts)` that:
//   - retries on transport errors and on 5xx responses
//   - does NOT retry on 4xx (those are "your fault, won't change on retry")
//   - waits 2^attempt * baseDelay between tries, plus random jitter
//   - gives up after maxAttempts and returns the last error / response
//
// We exercise it against https://httpbin.org/status/503 (always 503) to watch
// the backoff happen, then against /status/200 to confirm success on first try.
//
// Run:
//   go run .
package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

// getWithRetry GETs url, retrying transport errors and 5xx up to maxAttempts.
// Caller is responsible for closing the returned response Body on success.
func getWithRetry(client *http.Client, url string, maxAttempts int) (*http.Response, error) {
	const baseDelay = 200 * time.Millisecond

	// TODO: loop attempt = 0..maxAttempts-1:
	//   - resp, err := client.Get(url)
	//   - if err == nil && resp.StatusCode < 500 -> return resp, nil  (success OR 4xx)
	//   - if err == nil { resp.Body.Close() } so we don't leak fds on retry
	//   - if attempt == maxAttempts-1 -> return last result/err
	//   - sleep := baseDelay * (1 << attempt)
	//   - jitter := time.Duration(rand.Int63n(int64(sleep / 2)))
	//   - time.Sleep(sleep + jitter)

	_ = baseDelay
	_ = rand.Int63n
	return nil, fmt.Errorf("getWithRetry: not implemented")
}

func main() {
	client := &http.Client{Timeout: 5 * time.Second}

	for _, url := range []string{
		"https://httpbin.org/status/200",
		"https://httpbin.org/status/503",
	} {
		fmt.Printf("--- %s ---\n", url)
		resp, err := getWithRetry(client, url, 4)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			continue
		}
		fmt.Printf("status: %s\n", resp.Status)
		resp.Body.Close()
	}
}
