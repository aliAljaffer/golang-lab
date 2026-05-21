// 06-context-cancel — cancel a request mid-flight with context.
//
// Goal: hit a slow endpoint (httpbin.org/delay/5) and cancel it after 1s using
// context.WithTimeout — independent of any client.Timeout.
//
// Why both `context` and `client.Timeout`?
//   - client.Timeout is a budget for the WHOLE request (set once on the client).
//   - context is per-call and supports propagation: if your handler is called
//     with a cancellable ctx (e.g. the request's ctx), pass it down so cancel
//     fans out all the way to the network read.
//
// Run:
//   go run .
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

func main() {
	const url = "https://httpbin.org/delay/5"

	// TODO: ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	// TODO: defer cancel()
	// TODO: req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	// TODO: started := time.Now()
	// TODO: resp, err := http.DefaultClient.Do(req)
	// TODO: print err + elapsed.
	// The error should wrap context.DeadlineExceeded — check with errors.Is.

	_ = url
	_ = context.Background
	_ = http.NewRequestWithContext
	_ = fmt.Println
	_ = time.Now
}
