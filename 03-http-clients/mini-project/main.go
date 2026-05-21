// gh-repo-stats — fetch repo metadata for a list of GitHub repos.
//
// Spec (from ../PLAN.md):
//   gh-repo-stats --repos owner/name,owner2/name2 --cache cache.json -o stats.csv
//
// Behaviour:
//   - Calls https://api.github.com/repos/{owner}/{name} for each repo.
//   - Retries on 503/429 with exponential backoff + jitter.
//   - Sends If-None-Match with the cached ETag so unchanged repos return 304
//     and don't count against the rate limit.
//   - Writes a CSV with columns: name,stars,forks,pushed_at
//
// The cobra wiring is at the bottom; the testable functions are at the top
// so you can drive them from main_test.go without spinning up a real CLI.
package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

const defaultBaseURL = "https://api.github.com"

// Stats is the per-repo record we eventually write to CSV.
type Stats struct {
	Name     string `json:"full_name"`
	Stars    int    `json:"stargazers_count"`
	Forks    int    `json:"forks_count"`
	PushedAt string `json:"pushed_at"`
}

// CacheEntry is what we persist between runs so we can send If-None-Match
// and reuse the previous payload on 304 Not Modified.
type CacheEntry struct {
	ETag  string `json:"etag"`
	Stats Stats  `json:"stats"`
}

// fetchStats GETs {baseURL}/repos/{repo}. If priorETag is non-empty we send
// it as If-None-Match. Returns:
//   - stats, true, nil   when the server returned 200 with fresh data
//   - {},    false, nil  when the server returned 304 (caller should reuse cache)
//   - {},    false, err  on any non-recoverable error
//
// `repo` is "owner/name". `client` should already have a timeout configured.
func fetchStats(client *http.Client, baseURL, repo, priorETag string) (stats Stats, newETag string, fresh bool, err error) {
	url := baseURL + "/repos/" + repo

	// TODO: req, err := http.NewRequest("GET", url, nil)
	// TODO: req.Header.Set("Accept", "application/vnd.github+json")
	// TODO: if priorETag != "" { req.Header.Set("If-None-Match", priorETag) }
	// TODO: resp, err := doWithRetry(client, req, 4)  -- retry helper below
	// TODO: defer resp.Body.Close()
	// TODO:
	//   switch resp.StatusCode {
	//   case http.StatusNotModified:  // 304
	//       return Stats{}, priorETag, false, nil
	//   case http.StatusOK:           // 200
	//       var s Stats
	//       if err := json.NewDecoder(resp.Body).Decode(&s); err != nil { return ..., err }
	//       return s, resp.Header.Get("ETag"), true, nil
	//   default:
	//       return Stats{}, "", false, fmt.Errorf("unexpected status %s for %s", resp.Status, repo)
	//   }

	_ = url
	_ = json.NewDecoder
	return Stats{}, "", false, fmt.Errorf("fetchStats: not implemented")
}

// doWithRetry sends req via client. On transport errors or 5xx/429 responses
// it retries up to maxAttempts times with exponential backoff + jitter.
// The caller still owns resp.Body on the returned (final) response.
func doWithRetry(client *http.Client, req *http.Request, maxAttempts int) (*http.Response, error) {
	const baseDelay = 100 * time.Millisecond

	// TODO: loop attempt=0..maxAttempts-1:
	//   - resp, err := client.Do(req)
	//   - if err == nil and resp.StatusCode != 429 and resp.StatusCode < 500 -> return resp, nil
	//   - if err == nil { resp.Body.Close() } before retrying so we don't leak
	//   - if attempt == maxAttempts-1 -> return last resp/err (caller decides)
	//   - sleep := baseDelay * (1 << attempt) + jitter (up to half of sleep)
	//   - time.Sleep(sleep)

	_ = baseDelay
	_ = rand.Int63n
	return nil, fmt.Errorf("doWithRetry: not implemented")
}

// loadCache reads the JSON-on-disk cache. A missing file is treated as
// "no cache yet" (empty map, no error).
func loadCache(path string) (map[string]CacheEntry, error) {
	// TODO: os.ReadFile(path)
	//       if errors.Is(err, fs.ErrNotExist) { return map[string]CacheEntry{}, nil }
	//       json.Unmarshal into map[string]CacheEntry
	return nil, fmt.Errorf("loadCache: not implemented")
}

// saveCache writes the cache atomically (write tmp -> rename).
func saveCache(path string, c map[string]CacheEntry) error {
	// TODO: marshal `c` with json.MarshalIndent
	// TODO: write to path+".tmp" then os.Rename to path  (atomic replace)
	return fmt.Errorf("saveCache: not implemented")
}

// writeCSV writes rows to w. Header line is "name,stars,forks,pushed_at".
func writeCSV(w io.Writer, rows []Stats) error {
	// TODO: cw := csv.NewWriter(w)
	// TODO: cw.Write the header
	// TODO: for each row: cw.Write([]string{row.Name, strconv.Itoa(row.Stars), ...})
	// TODO: cw.Flush(); return cw.Error()
	_ = csv.NewWriter
	return fmt.Errorf("writeCSV: not implemented")
}

func newRootCmd() *cobra.Command {
	var (
		repos     []string
		cachePath string
		outPath   string
	)

	cmd := &cobra.Command{
		Use:   "gh-repo-stats",
		Short: "fetch stars/forks/last-push for GitHub repos, with cache + retry",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(repos) == 0 {
				return fmt.Errorf("--repos is required (comma-separated owner/name list)")
			}
			client := &http.Client{Timeout: 15 * time.Second}

			// TODO: cache, err := loadCache(cachePath)
			// TODO: for each repo:
			//         prior := cache[repo].ETag
			//         stats, etag, fresh, err := fetchStats(client, defaultBaseURL, repo, prior)
			//         if !fresh && prior != "" { stats = cache[repo].Stats }
			//         cache[repo] = CacheEntry{ETag: etag, Stats: stats}
			//         rows = append(rows, stats)
			// TODO: saveCache(cachePath, cache)
			// TODO: open outPath (or os.Stdout if "-"), call writeCSV(w, rows).

			_ = client
			return fmt.Errorf("gh-repo-stats: not implemented")
		},
	}
	cmd.Flags().StringSliceVar(&repos, "repos", nil, "comma-separated owner/name repos to fetch")
	cmd.Flags().StringVar(&cachePath, "cache", "gh-cache.json", "cache file for ETags + stats")
	cmd.Flags().StringVar(&outPath, "out", "-", "output CSV file (- for stdout)")
	return cmd
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
