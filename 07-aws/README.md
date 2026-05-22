# 07 — AWS SDK

> Status: ☑ scaffolded — see [`PLAN.md`](./PLAN.md)

If you're on AWS, sooner or later you write a tool that talks to AWS. The Go SDK v2 (`aws-sdk-go-v2`) is more verbose than `boto3` — typed input/output structs, separate paginator iterators, explicit context plumbing — but the verbosity buys you compile-time errors instead of runtime `KeyError: 'Bucket'` traces in CI.

This section walks the credentials chain, the S3 and EC2 surfaces you'll touch most, cross-account assume-role, and — most importantly — the **interface-at-the-consumption-site** pattern for testing SDK code without hitting AWS.

---

## What you'll learn

- `aws-sdk-go-v2` — the current generation SDK (v1 is deprecated; don't start anything new on it)
- The credentials chain: env vars → shared config file → IAM role → EC2/ECS metadata → SSO
- S3 operations: `ListBuckets`, `ListObjectsV2` (with the paginator), `PutObject`/`GetObject` round-trip, presigned URLs
- EC2: `DescribeInstances` paginator + `Reservations → Instances` flatten, filter DSL
- IAM: assume role across accounts with `sts.NewFromConfig` + `stscreds.NewAssumeRoleProvider` + `aws.NewCredentialsCache`
- Pagination patterns (the SDK v2 way — `NewXxxPaginator(client, input)` + `for p.HasMorePages()`)
- Mocking the SDK by defining an `S3API` interface in your code, not in the SDK

---

## Mental model from other languages

| Concept             | Go (`aws-sdk-go-v2`)                  | Python (boto3)                            |
| ------------------- | ------------------------------------- | ----------------------------------------- |
| Client construction | `s3.NewFromConfig(cfg)`               | `boto3.client('s3')`                      |
| Paginator           | `s3.NewListObjectsV2Paginator(...)`   | `client.get_paginator(...)`               |
| Credentials chain   | `config.LoadDefaultConfig(ctx)`       | implicit via env / `~/.aws/credentials`   |
| Assume role         | `stscreds.NewAssumeRoleProvider`      | `boto3.Session(profile_name=...)`         |
| Retry               | built into client (configurable)      | botocore retry config                     |
| Mock for tests      | define interface at consumption site  | `moto` / `unittest.mock`                  |

**The cultural difference:** SDK v2 separates "config" (region, retries, creds) from "client" (a `*s3.Client` constructed from that config). You almost always build one `aws.Config` near `main`, then pass `*s3.Client` / `*ec2.Client` constructed from it down through your code. Don't pass the `aws.Config` everywhere — pass the typed clients. This is also why mocking is *easy*: your function takes a 3-method `S3API` interface, not the 100-method `*s3.Client`.

---

## The DevOps angle

The boto3 user's instinctive complaint about SDK v2 is "why is everything more typing for the same call?" The answer earns itself the first time you refactor: rename a field, the compiler tells you every consumer to update. Boto3 lets you misspell `'CONTENTS'` and crash at runtime in production.

The non-obvious production details:

- **Always use `aws.NewCredentialsCache(provider)` when assuming a role.** Otherwise every API call re-calls `sts:AssumeRole` (a 1-second round-trip), and your tool gets throttled on STS before it makes its first S3 request.
- **Always use the paginator for `ListObjectsV2` / `DescribeInstances` / `ListBuckets`-style ops.** A bucket with 1,000,001 objects returns the first 1,000 with `NextContinuationToken` set; if you ignore it, you process the first 1,000 and silently miss the rest.
- **`EC2.DescribeInstances` returns `Reservations`, not `Instances` directly.** You always flatten — instances live nested under reservations, which is a leak of the old EC2 reservation model nobody uses anymore.

---

## Walkthrough

Read these in order. Each is a runnable example demonstrating one specific concept.

1. [`01-config-and-creds/`](./01-config-and-creds/) — `config.LoadDefaultConfig(ctx)` walks the credentials chain; `cfg.Credentials.Retrieve(ctx)` shows you which provider actually fired. The smallest end-to-end SDK call.
2. [`02-s3-list/`](./02-s3-list/) — `ListBuckets` + `NewListObjectsV2Paginator`. The first place you internalize the paginator iterator shape.
3. [`03-s3-upload-download/`](./03-s3-upload-download/) — `PutObject` with `Body: io.Reader` and `GetObject` with `Body: io.ReadCloser`. Round-trip a few KB to a bucket.
4. [`04-s3-presigned/`](./04-s3-presigned/) — `s3.NewPresignClient(client)` + `s3.WithPresignExpires(5*time.Minute)`. Generates a URL someone else can use without credentials — the standard "let user upload directly" pattern.
5. [`05-ec2-list/`](./05-ec2-list/) — `DescribeInstancesPaginator` + the Filter DSL (`Filters: []types.Filter{{Name: aws.String("tag:env"), Values: []string{"prod"}}}`) + the Reservations→Instances flatten.
6. [`06-assume-role/`](./06-assume-role/) — cross-account: `sts.NewFromConfig(cfg)` → `stscreds.NewAssumeRoleProvider(stsClient, roleArn)` → wrap with `aws.NewCredentialsCache` → build a new `aws.Config` whose `Credentials` is the cache → build target-account clients from it.
7. [`07-mocking-sdk/`](./07-mocking-sdk/) — **ships fully working.** Defines a 1-method `S3API` interface, a consumer that takes it, and a `fakeS3` test that verifies the consumer without ever calling AWS. The pattern reused in every later section.

---

## Mini-project: [`s3sync`](./mini-project/)

Mirror a local directory → an S3 bucket. Like `aws s3 sync` but smaller: walks the local tree, computes MD5 of each file, lists the bucket, plans uploads (new/changed) + optional deletes (`--delete`), runs with bounded concurrency, supports `--dry-run`.

The point: this exercises everything in the section — `ListObjectsV2`'s ETag, `PutObject` with a streamed body, `DeleteObject`, and the testing pattern from `07-mocking-sdk`. The 14 tests use a `fakeS3` and never touch AWS; they pin concurrency-peak, dry-run-makes-no-calls, ETag-quote unwrapping, and the `--force` flag actually threading through to `RemoveOptions`.

Spec and starter in [`mini-project/`](./mini-project/).

---

## Exercises

See [`exercises/`](./exercises/):

1. **[`01-bucket-inventory`](./exercises/01-bucket-inventory/)** — list every object in every bucket of the account; emit CSV (bucket, key, size, last-modified). Practices the paginator-of-paginators + RFC3339 timestamps + flatten.
2. **[`02-find-untagged`](./exercises/02-find-untagged/)** — list EC2 instances missing a required tag key. Practices `DescribeInstances` pagination + the must-consume-both-pages contract + reservation flatten.
3. **[`03-cleanup-old`](./exercises/03-cleanup-old/)** — delete S3 objects under a prefix older than a cutoff date, with `--dry-run`. Practices the listing + filtering + per-key delete loop + dry-run-makes-no-calls contract.

---

## Further reading

- [AWS SDK for Go v2 Developer Guide](https://aws.github.io/aws-sdk-go-v2/docs/) — the canonical reference
- [SDK v2: migrating from v1](https://aws.github.io/aws-sdk-go-v2/docs/migrating/) — the rationale for the surface change
- [IAM credentials chain order](https://docs.aws.amazon.com/sdkref/latest/guide/standardized-credentials.html) — the cross-SDK spec example 01 demonstrates
- [`aws-sdk-go-v2` GitHub](https://github.com/aws/aws-sdk-go-v2) — source of truth; check the `service/<svc>/types/types.go` for the typed structs
- [The pinned testing doctrine: small interfaces at consumption site](./07-mocking-sdk/) — example 07 in this section
