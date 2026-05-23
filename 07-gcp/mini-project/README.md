# Mini-project — `gcssync`

Mirror a local directory to a GCS bucket with CRC32C-based dedup, optional
`--delete`, and bounded concurrency. The GCP-flavoured sibling of
`07-aws/mini-project/s3sync`.

## What this exercises

| From PLAN.md                       | Where                                    |
| ---------------------------------- | ---------------------------------------- |
| ADC + storage client construction  | `newRootCmd` (production wiring TODO)    |
| Iterator pattern (via wrapper)     | `ListRemote` calling `api.List`          |
| CRC32C (Castagnoli) hashing        | `WalkLocal` + `castagnoli` table         |
| Interface-based mocking (3 methods) | `GCSAPI` + `fakeGCS` in tests           |
| Concurrency cap                    | `Sync`                                   |

## Surface

```go
type GCSAPI interface {
    List(ctx, bucket, prefix string) ([]RemoteObject, error)
    Upload(ctx, bucket, key string, body io.Reader) error
    Delete(ctx, bucket, key string) error
}

func WalkLocal(root string) ([]LocalFile, error)
func ListRemote(ctx context.Context, api GCSAPI, bucket string) (map[string]RemoteObject, error)
func Plan(locals []LocalFile, remotes map[string]RemoteObject, opts Options) []Action
func Sync(ctx context.Context, api GCSAPI, opts Options) (uploaded, deleted, skipped int, err error)
```

The 3-method `GCSAPI` is the same shape as `07-mocking-gcs/gcsutil.GCSAPI`.
Production wires it to a real `*storage.Client` adapter; tests pass `fakeGCS`.

## Run

```bash
go run . --bucket my-b --dir ./pages --concurrency 4
go run . --bucket my-b --dir ./pages --dry-run      # plan only
go run . --bucket my-b --dir ./pages --delete       # remove orphans
```

## Tests

```bash
go test -tags=exercise ./07-gcp/mini-project/...   # failing until you implement
```

The tests use `fakeGCS` to verify behaviour without any GCS network call:

- `TestWalkLocal_*` — local file scan, forward-slash keys, Castagnoli CRC
- `TestListRemote_*` — drain into keyed map
- `TestPlan_*` — diff rules with and without `--delete`; deterministic order
- `TestSync_HappyPath_*` — round-trip of upload + skip
- `TestSync_DryRunMakesNoMutatingCalls` — guards `--dry-run`
- `TestSync_RespectsConcurrencyLimit` — uses `atomic.Int32` + CAS to record peak
  in-flight from inside `fakeGCS.Upload`. Blocks uploads on a hold channel
  until the test confirms the worker pool fanned out to N.
- `TestSync_DeleteRemovesStaleKeys` — `--delete` removes orphans
- `TestSync_PropagatesError` — first API error short-circuits
- `TestSync_CRC32CMatchSkipsUpload` — the load-bearing GCS test:
  same-CRC → skip, not re-upload
- `TestCRCHelperUsesCastagnoli` — guards against accidentally using the
  IEEE polynomial (the `hash/crc32` default), which would mismatch GCS

## Hints

- **Use `crc32.MakeTable(crc32.Castagnoli)`, not the default IEEE table.** GCS
  computes the Castagnoli polynomial server-side. If you use IEEE, every
  file looks "different" forever — you re-upload the world every run.
- `attrs.CRC32C` from GCS is `uint32`. Match types: don't compare against an
  `int64` or a hex string.
- `Sync`'s worker pool can be either a semaphore channel (buffered chan of
  size N) or a fixed pool reading from a job channel. Either is fine; the
  test only checks the in-flight peak, not the architecture.
- The `cmd.RunE` body has a TODO marker for wiring the real GCSAPI — you'll
  use `gcsutil.NewReal(ctx)` (see `../07-mocking-gcs/`) and pass it as
  `api`. Don't add that wiring until your tests pass with `fakeGCS`.
