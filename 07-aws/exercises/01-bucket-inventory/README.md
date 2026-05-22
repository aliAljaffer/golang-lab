# Exercise 01 — `bucket-inventory`

Walk every bucket in the account, list every object, and emit a CSV row per
object: `bucket,key,size,last_modified`.

## What to implement

```go
func Inventory(ctx context.Context, api S3API) ([]Row, error)
func WriteCSV(w io.Writer, rows []Row) error
```

`api` is a small interface listing only what you need: `ListBuckets` +
`ListObjectsV2`. `*s3.Client` satisfies it for free — tests pass a `fakeS3`.

## How to test

```bash
go test -tags=exercise ./07-aws/exercises/01-bucket-inventory/...
```

5 tests will fail until both functions are implemented:

- Empty account returns no rows
- Two buckets, two objects → 2 rows, fields populated
- Bucket order matches `ListBuckets` order (no implicit sorting)
- An error from any single bucket aborts the whole call
- `WriteCSV` emits a header + RFC3339 timestamps via `encoding/csv`

## Hints

- `s3.NewListObjectsV2Paginator(api, ...)` — you don't write the loop manually.
- `*obj.Size` and `*obj.LastModified` need nil checks if you want to be
  defensive. The tests always populate them, so for the exercise you can
  deref directly.
- `encoding/csv.NewWriter` requires a `Flush()` at the end — or `Error()` to
  check whether the underlying writer ate it.
