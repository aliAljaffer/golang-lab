# Exercise 01 — namespace audit

Find namespaces missing a required label (e.g. `owner`). Common compliance
check at any org that mandates labels on namespaces.

## What you implement

```go
type NamespaceAPI interface {
    List(ctx context.Context, opts metav1.ListOptions) (*corev1.NamespaceList, error)
}

func Audit(ctx context.Context, api NamespaceAPI, requiredLabel string) ([]string, error)
```

A real `*kubernetes.Clientset` satisfies `NamespaceAPI` via
`clientset.CoreV1().Namespaces()` — same shape, the clientset's
`Namespaces()` method returns something with a `List` method that matches.

## Contract pinned by tests

- Returns the names of namespaces whose `.Labels` map does NOT contain
  `requiredLabel` as a key.
- An empty label value (`owner=""`) counts as "label present" — don't flag it.
- Preserves the order the API returned (no sort).
- Errors from `List` propagate.

## Run the failing suite

```bash
go test -tags=exercise ./08-kubernetes/exercises/01-namespace-audit/
```
