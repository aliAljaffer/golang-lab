# Exercise 02 — `find-untagged`

Find every EC2 instance missing a required tag (e.g. `Owner`). Classic
compliance scan.

## What to implement

```go
func FindUntagged(ctx context.Context, api EC2API, requiredKey string) ([]string, error)
```

Return instance IDs in the order EC2 returned them across all pages and
reservations.

## How to test

```bash
go test -tags=exercise ./07-aws/exercises/02-find-untagged/...
```

5 tests cover:

- A simple flag-the-missing case
- Empty tag VALUE counts as "tag is present" (key existence is what matters)
- Multiple Reservations flatten into one ordered slice
- Pagination — the test serves two pages with NextToken between them
- Errors bubble up unchanged

## Hints

- The pattern from example 05:

```go
for _, r := range page.Reservations {
    for _, inst := range r.Instances {
        // ...
    }
}
```

- To check if a tag KEY is present, walk `inst.Tags` once. Don't over-engineer
  with a `map[string]bool` — the slice is small and the linear scan is clearer.
- `ec2.NewDescribeInstancesPaginator(api, &ec2.DescribeInstancesInput{})` —
  same paginator shape as S3.
