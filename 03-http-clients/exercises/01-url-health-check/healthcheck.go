// 01-url-health-check — given a list of URLs, return which are 2xx
// and how long each took.
//
// You implement: CheckAll(client *http.Client, urls []string) []Result.
// The tests in healthcheck_test.go drive the design.
package healthcheck

import (
	"net/http"
	"time"
)

// Result is one URL's outcome.
type Result struct {
	URL      string
	Status   int           // 0 if the request failed before getting a response
	OK       bool          // true iff Status is in 200..299
	Duration time.Duration // wall-clock time to first response / failure
	Err      error         // non-nil on transport errors (DNS, refused, timeout)
}

// CheckAll runs GET against each URL using `client` and returns one Result
// per URL, in the same order as the input. Order is preserved even if you
// implement this concurrently.
//
// Transport errors do not abort the run — each URL gets its own Result.
func CheckAll(client *http.Client, urls []string) []Result {
	// TODO: loop over urls. For each:
	//         started := time.Now()
	//         resp, err := client.Get(url)
	//         elapsed := time.Since(started)
	//         build Result. If err != nil, Status=0, OK=false, Err=err.
	//         else Status=resp.StatusCode, OK=200<=Status<300, resp.Body.Close().
	// TODO: return results in the same order as `urls`.
	return nil
}
