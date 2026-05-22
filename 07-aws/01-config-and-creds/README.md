# 01 — config and credentials

`config.LoadDefaultConfig(ctx)` is the SDK v2 entry point. It returns an
`aws.Config` that every service client (`s3.NewFromConfig`, `ec2.NewFromConfig`,
…) consumes.

## The credentials chain

In order, the default chain checks:

1. Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`,
   `AWS_SESSION_TOKEN`)
2. `~/.aws/credentials` (the active profile, set by `AWS_PROFILE`)
3. `~/.aws/config` (SSO, role-from-profile)
4. IMDS — the EC2/ECS instance metadata service (only when running on AWS)

Whichever step yields creds first wins. `creds.Source` tells you which one.

## Lazy retrieval

`LoadDefaultConfig` only **resolves** the chain. It does not call AWS. The
actual credential read happens when something invokes
`cfg.Credentials.Retrieve(ctx)` — explicitly, or implicitly when an SDK call
signs a request. That's why bad creds don't fail until your first real API
call.

## Compare to boto3

| | Go (`config.LoadDefaultConfig`) | Python (`boto3`) |
|---|---|---|
| Where the chain lives | `config` package | `botocore.credentials` |
| Force a profile | `config.WithSharedConfigProfile("p")` | `boto3.Session(profile_name="p")` |
| Force a region | `config.WithRegion("eu-west-1")` | `boto3.Session(region_name=...)` |

## TODO

1. Uncomment the TODO block.
2. Run `go run .` — note what region and provider come back.
3. Run `AWS_REGION=ap-south-1 go run .` — confirm the override wins.
4. Add a second `LoadDefaultConfig` call with `config.WithSharedConfigProfile("nonexistent")` and observe what fails when.
