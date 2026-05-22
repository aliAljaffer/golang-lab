// Package buildtar builds an in-memory tar containing a Dockerfile and any
// referenced context files, ready to hand to `cli.ImageBuild`.
//
// Exercise surface:
//
//	type File struct { Name string; Body []byte; Mode int64 }
//	func BuildContext(files []File) ([]byte, error)
//
// Why this matters: `ImageBuild` takes an `io.Reader` of a tar containing the
// Dockerfile + any files COPY'd into it. The CLI builds this tar by walking
// the build context directory. When you're building programmatically — e.g.
// generating a Dockerfile on the fly and shipping it WITHOUT touching disk —
// you build the tar yourself in memory and pass `bytes.NewReader(tar)` as the
// build context. That's what this exercise teaches.
//
// What's left out (you can layer on top):
//   - Calling `cli.ImageBuild` with the produced tar. The interesting part is
//     the tar construction; once you have valid bytes, the SDK call is a
//     one-liner.
//   - tar.gz: ImageBuild also accepts gzipped tars. Build the tar first;
//     gzip it with `compress/gzip` if you want.
package buildtar

import (
	"errors"
)

// File is one entry in the build context.
type File struct {
	Name string // path inside the tar, e.g. "Dockerfile", "app/main.go"
	Body []byte // file contents
	Mode int64  // unix file mode; 0644 for normal files, 0755 for scripts
}

// BuildContext returns a tar archive containing the given files in the order
// supplied. The output is a "naked" tar (NOT gzipped) — ImageBuild accepts both.
//
// Contract pinned by tests:
//   - Output is a valid tar (round-trips through tar.NewReader).
//   - Every file is preserved exactly (Name, body bytes, Mode).
//   - Order is preserved.
//   - Empty input yields a valid empty tar (no entries; no error).
//   - Each entry uses tar.TypeReg (regular file) and a sensible ModTime.
//
// Hints:
//   - Build into a bytes.Buffer.
//   - For each file: tw.WriteHeader(&tar.Header{Name, Size: int64(len(Body)), Mode, Typeflag: tar.TypeReg, ModTime: now}), then tw.Write(Body).
//   - Don't forget tw.Close() — without it, the tar terminator block is missing
//     and tar.NewReader will error mid-stream.
func BuildContext(files []File) ([]byte, error) {
	// TODO: implement
	_ = errors.New
	return nil, errors.New("BuildContext not implemented")
}
