# 01 — bucket inventory

Walk every bucket in a project and emit one CSV row per object.

## Surface

```go
type GCSAPI interface {
    ListBuckets(ctx, project string) ([]BucketAttrs, error)
    ListObjects(ctx, bucket string) ([]ObjectAttrs, error)
}

func Inventory(ctx, api, project string) ([]Row, error)
func WriteCSV(w io.Writer, rows []Row) error
```

The `GCSAPI` interface lives in *this* package (the
07-mocking-gcs doctrine) so the test fake doesn't have to construct real
`*storage.BucketHandle` / `*storage.ObjectIterator`.

## Tests pin

- empty project → no rows, no error
- order: rows follow ListBuckets order, then ListObjects order within bucket
- one bucket erroring fails the whole call (fail-fast contract)
- WriteCSV emits header `bucket,name,size,updated` and RFC3339 timestamps

## What you're practicing

- **The iterator-of-iterators pattern.** Bucket list, then per-bucket object list.
- **`encoding/csv`** — header + per-row writer. (Don't `Sprintf` strings into
  CSV; quoting + escape rules will bite you.)
- **RFC3339 timestamps** — what every CSV consumer expects, what spreadsheets
  parse, what `gsutil` itself emits.
