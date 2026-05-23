# 03 — cleanup-old

Delete GCS objects under a prefix whose `Updated` timestamp is older than a
cutoff. Standard janitor — every team's bucket has a `tmp/` or `staging/`
prefix nobody wants to hand-clean.

## Surface

```go
type GCSAPI interface {
    ListObjects(ctx, bucket, prefix string) ([]ObjectAttrs, error)
    DeleteObject(ctx, bucket, name string) error
}

func Cleanup(ctx, api, bucket, prefix string, cutoff time.Time, dryRun bool) ([]string, error)
```

## Tests pin

- only old objects get deleted; the rest are left alone
- `--dry-run` mode lists what would go but makes zero Delete calls
- prefix passes through to the listing (server-side narrowing)
- list errors abort with no work done
- delete errors return the error (with partial progress in the returned slice)
- empty bucket is a validation error

## What you're practicing

- **List → filter → per-key delete loop.** The bread-and-butter shape of
  every GCS janitor tool.
- **`time.Time.Before(cutoff)`** — the cutoff comparison. `Updated` is
  `time.Time` directly (not a pointer like the AWS SDK's `*time.Time`) so
  the comparison is one expression with no nil checks.
- **Dry-run contract.** The most important contract for a destructive tool:
  the dry-run path must NOT call DeleteObject. The test fakes this with
  `len(f.deletes) == 0`.
- **Partial-progress reporting.** When delete fails halfway, the caller
  should see what was deleted so far. Don't return `nil, err` — return the
  partial list AND the error.
