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
	// TODO: compare current size against lastSize; three branches, each
	//   pinned by a test:
	//     - equal: nothing happened, return (nil, lastSize, nil).
	//     - smaller: file shrank (rotation / truncate), return an error so
	//       the caller can re-open from the start.
	//     - larger: read exactly the appended bytes (size-lastSize), starting
	//       at lastSize. io.ReadFull is the right primitive — io.Copy would
	//       try to read to EOF and you've already learned where that is.
	return nil, lastSize, nil
}
