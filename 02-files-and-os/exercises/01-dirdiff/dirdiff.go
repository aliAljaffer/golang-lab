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
	// TODO: hashTree(root) -> map[relPath]sha256sum, error
	// TODO: call it for left and right
	// TODO: walk the union of keys:
	//         - in left only             -> OnlyLeft
	//         - in right only            -> OnlyRight
	//         - in both, hashes differ   -> Modified
	//         - in both, hashes equal    -> skip
	// TODO: return the slice.
	return nil, nil
}
