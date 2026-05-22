# 02 — S3 list (buckets + paginated objects)

## Service clients

Each AWS service has its own Go package: `service/s3`, `service/ec2`,
`service/sts`, … You construct a client per service from the shared `aws.Config`:

```go
cfg, _ := config.LoadDefaultConfig(ctx)
s3c := s3.NewFromConfig(cfg)
ec2c := ec2.NewFromConfig(cfg)
```

Clients are cheap and safe to share. Don't pool them.

## Operation methods

Every API call follows the same shape:

```go
out, err := client.OpName(ctx, &service.OpNameInput{ /* fields */ })
```

The input is always a pointer to a struct. Many fields are `*string` /
`*int32` — that's the SDK's way of distinguishing "not set" from "set to
zero". Use `aws.String("x")` / `aws.Int32(1)` to build them. To read, deref
with `*field` (after a nil check on optional outputs).

## Paginators

In v1 you wrote `for { ListObjects(NextMarker: ...) }` loops. In v2 the SDK
ships a paginator type per paginatable operation:

```go
p := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{Bucket: &b})
for p.HasMorePages() {
    page, err := p.NextPage(ctx)
    // ...
}
```

It hides the continuation token and lets your loop body care only about the
current page.

## TODO

1. Fill in PART 1 (ListBuckets), run `go run .`.
2. Fill in PART 2, run `go run . <one-of-your-buckets>`.
3. Add `MaxKeys: aws.Int32(2)` to the input — confirm the paginator still
   iterates every object across many small pages.
