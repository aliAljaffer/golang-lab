# Exercise 03 — `cleanup-old`

Delete every S3 object under a prefix whose `LastModified` is older than a
cutoff. Supports `--dry-run` (plan only).

## What to implement

```go
func Cleanup(ctx context.Context, api S3API, bucket, prefix string, cutoff time.Time, dryRun bool) ([]string, error)
```

Return the keys that were deleted (or would have been, in dry-run). Order
matches the paginator's order.

## How to test

```bash
go test -tags=exercise ./07-aws/exercises/03-cleanup-old/...
```

5 tests cover:

- Only objects older than cutoff are deleted (`.Before(cutoff)`)
- `dryRun=true` returns the plan but makes 0 DeleteObject calls
- The prefix is passed through to `ListObjectsV2` (server-side narrowing)
- A list error aborts without any delete
- A delete error propagates; partial-success list is not required

## Hints

- `obj.LastModified` is `*time.Time`. `obj.LastModified.Before(cutoff)` does
  the right thing (the deref is implicit on a method call).
- `*s3.DeleteObjectInput{Bucket: &bucket, Key: obj.Key}` — `obj.Key` is
  already `*string`, so just pass it through.
- The paginator is `s3.NewListObjectsV2Paginator`. The test's `fakeS3` returns
  one page total — you can verify pagination works on a real bucket later.
- For a CLI wrapper with a confirmation prompt, do it in a separate `main.go`
  package — not in this `cleanup` package. Keep the unit-tested core pure.
