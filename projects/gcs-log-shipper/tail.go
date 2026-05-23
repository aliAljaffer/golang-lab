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
// One Tailer per file. Tailers are passed to projects/gcs-log-shipper's Run
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
	// TODO: implement the tail loop. The behaviour contract above is the spec.
	//   The interesting decisions:
	//     - cold-start vs resume: t.Store.Load decides which.
	//     - truncation detection: stat.Size < currentOffset means the file
	//       got rotated; the test pins "reset to 0 and keep going".
	//     - partial line at EOF: bufio.Reader.ReadBytes('\n') returns io.EOF
	//       *with* whatever bytes it read so far. Buffer those and wait for
	//       the newline rather than emitting a half line.
	//     - the offset persisted on Save is the offset AFTER the line — that's
	//       where the next read should resume from, including the '\n'.
	//     - missing file is retryable (the writer might not exist yet); a
	//       permission error is not.
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
	// TODO: load the offset for `path`. The "no offset yet" case must be
	//   silent: return (0, nil), not an error — the Tailer relies on that
	//   to cold-start. Other read or parse errors do propagate.
	return 0, errors.New("FileOffsetStore.Load not implemented")
}

// Save writes off to the sidecar atomically: write tmp file, fsync, rename.
func (s *FileOffsetStore) Save(path string, off int64) error {
	// TODO: persist `off` atomically. The pattern that survives a crash
	//   mid-write is: write to a temp file in the same directory, fsync,
	//   close, then os.Rename to the final name. Skipping the fsync is the
	//   classic "works in tests, loses data in prod" bug — the rename is
	//   only durable if the temp file's bytes are.
	return errors.New("FileOffsetStore.Save not implemented")
}
