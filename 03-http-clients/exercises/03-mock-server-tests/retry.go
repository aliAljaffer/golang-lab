// 03-mock-server-tests — practice testing real HTTP behaviour using
// httptest.NewServer instead of mocking the client.
//
// This exercise gives you a retry helper to implement. The tests
// (retry_test.go) spin up tiny in-process servers that return controlled
// sequences of 503/429/200, then assert your retry loop:
//   - retries the right things
//   - gives up after N attempts
//   - doesn't leak response bodies between retries
//   - propagates the final (non-retryable) error to the caller
package mocktest

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

// DoWithRetry sends req via client. It retries on:
//   - transport errors (err != nil from client.Do)
//   - response codes 429 (Too Many Requests) and 5xx
// up to maxAttempts total attempts.
//
// Backoff: baseDelay * 2^attempt + small jitter.
//
// The returned *http.Response is the result of the FINAL attempt; the caller
// owns its Body. On retried attempts, the helper MUST close those bodies so
// they don't leak file descriptors.
func DoWithRetry(client *http.Client, req *http.Request, maxAttempts int) (*http.Response, error) {
	const baseDelay = 50 * time.Millisecond

	if maxAttempts < 1 {
		return nil, fmt.Errorf("maxAttempts must be >= 1")
	}

	// TODO: for attempt := 0; attempt < maxAttempts; attempt++ {
	//          resp, err := client.Do(req)
	//          retry := err != nil || resp.StatusCode == 429 || resp.StatusCode >= 500
	//          if !retry {
	//              return resp, nil
	//          }
	//          if resp != nil { resp.Body.Close() }   // <-- don't leak!
	//          if attempt == maxAttempts-1 { return resp, err }
	//          sleep := baseDelay * (1 << attempt)
	//          jitter := time.Duration(rand.Int63n(int64(sleep) / 2 + 1))
	//          time.Sleep(sleep + jitter)
	//       }
	//       return nil, fmt.Errorf("unreachable")

	_ = baseDelay
	_ = rand.Int63n
	return nil, fmt.Errorf("DoWithRetry: not implemented")
}
