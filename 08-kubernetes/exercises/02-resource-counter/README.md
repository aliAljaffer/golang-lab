# Exercise 02 — resource counter

Tally pods, deployments, and services per namespace. Common dashboard data.

## What you implement

```go
type ClusterAPI interface {
    ListNamespaces(ctx, opts) (*corev1.NamespaceList, error)
    ListPods(ctx, ns, opts) (*corev1.PodList, error)
    ListDeployments(ctx, ns, opts) (*appsv1.DeploymentList, error)
    ListServices(ctx, ns, opts) (*corev1.ServiceList, error)
}

func Count(ctx, api ClusterAPI) ([]Row, error)
```

`Row{Namespace, Pods, Deployments, Services int}` — one per namespace.

## Contract pinned by tests

- One Row per namespace returned by `ListNamespaces`.
- Tally = `len(...List.Items)`.
- Order preserved from `ListNamespaces`.
- Errors propagate (`errors.Is` predicate).

## Wiring to a real clientset

```go
clientset, _ := kubernetes.NewForConfig(cfg)
api := realAPI{cs: clientset}  // tiny adapter
rows, _ := counter.Count(ctx, api)
```

Where `realAPI.ListPods` is just
`r.cs.CoreV1().Pods(ns).List(ctx, opts)`. Two lines of glue per method.
The narrow interface is what made the test-side fake possible.

## Run the failing suite

```bash
go test -tags=exercise ./08-kubernetes/exercises/02-resource-counter/
```
