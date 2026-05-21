// 07-atomic-write — write a file safely using temp-file + rename.
//
// Goal: implement atomicWrite(path, data, perm) error, then call it from main
// to write a small JSON-ish blob. The function should:
//   1. CreateTemp in the *same directory* as path.
//   2. Write the data.
//   3. Sync + Close.
//   4. Chmod to perm.
//   5. Rename onto the target.
//   6. On any failure, remove the temp file.
//
// Run:
//   go run .
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func atomicWrite(path string, data []byte, perm fs.FileMode) error {
	// TODO: dir := filepath.Dir(path)
	// TODO: tmp, err := os.CreateTemp(dir, ".tmp-*")
	// TODO: rollback := func() { os.Remove(tmp.Name()) } — call on every error path before return.
	// TODO: _, err = tmp.Write(data)        ; on err -> rollback
	// TODO: tmp.Sync()                       ; on err -> rollback
	// TODO: tmp.Close()                      ; on err -> rollback
	// TODO: os.Chmod(tmp.Name(), perm)       ; on err -> rollback
	// TODO: os.Rename(tmp.Name(), path)      ; on err -> rollback
	// TODO: return nil
	return fmt.Errorf("atomicWrite: not implemented")
}

func main() {
	const path = "/tmp/go-learning-07-atomic.json"
	defer os.Remove(path)

	if err := atomicWrite(path, []byte(`{"ok":true}`+"\n"), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "atomicWrite:", err)
		os.Exit(1)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read back:", err)
		os.Exit(1)
	}
	fmt.Printf("wrote %s atomically: %s", path, got)

	_ = filepath.Dir
}
