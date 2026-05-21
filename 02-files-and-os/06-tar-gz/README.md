# 06 — `tar` + `gzip`

Building and reading `.tar.gz` archives. The stdlib has both pieces; you
compose them with `io.Pipe` / wrapping writers.

## The layering

```bash
.tar.gz file
   └── gzip.Writer        (compresses bytes)
         └── tar.Writer   (frames bytes as tar entries)
               └── per-file: WriteHeader(...) + Write(payload)
```

Reading is the mirror image:

```bash
.tar.gz file
   └── gzip.Reader
         └── tar.Reader
               └── per-entry: Next() returns *tar.Header; Read() gives the payload
```

## The pattern (writing)

```go
out, _ := os.Create("archive.tar.gz")
defer out.Close()

gz := gzip.NewWriter(out)
defer gz.Close()

tw := tar.NewWriter(gz)
defer tw.Close()

for each file:
    hdr := &tar.Header{Name: "...", Mode: 0o644, Size: int64(len(data))}
    tw.WriteHeader(hdr)
    tw.Write(data)
```

The `defer` order matters. They run in **reverse**: tar closer flushes tar
records, gzip closer flushes compression, then the file closes. Get this
wrong and you'll have truncated archives.

## Things to learn

- `tar.Header.Size` must match the bytes you write. If you write fewer, `Close()` errors. If you write more, the next entry corrupts.
- For directories, set `Typeflag: tar.TypeDir` and `Size: 0`.
- The streaming model means you can pipe a tarball over HTTP without writing to disk — see section 04 for the equivalent shape there.
