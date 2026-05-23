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

	// TODO: implement the retry loop. The doc above defines what counts as
	//   retryable (429 + 5xx + transport errors). The load-bearing detail
	//   the tests check: when you retry, you MUST close the previous
	//   response's body — otherwise you leak the connection and the test
	//   server's accept loop blocks waiting for it. Backoff = baseDelay <<
	//   attempt + jitter.

	_ = baseDelay
	_ = rand.Int63n
	return nil, fmt.Errorf("DoWithRetry: not implemented")
}
