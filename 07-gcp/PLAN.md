# Plan: 07-gcp

> Self-contained. A student who skips 07-aws can do this section without losing context. If you *have* done 07-aws, you'll find the foundational concepts (credentials chain, pagination, interface-based mocking) faster — they map cleanly across both clouds, only the surface syntax differs.

## Prerequisites

- GCP project with billing enabled (the always-free tier covers a GCS standard bucket + an `e2-micro` VM, which is enough for everything here)
- `gcloud` installed
- One of: `gcloud auth application-default login` (interactive) **or** `GOOGLE_APPLICATION_CREDENTIALS` pointing at a service-account JSON (CI-style)
- A second service account in the same project for the impersonation example (06)
- Integration tests that hit real GCP are tagged `//go:build integration` to skip in CI

## Concepts to cover

Foundational (taught from scratch — skim if you've done 07-aws):

- [ ] `cloud.google.com/go/*` module layout — per-service Go modules (`storage`, `compute`, `iam/credentials/apiv1`), each `go get`'d individually. Why this is more verbose than the equivalent Python `google-cloud-storage` import but buys you compile-time errors when an API field renames.
- [ ] The credentials chain as a concept — every cloud SDK has one. What it means for a tool to "just work" locally and in a VM/container without code changes.
- [ ] Pagination as a concept — why every list API in every cloud SDK returns pages, and what happens if you ignore the next-page token (you silently process the first N and miss the rest).
- [ ] **Interface-at-consumption-site mocking** — the doctrine that your function should take a small interface containing only the SDK methods it actually calls, not the full SDK client. This makes tests trivial and lives in *your* code, not the SDK's. The single most useful pattern in the section.

GCP-specific:

- [ ] Application Default Credentials (ADC) — the GCP credentials chain: env var → `gcloud` user creds → GCE/Cloud Run metadata server. Entry points: `google.FindDefaultCredentials` (lower-level) or just passing nothing to the service client constructor (higher-level — ADC fires implicitly).
- [ ] Project ID resolution — GCP has no equivalent of an account ID baked into creds; project ID is a separate dimension, picked up from `GOOGLE_CLOUD_PROJECT` or `gcloud config get-value project`.
- [ ] The `Iterator` pattern — `it.Next()` returning `(T, error)` with `iterator.Done` as the terminating sentinel. GCP's answer to pagination.
- [ ] **Concrete types vs interfaces** — `*storage.Client`, `*storage.BucketHandle`, `*storage.ObjectIterator` are concrete with unexported fields. You can't just pick off methods to mock; you wrap your *own* thin abstraction (3–4 method interface). This wrinkle gets its own example.
- [ ] Service account impersonation — `impersonate.CredentialsTokenSource` — short-lived credentials for a target SA, cached by the token source itself. The GCP analogue of cross-account role assumption.
- [ ] V4 Signed URLs — `bucket.SignedURL(name, &storage.SignedURLOptions{...})`. Requires a signing identity (service account email + key, or IAM `signBlob` permission). Lets a third party upload/download without GCP credentials of their own.
- [ ] GCE `AggregatedList` shape — returns a per-zone map (`map[string]InstancesScopedList`). You always flatten across zones, which is the GCP-equivalent papercut to AWS's Reservations-wrapping-Instances quirk.
- [ ] gRPC under the hood for most GCP services — surfaces as `status.Status` error types and different retry knobs than REST-based SDKs.

## Examples to build

Each is a runnable `main.go` demonstrating one specific concept.

| Folder | Demonstrates |
|---|---|
| `01-adc-and-project/` | ADC chain walk; print which source fired; resolve project ID from env / metadata. The smallest end-to-end GCP call. |
| `02-gcs-list/` | List buckets in a project; list objects in a bucket using the `Iterator` pattern. The first place you internalize iterator shape. |
| `03-gcs-upload-download/` | `Writer` for upload, `Reader` for download — round-trip a few KB to a bucket. Note the input/output type asymmetry. |
| `04-gcs-signed-url/` | V4 signed URL for a GET, 5-minute expiry, signed by a service account. The standard "let user download/upload directly" pattern. |
| `05-gce-list/` | `AggregatedList` of instances + label filter + per-zone-map flatten. |
| `06-impersonate-sa/` | `impersonate.CredentialsTokenSource(ctx, impersonate.CredentialsConfig{...})` → use the resulting `option.ClientOption` to build a GCS client that acts as the target SA. |
| `07-mocking-gcs/` | **Ships fully working.** Wraps `*storage.Client` behind a 3-method interface; ships a `fakeGCS` test that verifies a consumer without ever calling GCP. The pattern reused in the mini-project. |

## Mini-project

**`gcssync`** — mirror a local directory → a GCS bucket. Like `gsutil rsync` but smaller: walks the local tree, computes CRC32C of each file, lists the bucket, plans uploads (new/changed) + optional deletes (`--delete`), runs with bounded concurrency, supports `--dry-run`.

- Change detection: **CRC32C** (GCS's default object integrity hash), computed locally with `hash/crc32` + the Castagnoli table, compared against `attrs.CRC32C`.
- Flags: `--dry-run`, `--delete`, `--concurrency N`
- Tests use the wrapper interface from `07-mocking-gcs` and never hit GCP. Pin:
  - CRC32C match → no upload
  - `--dry-run` → zero Upload/Delete calls
  - Concurrency limit respected (peak count never exceeds `N`)
  - `--delete` removes only objects not present locally

The point: this exercises everything in the section — the Iterator pattern, `Writer`/`Reader` round-trip, `Delete`, and the testing pattern from `07-mocking-gcs`.

## Exercises

1. **`01-bucket-inventory`** — CSV of every object in every bucket in the project (bucket, name, size, updated). Practices the iterator-of-iterators pattern + project-scoped bucket listing + RFC3339 timestamps.
2. **`02-find-unlabeled`** — GCE instances missing a required label key. Practices `AggregatedList` pagination + per-zone flatten + label-map semantics (labels are `map[string]string`, not a list of tag structs).
3. **`03-cleanup-old`** — delete GCS objects under a prefix older than a cutoff (`Updated` timestamp), with `--dry-run` first. Practices listing + filtering + per-object delete loop + dry-run-makes-no-calls contract.

## Status

- [x] Concepts in README walkthrough
- [x] Examples 01–07 built
- [x] Mini-project `gcssync` built + tested
- [x] Exercises scaffolded

## Session Log

When a Claude session does work in this section, append an entry to the root [`SESSIONS.md`](../SESSIONS.md) before ending — do **not** log session history in this file. `PLAN.md` is the plan; `SESSIONS.md` is the history. Tick the Status boxes above as items complete.
