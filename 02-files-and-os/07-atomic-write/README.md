# 07 — Atomic file writes

The pattern that survives crashes, power loss, and concurrent readers.

## The problem

```go
os.WriteFile("config.yaml", data, 0o644)
```

If the process dies between `truncate` and the final `write`, you've replaced
`config.yaml` with **half** a config. Anyone reading at the wrong moment sees
garbage.

## The fix: temp-file-plus-rename

```go
dir := filepath.Dir(target)
tmp, err := os.CreateTemp(dir, ".tmp-*")     // same filesystem as target!
// write everything to tmp
tmp.Sync()                                    // flush kernel buffers
tmp.Close()
os.Rename(tmp.Name(), target)                 // atomic on POSIX
```

POSIX guarantees `rename(2)` is atomic *on the same filesystem*. Readers
either see the old file or the new file, never a half-written one.

## Things to learn

- The temp file **must be on the same filesystem** as the target — otherwise `os.Rename` falls back to copy+delete, which is no longer atomic. Putting it in the same directory is the easiest way to guarantee this.
- `Sync()` matters: rename is atomic at the filesystem-metadata level, but the data blocks might not be on disk yet. Without `Sync`, a power loss can leave you with the new filename pointing at zeros. (For most apps the metadata atomicity alone is enough — but for `etcd`-style durability, sync.)
- Clean up the temp file if the write fails — `defer os.Remove(tmp.Name())` after a failure flag, or use a `func() { ... }()` wrapper.
- Permissions: `CreateTemp` makes a `0o600` file. If you need `0o644` after rename, `os.Chmod` it before the rename.

## Why this matters in DevOps

Anything that writes a config (`/etc/...`), a state file (`/var/lib/...`), or
a credentials file should use this pattern. Half-written `kubeconfig` is a
support ticket; half-written `/etc/passwd` is an outage.
