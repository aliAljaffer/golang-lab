# 02 — Line-by-line with `bufio.Scanner`

When the file is bigger than you want to hold in memory, scan it.

## The pattern

```go
f, err := os.Open(path)
if err != nil { ... }
defer f.Close()

scanner := bufio.NewScanner(f)
for scanner.Scan() {
    line := scanner.Text()   // trailing newline already stripped
    // do something with line
}
if err := scanner.Err(); err != nil { ... }
```

## Gotchas

- **Default buffer is 64 KB per line.** Lines longer than that error out. Bump it with `scanner.Buffer(make([]byte, 1024*1024), 1024*1024)`.
- `scanner.Text()` reuses an internal buffer; if you keep the string around past the next `Scan()`, copy it (the `string()` conversion already does this).
- Check `scanner.Err()` after the loop — `Scan()` returning false hides the error.

## Why not `io.ReadAll`?

`io.ReadAll(f)` works, but you've now loaded the whole file into a `[]byte`. For a 1 GB log file that's a problem. Scanner streams.

## Comparison

| Language | Idiom |
|---|---|
| Go | `for scanner.Scan() { ... }` |
| Python | `for line in open(path):` |
| Bash | `while IFS= read -r line; do ... done < file` |
