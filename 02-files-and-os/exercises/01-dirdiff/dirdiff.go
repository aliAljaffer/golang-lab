// 01-dirdiff — compare two directory trees by content hash.
//
// You implement: Diff(left, right string) ([]Entry, error).
// The tests in dirdiff_test.go drive the design.
package dirdiff

// Kind classifies how a relative path appears across the two trees.
type Kind int

const (
	OnlyLeft Kind = iota
	OnlyRight
	Modified
)

// Entry is one difference.
type Entry struct {
	Path string // relative to the input roots
	Kind Kind
}

// Diff walks `left` and `right` and returns one Entry per file that differs.
// Files present in both with identical sha256 contents are omitted.
// Only regular files are considered; directories are ignored.
func Diff(left, right string) ([]Entry, error) {
	// TODO: build {relPath -> sha256} for each tree, then walk the union of
	//   keys and classify each into OnlyLeft / OnlyRight / Modified. Files
	//   present on both sides with identical hashes contribute no Entry.
	//   You'll likely want a helper that does the walk + hash; the test
	//   doesn't care about its shape, only the Entry list it produces.
	return nil, nil
}
