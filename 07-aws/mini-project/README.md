# Mini-project — `s3sync`

Mirror a local directory to an S3 bucket with ETag-based dedup, optional
`--delete`, and bounded concurrency.

## What this exercises

| From PLAN.md | Where |
|---|---|
| `config.LoadDefaultConfig` | `newRootCmd` |
| Service client construction | `s3.NewFromConfig(cfg)` |
| Pagination | `ListRemote` |
| ETag = md5 for single-PUT | `Plan` (dedup) |
| Interface-based mocking | `S3API` + `fakeS3` in tests |
| Concurrency cap | `Sync` |

## Surface

```go
type S3API interface {
    ListObjectsV2(ctx, *s3.ListObjectsV2Input, ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
    PutObject(ctx, *s3.PutObjectInput, ...func(*s3.Options)) (*s3.PutObjectOutput, error)
    DeleteObject(ctx, *s3.DeleteObjectInput, ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

func WalkLocal(root string) ([]LocalFile, error)
func ListRemote(ctx context.Context, api S3API, bucket string) (map[string]string, error)
func Plan(locals []LocalFile, remotes map[string]string, opts Options) []Action
func Sync(ctx context.Context, api S3API, opts Options) (uploaded, deleted, skipped int, err error)
```

`*s3.Client` satisfies `S3API` automatically — production passes it; tests
pass `fakeS3`.

## Run

```bash
go run . --bucket my-b --dir ./pages --concurrency 4
go run . --bucket my-b --dir ./pages --dry-run      # plan only
go run . --bucket my-b --dir ./pages --delete       # remove orphaned keys
```

## Tests

```bash
go test -tags=exercise ./07-aws/mini-project/...   # 7 failing tests until you implement
```

The tests use `fakeS3` to verify behavior without any AWS network call:

- `TestWalkLocal_*` — local file scan, forward-slash keys
- `TestListRemote_*` — paginator + ETag unquoting
- `TestPlan_*` — diff rules with and without `--delete`
- `TestSync_HappyPath_*` — round-trip of upload + skip
- `TestSync_DryRunDoesNotCallPutOrDelete` — guards `--dry-run`
- `TestSync_RespectsConcurrencyLimit` — uses `atomic.Int32` + CAS to record peak
  in-flight from inside `fakeS3.PutObject`. Blocks puts on a hold channel
  until the test confirms the worker pool fanned out to N
- `TestSync_DeleteRemovesStaleKeys` — `--delete` removes orphans
- `TestSync_PropagatesError` — first API error short-circuits

## Hints

- For ETag comparison, single-PUT objects have ETag = `"md5hex"` (with literal
  quote marks). Strip the quotes when reading; don't add them when computing.
- Files larger than 5 MB get multipart uploads whose ETag is **not** their md5
  — out of scope for this project, but worth knowing before you ship to prod.
- `Sync`'s worker pool can be either a semaphore channel (buffered chan of
  size N) or a fixed pool reading from a job channel. Either is fine; the
  test only checks the in-flight peak, not the architecture.
