# 05 — GCE list (AggregatedList + label filter)

GCE instances are zonal. Listing "all instances in a project" means a list
call per zone — but the API exposes a convenience: `AggregatedList` rolls
every zone into one paginated response.

## The shape

```go
it := client.AggregatedList(ctx, &computepb.AggregatedListInstancesRequest{
    Project: "my-project",
    Filter:  proto.String("labels.env=prod"),
})
for {
    pair, err := it.Next()  // pair.Key = "zones/us-central1-a", pair.Value.Instances = []*Instance
    if errors.Is(err, iterator.Done) { break }
    if err != nil { /* real error */ }
    for _, inst := range pair.Value.Instances {
        // use inst
    }
}
```

You **always flatten across zones.** Production tools never iterate one zone
at a time — `AggregatedList` is the API you reach for.

## The empty-zone wrinkle

`AggregatedList` returns one entry per zone *even if the zone has no
instances*. A small project gets ~40 empty zones and 1-2 populated ones.
Skip empty zones explicitly (`if pair.Value == nil || len(pair.Value.Instances) == 0 { continue }`) — your output should be the populated zones, not "zones/us-west4-a (empty)" 40 times.

## The Filter DSL

GCP's server-side filter is a typed expression language:

- `labels.env=prod` — equality
- `labels.env!=staging` — inequality
- `(labels.env=prod) AND (status=RUNNING)` — composition
- `name:nginx-*` — prefix (the `:` operator is "has prefix")

This is GCP-wide, not just compute — Cloud Logging, Cloud Asset Inventory,
GCS bucket listing, BigQuery dataset listing all use a similar DSL.

## Labels (not tags)

GCE labels are `map[string]string`, server-modeled. AWS calls them tags;
GCP also has a separate "tag" concept that's role-binding metadata, not
key-value descriptors. Labels are the equivalent of AWS tags for grouping
and filtering.

## Compare to AWS

| Concept                   | Go (GCE `AggregatedList`)                | Go (AWS `DescribeInstances`)                  |
| ------------------------- | ---------------------------------------- | --------------------------------------------- |
| Cross-zone/region list    | one call: `AggregatedList(ctx, req)`     | one call, but `Reservations[].Instances[]` flatten |
| Per-zone grouping         | `pair.Key = "zones/<zone>"`              | `inst.Placement.AvailabilityZone`             |
| Server-side filter        | `Filter: "labels.k=v"` (typed DSL)       | `Filters: []Filter{{Name, Values}}`           |
| Empty-zone handling       | nil-check `pair.Value`                   | empty `Reservations` slice                    |

## TODO

1. Uncomment the iterator block. Run `go run . <project>`. Confirm output covers
   every populated zone you expected.
2. Spin up a test instance with `gcloud compute instances create test-vm --labels=env=demo --zone=us-central1-a --machine-type=e2-micro`. Run
   `go run . <project> env demo` — confirm only the test VM is listed.
3. `gcloud compute instances delete test-vm --zone=us-central1-a -q` when done.
4. Try `Filter: "status=TERMINATED"` (stopped instances). Useful for
   janitor scripts.
