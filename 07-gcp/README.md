# 07 — GCP SDK

> Status: ☑ scaffolded — see [`PLAN.md`](./PLAN.md)

If you're on GCP, sooner or later you write a tool that talks to GCP. The Go client libraries (`cloud.google.com/go/*`) are more verbose than the Python `google-cloud-*` equivalents — per-service modules, an explicit Iterator pattern, concrete return types you can't directly mock. The verbosity buys you compile-time errors instead of runtime `KeyError: 'name'` traces in CI.

This section is **self-contained**. If you did `07-aws` first, you'll find the foundational concepts (credentials chain, pagination, interface-based mocking) faster — they map cleanly across clouds, only the surface syntax differs. If you skipped `07-aws`, nothing here assumes it.

---

## What you'll learn

- `cloud.google.com/go/*` — the modern Go client libraries (per-service modules, each `go get`'d on its own)
- Application Default Credentials: env var → `gcloud` user creds → GCE/Cloud Run metadata server
- Project ID resolution — separate from credentials in GCP, unlike AWS where account ID is baked in
- GCS operations: list buckets + objects (`Iterator` pattern), `Writer`/`Reader` round-trip, V4 signed URLs
- GCE: `AggregatedList` across zones + label filter (`labels.env=prod`) + per-zone-map flatten
- IAM: short-lived credentials for a target SA via `impersonate.CredentialsTokenSource`
- **Interface-at-the-consumption-site mocking** with a thin wrapper (the GCP-specific wrinkle — see example 07)

---

## Mental model from other languages

| Concept                  | Go (`cloud.google.com/go/*`)                       | Python (`google-cloud-*`)                  | Go (AWS, for sibling-cloud reference)             |
| ------------------------ | -------------------------------------------------- | ------------------------------------------ | ------------------------------------------------- |
| Client construction      | `storage.NewClient(ctx)`                           | `storage.Client()`                         | `s3.NewFromConfig(cfg)`                           |
| Pagination               | iterator: `it.Next()` + `iterator.Done`            | `client.list_blobs()` (iter exhausted)     | `s3.NewListObjectsV2Paginator(...)`               |
| Credentials              | ADC chain via `google.FindDefaultCredentials`      | implicit (`google.auth.default()`)         | `config.LoadDefaultConfig(ctx)`                   |
| Cross-identity creds     | `impersonate.CredentialsTokenSource`               | `google.auth.impersonated_credentials`     | `stscreds.NewAssumeRoleProvider`                  |
| Server-side filter       | `labels.env=prod` DSL                              | same DSL string                            | `Filters: []ec2types.Filter{{Name, Values}}`      |
| Mock for tests           | wrap `*storage.Client` → custom interface          | `unittest.mock` / `pytest-mock`            | interface that matches `*s3.Client` 1:1           |

**The cultural difference vs AWS:** GCS client methods chain through concrete types (`client.Bucket(b).Object(k).NewReader(ctx)` — every step returns a `*storage.BucketHandle` / `*storage.ObjectHandle` / `*storage.Reader` you can't mock). So unlike AWS SDK v2 — where one interface satisfies `*s3.Client` directly — GCP wants you to write a **thin adapter** between the SDK and your business code. See example `07-mocking-gcs/` for the canonical shape.

---

## The DevOps angle

The boto-/Python-user's instinctive complaint about the Go GCP libraries is "why is the surface different per service?" The answer: each `cloud.google.com/go/<svc>` is a separate Go module on its own release cadence, and most are now generated from the same proto schemas as the gRPC backends. Compile-time errors when an API field renames is the payoff.

The non-obvious production details:

- **GCS CRC32C uses the Castagnoli polynomial, NOT the IEEE polynomial.** `crc32.MakeTable(crc32.Castagnoli)` is mandatory — the default `crc32.IEEE` table will compute a different hash than the server's, and your "did this file change?" tool will re-upload every file every run. Pinned in the mini-project test `TestCRCHelperUsesCastagnoli`.
- **Always use the iterator pattern (`it.Next()` + `iterator.Done`).** A list call that returns 1,000,001 objects spreads them across multiple pages; the iterator handles the page-token juggling. Skip the iterator and you process the first 1,000 and silently miss the rest.
- **`AggregatedList` returns empty zones.** Most projects use ~3 zones; `AggregatedList` returns ~40 entries (one per zone in your region), most empty. Always `if pair.Value == nil || len(pair.Value.Instances) == 0 { continue }` — otherwise your output is "us-west4-a (empty)" spammed 40 times.
- **Service account impersonation > JSON keys.** The Go SDK supports both, but mounting a JSON key on every container is a credential-management nightmare. `impersonate.CredentialsTokenSource` gives the same result with short-lived tokens; the only setup is a `roles/iam.serviceAccountTokenCreator` binding.
- **`Writer.Close()` is what commits an upload.** A bare `defer w.Close()` hides the upload error; the buffered bytes aren't on disk until `Close` succeeds. Call Close explicitly and check the error.

---

## Walkthrough

Read these in order. Each is a runnable example demonstrating one specific concept.

1. [`01-adc-and-project/`](./01-adc-and-project/) — `google.FindDefaultCredentials(ctx, scope)` walks the ADC chain; `creds.TokenSource.Token()` shows the actual token mint. Project ID resolution is documented as the separate, parallel chain.
2. [`02-gcs-list/`](./02-gcs-list/) — `client.Buckets(ctx, project)` + `bucket.Objects(ctx, &Query{})`. The first place you internalize the iterator-+-`iterator.Done` shape.
3. [`03-gcs-upload-download/`](./03-gcs-upload-download/) — `NewWriter` (upload) + `NewReader` (download). Round-trip a few KB. The asymmetric type shape and the `Writer.Close()` commit footgun are the headline.
4. [`04-gcs-signed-url/`](./04-gcs-signed-url/) — `bucket.SignedURL(name, &SignedURLOptions{Scheme: SigningSchemeV4, Method: "GET", Expires: time.Now().Add(5*time.Minute)})`. The 5-minute-window upload/download URL for a third party with no GCP creds.
5. [`05-gce-list/`](./05-gce-list/) — `compute.NewInstancesRESTClient(ctx)` + `AggregatedList(ctx, req)` + the `labels.env=prod` filter DSL + the per-zone-map flatten. GCP's analogue of AWS DescribeInstances + the Reservations wrapping quirk.
6. [`06-impersonate-sa/`](./06-impersonate-sa/) — `impersonate.CredentialsTokenSource(ctx, CredentialsConfig{TargetPrincipal: ...})` + `option.WithTokenSource(ts)` on the storage client. Short-lived creds for a target SA; the GCP analogue of AWS STS AssumeRole.
7. [`07-mocking-gcs/`](./07-mocking-gcs/) — **ships fully working.** Wraps `*storage.Client` behind a 3-method `GCSAPI` interface; a `fakeGCS` test verifies consumers without ever calling GCP. The pattern reused in the mini-project and exercises.

---

## Mini-project: [`gcssync`](./mini-project/)

Mirror a local directory → a GCS bucket. Like `gsutil rsync` but smaller: walks the local tree, computes Castagnoli CRC32 of each file, lists the bucket, plans uploads (new/changed) + optional deletes (`--delete`), runs with bounded concurrency, supports `--dry-run`.

The point: this exercises everything in the section — the iterator pattern (through the wrapper), `Writer`/`Reader` round-trip semantics, the load-bearing **CRC32C-with-Castagnoli** quirk, and the testing pattern from `07-mocking-gcs`. The 10 tests use a `fakeGCS` and never touch GCS; they pin concurrency-peak, dry-run-makes-no-calls, CRC32C-match-skips-upload, and a regression check that the helper actually uses Castagnoli (not the IEEE default).

Spec and starter in [`mini-project/`](./mini-project/).

---

## Exercises

See [`exercises/`](./exercises/):

1. **[`01-bucket-inventory`](./exercises/01-bucket-inventory/)** — list every object in every bucket of the project; emit CSV (bucket, name, size, updated). Practices the iterator-of-iterators + RFC3339 timestamps + project-scoped bucket listing.
2. **[`02-find-unlabeled`](./exercises/02-find-unlabeled/)** — list GCE instances missing a required label key. Practices `AggregatedList` (via the wrapper) + label-map semantics (`map[string]string` not a tag slice) + empty-value-counts-as-present.
3. **[`03-cleanup-old`](./exercises/03-cleanup-old/)** — delete GCS objects under a prefix whose `Updated` is older than a cutoff, with `--dry-run`. Practices listing + filtering + per-key delete + dry-run-makes-no-calls + partial-progress reporting.

---

## Further reading

- [Cloud client libraries for Go (overview)](https://cloud.google.com/go/docs/reference) — the per-service module index
- [`cloud.google.com/go/storage` godoc](https://pkg.go.dev/cloud.google.com/go/storage) — the GCS API; bookmark this
- [Application Default Credentials guide](https://cloud.google.com/docs/authentication/application-default-credentials) — the ADC chain spec, cross-language
- [`google.golang.org/api/iterator`](https://pkg.go.dev/google.golang.org/api/iterator) — the `iterator.Done` sentinel that every GCP list method uses
- [Service account impersonation](https://cloud.google.com/iam/docs/impersonating-service-accounts) — the IAM model behind `impersonate.CredentialsTokenSource`
- [V4 signed URLs](https://cloud.google.com/storage/docs/access-control/signed-urls) — the V4 algorithm + the two signing identity paths (JSON key vs. IAM signBlob)
- [The pinned testing doctrine: small interfaces at consumption site](./07-mocking-gcs/) — example 07 in this section (the GCP-flavoured version with the adapter wrinkle)
