# 03 — `filepath.WalkDir`

The recursive directory traversal in Go's stdlib. Prefer it over the older `filepath.Walk` — `WalkDir` doesn't `os.Stat` every entry, so it's much faster on big trees.

## The shape

```go
err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
    if err != nil { return err }       // propagate
    if d.IsDir() { return nil }        // skip dirs, only act on files
    // do something with `path`
    return nil
})
```

The walk function can return three magic values:

| Return | Meaning |
|---|---|
| `nil` | Keep walking |
| `filepath.SkipDir` | Don't descend into this directory |
| any other error | Stop the walk, return that error from `WalkDir` |

## Why `path/filepath`, not `path`?

`path` uses `/` always — it's for **URL-like** paths.
`path/filepath` uses the OS separator (`\` on Windows). Always use `filepath` for actual filesystem paths.

## Comparison

| Language | Idiom |
|---|---|
| Go | `filepath.WalkDir(root, fn)` |
| Python | `os.walk(root)` |
| Bash | `find root -type f` |
