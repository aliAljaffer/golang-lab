# Exercises — 08-kubernetes

Each subfolder is an exercise with failing tests behind `//go:build exercise`.

Run the failing suite:

```bash
go test -tags=exercise ./08-kubernetes/exercises/...
```

Default `go test ./...` stays green — exercises are tag-gated.

All exercises target `kubernetes.Interface` or a narrower interface, so tests
use `k8s.io/client-go/kubernetes/fake` and never need a real cluster.

| # | Folder | What you implement |
|---|---|---|
| 01 | `01-namespace-audit/` | `Audit(ctx, api, requiredLabel) ([]string, error)` — list namespaces missing a required label |
| 02 | `02-resource-counter/` | `Count(ctx, api) ([]Row, error)` — count pods/deployments/services per namespace |
| 03 | `03-rolling-restart/` | `RollingRestart(ctx, api, ns, name) error` — patch a Deployment to roll its pods |

The lesson worth repeating: **each exercise defines its OWN narrow interface
naming only the methods it needs.** That's how you make production code
testable without dragging the whole clientset into your business logic.

See parent [`PLAN.md`](../PLAN.md).
