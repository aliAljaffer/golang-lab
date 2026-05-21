// 07-worker-pool — bounded worker pool: jobs channel + N workers + results channel.
//
// Goal: process M jobs with at most N concurrent workers. The pattern shows
// up everywhere — image processing, HTTP fanout, batch DB updates.
//
// Run:
//   go run .
package main

import (
	"fmt"
	"sync"
	"time"
)

type Job struct {
	ID int
}

type Result struct {
	JobID int
	Out   string
}

// worker pulls Jobs off `jobs`, simulates work, pushes Results onto `results`.
// It exits when `jobs` is closed *and drained* — the for-range handles that for us.
func worker(id int, jobs <-chan Job, results chan<- Result) {
	for j := range jobs {
		// TODO: time.Sleep(50 * time.Millisecond) // simulate work
		// TODO: results <- Result{JobID: j.ID, Out: fmt.Sprintf("worker-%d handled job %d", id, j.ID)}

		_ = j
		_ = id
		_ = time.Sleep
	}
}

func main() {
	const (
		nWorkers = 3
		nJobs    = 10
	)

	jobs := make(chan Job)
	results := make(chan Result, nJobs)

	// 1. Spawn N workers. They block on `range jobs` until something arrives.
	var wg sync.WaitGroup
	for w := 1; w <= nWorkers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			worker(id, jobs, results)
		}(w)
	}

	// 2. Feed jobs in. Close the channel when done so workers can exit.
	for j := 1; j <= nJobs; j++ {
		jobs <- Job{ID: j}
	}
	close(jobs)

	// 3. Wait for workers to drain, THEN close results so range exits below.
	// Closing results before workers finish would panic on the next send.
	go func() {
		wg.Wait()
		close(results)
	}()

	// 4. Consume results as they arrive.
	for r := range results {
		fmt.Println(r.Out)
	}
}
