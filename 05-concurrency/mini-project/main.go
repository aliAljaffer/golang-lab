// fanout-ping — concurrent URL health checker with bounded parallelism.
//
// Spec (from ../PLAN.md):
//   - Reads URLs from CLI args (or stdin if none).
//   - Checks each with HTTP GET.
//   - At most --concurrency in flight at once (semaphore channel).
//   - --timeout caps each individual request.
//   - Streams results as they arrive (not buffered until the end).
//   - SIGINT/SIGTERM cancels remaining work cleanly.
//
// Testable surface (top of the file). cobra wiring is at the bottom.
//
// Run:
//   go run . --concurrency 4 --timeout 2s \
//     https://example.com https://example.org https://example.net
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

// Result is one URL's outcome.
//
// Status is the HTTP status code (e.g. 200, 503). Status is 0 if the request
// could not complete (timeout, ctx-cancel, DNS error). In that case Err is set.
type Result struct {
	URL      string
	Status   int
	Duration time.Duration
	Err      error
}

// Check performs a single GET against url, using ctx for cancellation.
// On a transport error (timeout, ctx cancel, DNS), Status is 0 and Err is set.
// On a successful HTTP exchange, Status is the response code and Err is nil
// (even for 4xx / 5xx — those are not transport failures).
func Check(ctx context.Context, client *http.Client, url string) Result {
	start := time.Now()
	// TODO: req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	// TODO: if err != nil { return Result{URL: url, Duration: time.Since(start), Err: err} }
	// TODO: resp, err := client.Do(req)
	// TODO: if err != nil { return Result{URL: url, Duration: time.Since(start), Err: err} }
	// TODO: defer resp.Body.Close()
	// TODO: io.Copy(io.Discard, resp.Body) // drain so the connection can be reused
	// TODO: return Result{URL: url, Status: resp.StatusCode, Duration: time.Since(start)}

	_ = start
	_ = http.NewRequestWithContext
	return Result{URL: url, Err: errors.New("Check: not implemented")}
}

// Run fans out Check over urls with at most `concurrency` in flight.
// Results stream out on the returned channel as they arrive (no ordering guarantee).
// The channel is closed once every URL has produced exactly one Result OR ctx is
// cancelled (in which case remaining URLs may produce ctx-cancelled Results).
func Run(ctx context.Context, client *http.Client, urls []string, concurrency int) <-chan Result {
	out := make(chan Result, len(urls))

	if concurrency <= 0 {
		concurrency = 1
	}
	// TODO: sem := make(chan struct{}, concurrency)  // semaphore: at most `concurrency` slots
	// TODO: var wg sync.WaitGroup
	// TODO: for _, u := range urls {
	// TODO:     wg.Add(1)
	// TODO:     go func(url string) {
	// TODO:         defer wg.Done()
	// TODO:         select {
	// TODO:         case sem <- struct{}{}:                // acquire
	// TODO:         case <-ctx.Done():
	// TODO:             out <- Result{URL: url, Err: ctx.Err()}
	// TODO:             return
	// TODO:         }
	// TODO:         defer func() { <-sem }()              // release
	// TODO:         out <- Check(ctx, client, url)
	// TODO:     }(u)
	// TODO: }
	// TODO: go func() { wg.Wait(); close(out) }()

	_ = sync.WaitGroup{}
	close(out)
	return out
}

func newRootCmd() *cobra.Command {
	var (
		concurrency int
		timeout     time.Duration
	)
	cmd := &cobra.Command{
		Use:   "fanout-ping [urls...]",
		Short: "Concurrently health-check URLs with bounded parallelism.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := &http.Client{Timeout: timeout}

			// Parent ctx is cancelled on SIGINT/SIGTERM so in-flight requests abort.
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			results := Run(ctx, client, args, concurrency)
			anyFail := false
			for r := range results {
				if r.Err != nil {
					anyFail = true
					fmt.Printf("FAIL  %s  err=%v  (%s)\n", r.URL, r.Err, r.Duration.Round(time.Millisecond))
					continue
				}
				if r.Status >= 400 {
					anyFail = true
					fmt.Printf("BAD   %s  status=%d  (%s)\n", r.URL, r.Status, r.Duration.Round(time.Millisecond))
					continue
				}
				fmt.Printf("OK    %s  status=%d  (%s)\n", r.URL, r.Status, r.Duration.Round(time.Millisecond))
			}
			if anyFail {
				return errors.New("at least one URL failed or returned >=400")
			}
			return nil
		},
	}
	cmd.Flags().IntVarP(&concurrency, "concurrency", "c", 4, "max requests in flight")
	cmd.Flags().DurationVarP(&timeout, "timeout", "t", 5*time.Second, "per-request timeout")
	return cmd
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
