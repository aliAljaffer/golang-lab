# 04 — presigned URLs

A presigned URL is a regular HTTPS URL with a baked-in signature that grants
a single operation (GetObject, PutObject, …) on a single object for a bounded
window. Anyone holding the URL can perform that operation. The signer must
hold credentials with the underlying permission.

## When to use

- Letting a browser upload directly to S3 without proxying through your server
- Issuing time-limited download links to non-AWS users
- Webhook payload "here's a one-off link to fetch the result"

## Mechanics

`s3.NewPresignClient(client).PresignGetObject(ctx, input, opts...)` does the
signing **locally** — no AWS call is made. So calling Presign isn't billable
and doesn't burn rate limits. The returned `req.URL` is the result.

The expiry is set with `s3.WithPresignExpires(d)`. Max 7 days when using
SigV4 with long-term creds; 12 hours with role-derived (session) creds.

## Caveat — clock skew

The signature is time-bound. If the machine signing the URL has a clock more
than ~5 minutes off from AWS time, the URL is invalid the moment it's issued.
NTP isn't optional in production.

## TODO

1. Uncomment the TODO block.
2. Upload a file (example 03), then presign it: `go run . my-bucket test.txt`
3. `curl -v "$(go run . my-bucket test.txt)"` — should 200 the first time.
4. Wait 6 minutes, repeat — should 403 with `<Code>AccessDenied</Code>`.
