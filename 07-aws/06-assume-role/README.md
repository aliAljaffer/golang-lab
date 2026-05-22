# 06 — assume role (STS, cross-account)

`AssumeRole` swaps your current principal for a different one for a bounded
session — minutes to hours. It's the canonical pattern for:

- A CI runner in account A pushing artifacts to S3 in account B
- A central security tool with permission to read every account's IAM
- A least-privilege "deploy" role that escalates from a long-lived dev role

## The chain you build

```
source creds → STS client → AssumeRoleProvider → new config → service client
```

Each step is its own object. The last step is the only one that's billable.

## Why `aws.NewCredentialsCache`

`AssumeRoleProvider` calls STS every time it's asked for fresh creds. Wrap
it with `aws.NewCredentialsCache(provider)` and the SDK keeps the result in
memory until expiry (default 15 min), refreshing automatically. Without the
cache you'll burn an STS API call per AWS request — measurable on hot paths,
and visible in CloudTrail.

## The role's trust policy

Before this works, the role you're assuming must trust your source principal.
Trust policy on the target role:

```json
{
  "Effect": "Allow",
  "Principal": { "AWS": "arn:aws:iam::SOURCE-ACCOUNT:user/your-user" },
  "Action": "sts:AssumeRole"
}
```

Errors like `AccessDenied: User ... is not authorized to perform: sts:AssumeRole`
mean the trust policy doesn't cover you — that's an *IAM* problem, not a
*code* problem.

## TODO

1. Create a role in your account (or a friend's) and attach a trust policy
   for your current IAM principal.
2. Fill in PARTs 2-4.
3. Run `go run . arn:aws:iam::ACCOUNT:role/ROLE` — expect a bucket count.
4. Print `provider.Retrieve(ctx).Source` to confirm `AssumeRoleProvider` is
   in the chain.
