# 03 — GCS upload / download (Writer + Reader)

GCS exposes the object as a streaming Writer (for upload) and ReadCloser
(for download). No `Body io.Reader` field on an input struct like S3 — the
client API is shape-asymmetric: `NewWriter`/`NewReader` rather than
`Put`/`Get`.

## Upload: Close() commits

```go
w := obj.NewWriter(ctx)
io.Copy(w, src)
if err := w.Close(); err != nil { /* real error */ }
```

The bytes are buffered/chunked while you write. **`w.Close()` is what actually
finalizes the upload.** A bare `defer w.Close()` swallows the upload error —
call Close explicitly and inspect the error. The `defer` is only useful for
the early-return case (cleanup after a partial write).

## Download: NewReader streams

```go
r, _ := obj.NewReader(ctx)
defer r.Close()
io.Copy(dst, r)
```

`NewReader` performs the request lazily — the HTTP call only fires on the
first `Read`. Easy to miss: a `NewReader` that "succeeds" doesn't mean the
object exists. You'll learn that the first time the first `Read` returns a
404-shaped error.

`Bucket(name).Object(key).NewRangeReader(ctx, offset, length)` is the same
shape with byte-range support — handy for resumable downloads.

## Compare to AWS / Python

| Concept              | Go (GCS)                             | Go (AWS S3)                                       | Python (`google-cloud-storage`)        |
| -------------------- | ------------------------------------ | ------------------------------------------------- | -------------------------------------- |
| Upload               | `obj.NewWriter(ctx)` + Close commits | `client.PutObject(ctx, &PutObjectInput{Body: r})` | `blob.upload_from_filename(path)`      |
| Download             | `obj.NewReader(ctx)` ReadCloser      | `client.GetObject(ctx, ...).Body`                 | `blob.download_to_filename(path)`      |
| Resumable for >5 MB  | automatic (Writer.ChunkSize default) | use `feature/s3/manager.Uploader`                 | automatic                              |
| Where errors surface | first Read / Close                   | the call itself                                   | the function                           |

## Gotchas

- **Forgetting `Close()` silently drops uploads.** Writer buffers in memory
  until then. If your program exits before Close, you uploaded nothing.
- **The Reader doesn't fail until first Read.** Don't gate on
  `_, err := obj.NewReader(ctx); err != nil` and expect "not found" to fire.
  Wrap a `storage.ErrObjectNotExist` check around the first `Read` (or
  call `obj.Attrs(ctx)` first if you want a cheap existence check).
- **Reader has `Attrs()` after first Read** — `r.Attrs.Size`, `r.Attrs.ContentType`,
  etc. Inspect after Read, not before.

## TODO

1. Uncomment PART 1, run `go run . <bucket> hello.txt /etc/hosts`. Verify with
   `gsutil ls gs://<bucket>` or example 02.
2. Uncomment PART 2, confirm the round-trip works.
3. Comment out `w.Close()` (PART 1) and run again — confirm the object is NOT
   uploaded. This is the most common GCS-Go footgun in production.
4. Make the upload set `w.ContentType = "text/plain"` and `w.Metadata = map[string]string{"src": "07-gcp-example-03"}` before `io.Copy`. Confirm with `obj.Attrs(ctx)`.
