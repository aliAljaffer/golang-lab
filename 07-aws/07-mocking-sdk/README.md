# 07 — mocking the AWS SDK

The lesson: **don't depend on `*s3.Client`** in business code. Depend on a
small interface that names only the methods you actually call. Production
passes the real client (Go satisfies interfaces structurally — no `implements`
keyword needed). Tests pass a fake.

## The interface

```go
type S3GetObjectAPI interface {
    GetObject(ctx context.Context, in *s3.GetObjectInput, opts ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}
```

This matches `(*s3.Client).GetObject`'s signature exactly. `*s3.Client` is
automatically assignable to `S3GetObjectAPI` — no wrapper needed.

## The fake

`fakeS3` is a `struct + []callRecord` — same pattern as 06-testing's
`fakeNotifier`. It records every call and returns whatever the test set up.
No `gomock`, no `mockery`, no generated code. Just Go.

## Why this beats "real client + httptest"

For SDK v2, the request signing chain is non-trivial — wiring a real client
to point at `httptest.NewServer` is doable but noisy. An interface fake is
3 lines + the methods you care about. Reserve `httptest.NewServer` for code
where the *HTTP* behavior is what's under test (retries, headers, etc.) —
that's what example `05-httptest` in `06-testing` is for.

## One interface per slice

Don't write a god-interface listing 30 methods. Write one interface per
function or struct that uses the SDK, listing only what it touches. Search
for "interface segregation" — same idea.

## TODO

1. Read `s3util.go` + `s3util_test.go`. Run `go test ./07-aws/07-mocking-sdk/`.
2. Add an `S3PutObjectAPI` interface + a function `UploadKey(ctx, api, bucket, key, body)`. Write 2 tests using the same fake pattern.
3. Try writing the interface to ALSO embed `S3GetObjectAPI` — verify a single
   fake can satisfy both.
