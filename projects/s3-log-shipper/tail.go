package main

import (
	"context"
	"errors"
	"time"
)

// Line is one log line plus the byte offset *after* it in the source file.
// Persisting Offset is what lets a restart resume without re-shipping the
// world.
type Line struct {
	Path   string
	Body   []byte // does NOT include the trailing newline
	Offset int64  // offset AFTER this line (where to resume reading next)
}

// OffsetStore persists per-path file offsets across process restarts.
// The default FileOffsetStore impl writes a sibling "<path>.offset" file
// using the atomic-write pattern from section 02 (write tmp, fsync, rename).
type OffsetStore interface {
	// Load returns the last persisted offset for path. If no offset has ever
	// been saved (a "fresh" file), Load returns (0, nil) — not an error.
	Load(path string) (int64, error)
	// Save replaces the persisted offset atomically.
	Save(path string, off int64) error
}

// Tailer reads one file from a persisted offset, emits Lines on the channel
// passed to Run, and persists the new offset after each successful emit.
// One Tailer per file. Tailers are passed to projects/s3-log-shipper's Run
// which fans them in.
//
// Behavior contract (each pinned by a test in main_test.go):
//   - Cold start with no offset file: read from offset 0.
//   - Resume: read from OffsetStore.Load(Path).
//   - Truncation (current offset > file size): reset to 0 and continue.
//   - EOF then append: sleep PollInterval, then resume reading.
//   - Partial line at EOF (no trailing '\n'): buffer the bytes; emit only
//     once the newline arrives.
type Tailer struct {
	Path  string
	Store OffsetStore
	// PollInterval is how long Run sleeps after hitting EOF before checking
	// the file for new data. Zero is invalid (Run should pick a default or
	// the test sets a short value, e.g. 10ms).
	PollInterval time.Duration
}

// Run reads t.Path from its stored offset, emits one Line per newline-
// terminated record on out, and persists the post-emit offset via t.Store.
// Returns nil on ctx cancellation; returns an error on a non-retryable read
// error (e.g. permission denied; missing-file is NOT non-retryable — the
// file may appear later).
func (t *Tailer) Run(ctx context.Context, out chan<- Line) error {
	// TODO: load the starting offset via t.Store.Load(t.Path).
	// TODO: open the file (os.Open). If it doesn't exist yet, sleep PollInterval and retry until ctx is done.
	// TODO: stat the file; if start offset > size, reset to 0 (truncation).
	// TODO: seek to the start offset.
	// TODO: wrap the *os.File in a bufio.Reader; track current offset locally.
	// TODO: loop:
	//         - ReadBytes('\n'). If io.EOF and the buffer has no trailing '\n':
	//             - hold the partial bytes; sleep PollInterval; restat for truncation; continue.
	//         - on a full line, strip the '\n', advance currentOffset by len(line)+1,
	//           and select { case out <- Line{Path, body, currentOffset}: t.Store.Save(...) ; case <-ctx.Done(): return nil }.
	// TODO: respect ctx.Done() between reads.
	return errors.New("Tailer.Run not implemented")
}

// FileOffsetStore is the default OffsetStore. Writes "<path>.offset" beside
// each tailed file (or under Dir if Dir != ""), using the atomic-write
// pattern from section 02: write a tmp file, fsync, rename to the final
// name. The rename is what makes "did the offset get saved?" a yes/no, not
// a maybe.
type FileOffsetStore struct {
	// Dir, if non-empty, redirects all sidecar files into one directory.
	// The sidecar name is derived from the absolute path of the input file
	// (e.g. "/var/log/myapp.log" -> filepath.Join(Dir, url.PathEscape("/var/log/myapp.log") + ".offset")).
	Dir string
}

// Load reads the offset for path. Returns (0, nil) if no offset has been
// saved yet (the file genuinely doesn't exist).
func (s *FileOffsetStore) Load(path string) (int64, error) {
	// TODO: compute the sidecar path (sibling file or under s.Dir).
	// TODO: read its bytes; if os.IsNotExist(err), return (0, nil).
	// TODO: parse the bytes as a base-10 int64; return it.
	return 0, errors.New("FileOffsetStore.Load not implemented")
}

// Save writes off to the sidecar atomically: write tmp file, fsync, rename.
func (s *FileOffsetStore) Save(path string, off int64) error {
	// TODO: compute the sidecar path.
	// TODO: ensure the parent dir exists (os.MkdirAll if s.Dir was provided).
	// TODO: create a tmp file in the same dir; write strconv.FormatInt(off, 10); fsync; close; os.Rename to the final name.
	return errors.New("FileOffsetStore.Save not implemented")
}
