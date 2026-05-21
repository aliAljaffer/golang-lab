# 01 — Read & Write

The four functions you'll reach for 90% of the time.

| Function | Use when |
|---|---|
| `os.ReadFile(path)` | File fits in memory, you want the whole thing. |
| `os.WriteFile(path, data, perm)` | Same, in reverse. Truncates existing file. |
| `os.Open(path)` | Streaming reads — large files, line-by-line, etc. Returns `*os.File` (an `io.Reader`). |
| `os.Create(path)` | Streaming writes. Returns `*os.File` (an `io.Writer`). Truncates existing file. |

## Things to notice

- `os.ReadFile` / `os.WriteFile` are shortcuts for the four-step `Open → Read/Write → Close → handle errors` dance. Use them whenever the file is small.
- `*os.File` is both an `io.Reader` and an `io.Writer`. That's why you can pass it to `io.Copy`, `bufio.NewScanner`, `json.NewEncoder`, etc.
- **Always `defer f.Close()`** after a successful `Open`/`Create`. Forgetting is a leaked file descriptor.
- File permissions are octal. `0o644` = owner rw, group r, other r. `0o600` for secrets.

## Comparison

| Concept | Go | Python | Bash |
|---|---|---|---|
| Read whole file | `os.ReadFile` | `pathlib.read_text()` | `$(<file)` |
| Write whole file | `os.WriteFile` | `pathlib.write_text()` | `> file` |
| Streaming read | `os.Open` + `bufio.Scanner` | `with open(...) as f:` | `while read line` |
