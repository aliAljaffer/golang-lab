# Plan: 07-aws

## Prerequisites

- AWS account (free tier is enough for these examples)
- AWS CLI configured locally (`aws configure`) or env vars set
- Note: most examples need real credentials; tests that hit AWS are tagged `//go:build integration` to skip in CI

## Concepts to cover

- [ ] `aws-sdk-go-v2` vs v1 — why v2 (modular, context-aware, faster)
- [ ] `config.LoadDefaultConfig(ctx)` — the credentials chain
- [ ] Region resolution
- [ ] Service clients: one import per service (`s3`, `ec2`, `iam`, etc.)
- [ ] Operation methods: every API call takes a `context.Context` + input struct
- [ ] Paginators: SDK v2 makes them first-class
- [ ] Assume role across accounts (`stscreds`)
- [ ] Presigned URLs for time-limited access
- [ ] Mocking AWS calls in tests via interfaces

## Examples to build

| Folder | Demonstrates |
|---|---|
| `01-config-and-creds/` | Loading config, inspecting credentials, region |
| `02-s3-list/` | List buckets, list objects in a bucket (with paginator) |
| `03-s3-upload-download/` | Upload a local file, download it back |
| `04-s3-presigned/` | Generate a 5-minute presigned URL |
| `05-ec2-list/` | List EC2 instances, filter by tag |
| `06-assume-role/` | Use STS to assume a cross-account role |
| `07-mocking-sdk/` | Define an interface for the S3 operations you use; test against a fake |

## Mini-project

**`s3sync`** — mirrors a local directory to an S3 bucket. Skips files whose ETag (md5 for small files) matches. Supports `--dry-run`, `--delete` (remove S3 objects not present locally), `--concurrency N`.

Tests verify:
- Skips identical files
- `--dry-run` doesn't modify the bucket
- Concurrency limit is respected (use a mocked client to count parallel calls)

## Exercises

1. **`01-bucket-inventory`** — write a tool that lists all objects across all buckets in a region, outputs CSV
2. **`02-find-untagged`** — find EC2 instances missing a required tag (e.g. `Owner`)
3. **`03-cleanup-old`** — delete S3 objects older than N days from a given prefix (with confirmation prompt)

## Status

- [x] Concepts in README walkthrough
- [x] Examples 01-07 built
- [x] Mini-project `s3sync` built + tested
- [x] Exercises scaffolded

## Session Log

When a Claude session does work in this section, append an entry to the root [`SESSIONS.md`](../SESSIONS.md) before ending — do **not** log session history in this file. `PLAN.md` is the plan; `SESSIONS.md` is the history. Tick the Status boxes above as items complete.
