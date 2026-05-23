# 02 — GCS list (buckets + objects via Iterator)

## Service clients

GCS lives in `cloud.google.com/go/storage`. Construct one client per process:

```go
client, _ := storage.NewClient(ctx)
defer client.Close()
```

The client picks up ADC automatically — no project ID here, no region. Bucket
names are a global namespace.

## The Iterator pattern

Every list call in every GCP Go client returns an iterator type:

```go
it := client.Buckets(ctx, projectID)
for {
    attrs, err := it.Next()
    if errors.Is(err, iterator.Done) { break }
    if err != nil { /* real failure */ }
    // use attrs
}
```

`iterator.Done` is the sentinel. Anything else returned from `Next()` is a
real error. The iterator hides paging — it fetches a page, hands you items
one at a time, fetches the next page when you exhaust the current one.

**Why a sentinel, not a `HasMore()`?** GCP-Go borrowed the idea from
[`google.golang.org/api/iterator`](https://pkg.go.dev/google.golang.org/api/iterator)
(it's literally the same `iterator.Done` constant across all `google.golang.org/api/*`
clients), and a sentinel composes better with `errors.Is` and ctx-cancel
handling. AWS SDK v2's `HasMorePages()`/`NextPage(ctx)` is the same idea with
a different surface.

## Compare to AWS / Python

| Concept                  | Go (GCS)                              | Go (AWS S3)                            | Python (`google-cloud-storage`) |
| ------------------------ | ------------------------------------- | -------------------------------------- | ------------------------------- |
| List buckets             | `client.Buckets(ctx, project)`        | `client.ListBuckets(ctx, in)`          | `client.list_buckets()`         |
| Page through objects     | `bucket.Objects(ctx, &Query{})`       | `s3.NewListObjectsV2Paginator(...)`    | `bucket.list_blobs()`           |
| Termination              | `errors.Is(err, iterator.Done)`       | `for p.HasMorePages()`                 | iter exhausted naturally        |
| Filter at source         | `&Query{Prefix: "logs/"}`             | `Input.Prefix`                         | `prefix=` kwarg                 |

## TODO

1. Uncomment PART 1, run `go run . <your-project-id>`. Confirm you see your
   buckets in roughly-creation order.
2. Uncomment PART 2, run `go run . <project> <one-of-your-buckets>`. Note that
   `attrs.Updated` (not `attrs.Created`) is the field you'd compare for staleness.
3. Add `Prefix: "logs/"` to the `Query{}` — confirm the iterator narrows
   server-side. (Server-side filter is free; client-side filter pays for the
   bytes you'll throw away.)
