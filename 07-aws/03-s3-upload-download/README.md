# 03 — S3 upload + download

Two operations, one round-trip: read a local file → `PutObject`, then
`GetObject` → write to disk.

## The streaming shape

`PutObject` takes `Body io.Reader`. Anything that implements Read works:
`*os.File`, `bytes.Buffer`, `strings.Reader`. The SDK will stream from it
and won't buffer the whole thing in memory.

`GetObject` returns `*GetObjectOutput` whose `Body` is `io.ReadCloser`.
**You must Close it** even if you read to EOF — otherwise the connection
won't return to the pool. `defer out.Body.Close()`.

## Small vs large files

| Size | Use |
|---|---|
| < 5 MB | Plain `PutObject` (this example) |
| ≥ 5 MB or unknown | `feature/s3/manager.Uploader` — concurrent multipart, automatic |

The mini-project `s3sync` uses plain PutObject for simplicity. In production,
prefer `manager.Uploader` for unbounded inputs (it falls back to a single PUT
for small files anyway).

## ETag = MD5 (sometimes)

For files uploaded with a single PUT (no multipart), the object's ETag is
the hex MD5 of the body. This is the trick `s3sync` uses to skip unchanged
files. **For multipart uploads, ETag is not the MD5** — it's `<md5-of-md5s>-N`.

## TODO

1. Fill in PART 1 + PART 2.
2. Round-trip a small file: `go run . my-bucket test.txt ./README.md`
3. Compare `md5 ./README.md` to the ETag — should match (single-PUT path).
