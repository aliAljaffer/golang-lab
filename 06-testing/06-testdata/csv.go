// 06-testdata — fixture files live in `testdata/`. The Go toolchain treats it
// as a magic directory: ignored by the build (so non-.go files don't error out)
// but committed to VCS like anything else.
package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
)

// Row is one parsed line.
type Row struct {
	Name  string
	Score int
}

// Parse reads CSV with header "name,score" and returns a slice of Row.
// Returns an error for malformed numbers or missing columns.
func Parse(r io.Reader) ([]Row, error) {
	cr := csv.NewReader(r)
	records, err := cr.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("empty input")
	}
	if len(records[0]) != 2 || records[0][0] != "name" || records[0][1] != "score" {
		return nil, fmt.Errorf("header = %v, want [name score]", records[0])
	}

	out := make([]Row, 0, len(records)-1)
	for i, rec := range records[1:] {
		score, err := strconv.Atoi(rec[1])
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid score %q: %w", i+2, rec[1], err)
		}
		out = append(out, Row{Name: rec[0], Score: score})
	}
	return out, nil
}
