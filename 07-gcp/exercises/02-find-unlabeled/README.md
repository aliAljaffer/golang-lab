# 02 — find unlabeled GCE instances

List GCE instances missing a required label key. Compliance staple — your
billing team wants every VM tagged with Owner / CostCenter / Env so charges
can be allocated.

## Surface

```go
type ComputeAPI interface {
    AggregatedListInstances(ctx, project string) ([]InstanceSummary, error)
}

func FindUnlabeled(ctx, api, project, requiredKey string) ([]string, error)
```

The wrapper flattens the per-zone map (see example `05-gce-list/` for the
raw shape) into one flat slice. Tests don't need to construct
`computepb.InstancesScopedList` values.

## Tests pin

- a missing-key instance is flagged
- empty value (`env=""`) counts as present (key existing is what matters)
- order across instances/zones is preserved
- empty requiredKey is a validation error
- all-instances-labeled → empty result, no error

## What you're practicing

- **Label-map semantics.** GCE labels are `map[string]string`, server-modeled.
  Different from AWS tags (which are a `[]Tag{Key, Value}` slice). Same
  observable behaviour — different shape on the wire.
- **Per-zone-map flatten.** Real `AggregatedList` returns
  `map[string]InstancesScopedList`; the wrapper hides that.
- **Empty-vs-missing distinction.** `m["env"]` returns the zero value for a
  missing key; you need `_, ok := m["env"]` (or check the second result of
  the lookup) to distinguish `env=""` from "env not set."
