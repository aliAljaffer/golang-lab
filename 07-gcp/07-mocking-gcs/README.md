# 07 — mocking GCS

The lesson: **don't depend on `*storage.Client` in business code.** Depend
on a small interface that names only the operations you actually call.
Production passes the real client (wrapped); tests pass a fake.

This is the same doctrine as `07-aws/07-mocking-sdk` — but GCS adds a
wrinkle.

## The wrinkle: concrete types, not method-on-client

AWS SDK v2:

```go
type S3GetObjectAPI interface {
    GetObject(ctx, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}
// *s3.Client satisfies S3GetObjectAPI automatically. Production = pass *s3.Client.
```

GCS:

```go
// You'd LIKE to write this:
type GCS interface {
    Bucket(string) *storage.BucketHandle
}
// ...but *storage.BucketHandle is a concrete type with unexported fields.
// Tests can't construct one. There's no way to mock the chain.
```

The chain `client.Bucket(b).Object(k).NewReader(ctx)` produces concrete types
all the way down. You can't pick off methods piecemeal.

## The fix: wrap your own thin abstraction

Define an interface (`GCSAPI`) whose methods are the operations *your code*
needs — Read, Write, List, whatever. The interface lives in your package.
Then write:

- `RealGCS` — implements `GCSAPI` by driving a real `*storage.Client`.
  Production passes one of these.
- `fakeGCS` (in `*_test.go`) — implements `GCSAPI` with hand-rolled in-memory
  behavior. Tests pass one of these.

Consumers depend on `GCSAPI`, not `*storage.Client`. The chain-of-concretes
problem is isolated to `RealGCS`, where you actually want it.

## The interface here: 3 methods

```go
type GCSAPI interface {
    Read(ctx, bucket, key string, maxBytes int64) ([]byte, error)
    Write(ctx, bucket, key string, body []byte) error
    List(ctx, bucket, prefix string) ([]ObjectInfo, error)
}
```

Small enough that a hand-rolled fake fits in 30 lines. The fake doesn't
need to simulate streaming, iterator behavior, or 6-step upload protocols —
the wrapper hides all that.

## Compare to AWS

|                                | Go (GCS)                                  | Go (AWS S3)                             |
| ------------------------------ | ----------------------------------------- | --------------------------------------- |
| Can the SDK client satisfy a slim interface directly? | No — chain returns concrete types | Yes — methods live on the client |
| Adapter needed?                | Yes (`RealGCS`)                           | No                                      |
| Test fake complexity           | matches your abstraction, not SDK's API   | matches the SDK method signatures       |
| Lines of code for the pattern  | ~70 (wrapper + fake)                      | ~40 (interface + fake)                  |

## When the wrapper is overkill

If a consumer needs only one GCS operation, you can drop the interface
entirely and define a tiny function-typed parameter:

```go
func FetchKey(ctx context.Context, read func(string, string) ([]byte, error), bucket, key string) ([]byte, error)
```

For multi-method consumers, the interface scales better — but for a 1-method
need, a function value is the lowest-ceremony option.

## TODO (this example ships fully working — `go test ./07-gcp/07-mocking-gcs/`)

1. Read `gcsutil.go` + `gcsutil_test.go`. Run the tests.
2. The `Write` and `List` paths have consumers (`TotalSize` is one) but only
   `TotalSize` has tests. Add a `Mirror(ctx, api GCSAPI, src, dst string)`
   that reads every object under one prefix and writes it to another, with
   tests using `fakeGCS`.
3. Extend `ObjectInfo` with `ContentType` and add a consumer + test pair
   that filters by content type.
4. The `mini-project/gcssync/` in this section reuses this exact pattern —
   read its `S3API`-equivalent interface and notice the symmetry.
