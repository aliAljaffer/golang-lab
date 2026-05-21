// logstats — aggregates log lines by level + throughput.
//
// This mini-project is deliberately designed so its test file exercises every
// pattern from examples 01-08 on a SINGLE piece of code:
//
//   01 basic-test     -> TestFormatRate_HappyPath
//   02 table-driven   -> TestParse table
//   03 subtests       -> TestFormatRate subtests via t.Run
//   04 mock-interface -> Source interface + recordingSource fake
//   05 httptest       -> TestHTTPSource_Fetch hits an httptest.Server
//   06 testdata       -> TestFileSource_Fetch reads testdata/lines.log
//   07 benchmark      -> BenchmarkParse_1k
//   08 fuzz           -> FuzzParse: round-trip + invariant ("no panic, accepted entries have a known level")
//
// Spec:
//   logstats [flags] <source>
//     <source>  either a file path or an http(s):// URL
//     --format  text (default) | json
//
// Run:
//   go run . ./testdata/lines.log
//   go run . --format json https://example.com/access.log
//
// Tests:
//   go test -tags=exercise ./06-testing/mini-project/...
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// ---------- Domain ----------

// Level is one of the recognized severity labels.
type Level string

const (
	LevelDebug Level = "DEBUG"
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

// KnownLevels is the closed set Parse accepts.
var KnownLevels = []Level{LevelDebug, LevelInfo, LevelWarn, LevelError}

// Entry is one parsed log line.
type Entry struct {
	Level Level
	Body  string
}

// Stats is the aggregated result.
type Stats struct {
	ByLevel  map[Level]int `json:"by_level"`
	Total    int           `json:"total"`
	Duration time.Duration `json:"duration"`
}

// ---------- Pure functions (table-tested, fuzz-tested, benchmarked) ----------

// Parse turns a raw log line into an Entry.
//
// Accepted format: "[LEVEL] body text" where LEVEL is one of KnownLevels.
// Whitespace around the brackets is tolerated. The body may be empty.
// Returns an error for unknown levels or malformed structure.
func Parse(line string) (Entry, error) {
	// TODO: trim leading whitespace
	// TODO: require a leading '[' and a matching ']'
	// TODO: extract the bracketed level; reject if not in KnownLevels
	// TODO: the body is everything after ']' with a single leading space stripped
	// TODO: return Entry{Level: lvl, Body: body}, nil

	_ = strings.TrimSpace
	return Entry{}, errors.New("Parse: not implemented")
}

// FormatRate renders an events-per-second figure with one decimal place.
// dur == 0 should return "0.0 events/s" (avoid div-by-zero).
func FormatRate(events int, dur time.Duration) string {
	// TODO: if dur <= 0, return "0.0 events/s"
	// TODO: compute rate := float64(events) / dur.Seconds()
	// TODO: return fmt.Sprintf("%.1f events/s", rate)

	return ""
}

// ---------- Stateful aggregator ----------

// Aggregator counts entries per level and tracks how long ingestion took.
// Zero value is ready to use.
type Aggregator struct {
	counts map[Level]int
	total  int
	start  time.Time
	end    time.Time
}

// Add records one entry. The first Add sets the start time; every Add updates
// the end time. Snapshot then reports `end - start` as Duration.
func (a *Aggregator) Add(e Entry) {
	// TODO: if a.counts == nil, initialize it
	// TODO: if a.start is zero, set a.start = time.Now()
	// TODO: a.counts[e.Level]++; a.total++; a.end = time.Now()
}

// Snapshot returns the current aggregated state. Safe to call multiple times.
func (a *Aggregator) Snapshot() Stats {
	// TODO: return Stats{
	// TODO:   ByLevel:  copyMap(a.counts),
	// TODO:   Total:    a.total,
	// TODO:   Duration: a.end.Sub(a.start),
	// TODO: }
	return Stats{}
}

// ---------- Source interface (mocked + httptest'd) ----------

// Source provides bytes to summarize. Production code uses FileSource or
// HTTPSource; tests inject a fake.
type Source interface {
	Fetch(ctx context.Context) (io.ReadCloser, error)
}

// FileSource reads from a local path.
type FileSource struct{ Path string }

func (f FileSource) Fetch(ctx context.Context) (io.ReadCloser, error) {
	// TODO: return os.Open(f.Path)
	_ = ctx
	return nil, errors.New("FileSource.Fetch: not implemented")
}

// HTTPSource performs a GET against URL using the injected HTTP client.
type HTTPSource struct {
	URL    string
	Client *http.Client
}

func (h HTTPSource) Fetch(ctx context.Context) (io.ReadCloser, error) {
	// TODO: req, err := http.NewRequestWithContext(ctx, "GET", h.URL, nil)
	// TODO: if err != nil { return nil, err }
	// TODO: resp, err := h.Client.Do(req)
	// TODO: if err != nil { return nil, err }
	// TODO: if resp.StatusCode != 200 { resp.Body.Close(); return nil, fmt.Errorf("status %d", resp.StatusCode) }
	// TODO: return resp.Body, nil
	_ = ctx
	return nil, errors.New("HTTPSource.Fetch: not implemented")
}

// ---------- Pipeline ----------

// Summarize pulls bytes from src, parses each line, and returns aggregated Stats.
// Lines that fail to Parse are skipped (counted separately would be a stretch).
func Summarize(ctx context.Context, src Source) (Stats, error) {
	// TODO: rc, err := src.Fetch(ctx); if err != nil { return Stats{}, err }
	// TODO: defer rc.Close()
	// TODO: var agg Aggregator
	// TODO: sc := bufio.NewScanner(rc)
	// TODO: for sc.Scan() {
	// TODO:     if e, err := Parse(sc.Text()); err == nil { agg.Add(e) }
	// TODO: }
	// TODO: return agg.Snapshot(), sc.Err()

	_ = bufio.NewScanner
	return Stats{}, errors.New("Summarize: not implemented")
}

// ---------- CLI ----------

func newRootCmd() *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "logstats <source>",
		Short: "Aggregate log lines by level + report throughput.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			src := pickSource(args[0])

			stats, err := Summarize(context.Background(), src)
			if err != nil {
				return err
			}

			switch format {
			case "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(stats)
			default:
				fmt.Printf("total: %d\n", stats.Total)
				for _, lvl := range KnownLevels {
					fmt.Printf("  %-5s %d\n", lvl, stats.ByLevel[lvl])
				}
				fmt.Printf("rate:  %s\n", FormatRate(stats.Total, stats.Duration))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&format, "format", "text", "output format: text | json")
	return cmd
}

func pickSource(arg string) Source {
	if strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "https://") {
		return HTTPSource{URL: arg, Client: &http.Client{Timeout: 5 * time.Second}}
	}
	return FileSource{Path: arg}
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
