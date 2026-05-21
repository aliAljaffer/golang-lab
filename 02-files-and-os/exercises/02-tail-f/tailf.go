// 02-tail-f — the testable kernel of `tail -f`.
//
// You implement: ReadAppend.
// The tests in tailf_test.go drive the design.
package tailf

import "os"

// ReadAppend reads everything in `f` from `lastSize` to EOF and returns those
// new bytes plus the file's new size.
//
// Contract:
//   - Caller owns `f`; do not Close or Open here.
//   - If the file grew, return the appended bytes.
//   - If the file didn't grow, return (nil, lastSize, nil).
//   - If the file shrank (truncation / rotation), return a non-nil error.
func ReadAppend(f *os.File, lastSize int64) ([]byte, int64, error) {
	// TODO: stat the file to learn its current size.
	// TODO: if size == lastSize, return (nil, lastSize, nil).
	// TODO: if size < lastSize, return an error (rotation/truncate).
	// TODO: f.Seek(lastSize, io.SeekStart)
	// TODO: buf := make([]byte, size - lastSize); io.ReadFull(f, buf)
	// TODO: return buf, size, nil
	return nil, lastSize, nil
}
