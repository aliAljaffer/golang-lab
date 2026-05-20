# 07 — AWS SDK

> Status: ☐ not started — see [`PLAN.md`](./PLAN.md)

## What you'll learn

- Using `aws-sdk-go-v2` — the current generation SDK
- The credentials chain (env vars → config files → IAM role)
- S3 operations: list, upload, download, presigned URLs
- EC2: list instances, filter by tags
- IAM: assume role across accounts
- Pagination patterns (the SDK v2 way)

## Mental model from other languages

| Concept | Go (`aws-sdk-go-v2`) | Python (boto3) |
|---|---|---|
| Client construction | `s3.NewFromConfig(cfg)` | `boto3.client('s3')` |
| Paginator | `s3.NewListObjectsV2Paginator(...)` | `client.get_paginator(...)` |
| Credentials chain | `config.LoadDefaultConfig` | implicit via env / `~/.aws` |
| Assume role | `stscreds.NewAssumeRoleProvider` | `boto3.Session(profile_name=...)` |
| Retry | built into client (configurable) | botocore retry config |

## The DevOps angle

If you're on AWS, you'll write Go tools that talk to AWS. The SDK v2 is more verbose than boto3 but type-safe — no more dictionary typos in resource names.

See [`PLAN.md`](./PLAN.md).
