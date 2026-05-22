# 03 — get a deployment

`Get` is symmetric to `List`: same shape, takes a name, returns one object.

```go
dep, err := clientset.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
```

## API groups, briefly

| Group / Version  | What lives here                                  | Accessor           |
|------------------|--------------------------------------------------|--------------------|
| `core/v1`        | Pods, Services, ConfigMaps, Secrets, Nodes       | `CoreV1()`         |
| `apps/v1`        | Deployments, StatefulSets, DaemonSets, ReplicaSets | `AppsV1()`       |
| `batch/v1`       | Jobs, CronJobs                                   | `BatchV1()`        |
| `networking/v1`  | Ingress, NetworkPolicy                           | `NetworkingV1()`   |
| `rbac/v1`        | Roles, RoleBindings, ClusterRoles                | `RbacV1()`         |

You **always** know which group a resource is in by running
`kubectl api-resources` once and bookmarking it.

## Spec vs Status

This is the most important concept in Kubernetes. Every resource has:

- **Spec** — what you (the user) want. You write this.
- **Status** — what the controller has reconciled. The controller writes this.

For a Deployment:

| Field                            | Meaning                                              |
|----------------------------------|------------------------------------------------------|
| `dep.Spec.Replicas`              | How many pods you want                               |
| `dep.Status.Replicas`            | How many ReplicaSets are reporting                   |
| `dep.Status.ReadyReplicas`       | How many pods passed their readiness probe           |
| `dep.Status.AvailableReplicas`   | ReadyReplicas that have stayed ready ≥ MinReadySeconds |

`dep.Spec.Replicas` is a `*int32` (nullable — `nil` means "use the default,"
which is 1). Always nil-check before dereferencing.

## Error handling: `apierrors.IsNotFound`

The k8s API server returns structured errors. The `apierrors` package gives
you typed predicates:

```go
if apierrors.IsNotFound(err) { ... }
if apierrors.IsForbidden(err) { ... }
if apierrors.IsAlreadyExists(err) { ... }
if apierrors.IsConflict(err) { ... }   // optimistic concurrency loss
```

Bare `err != nil` lumps "API server unreachable" with "object not found."
Use the predicates so your tool can give the right exit code.

## TODO

1. Uncomment the TODO block.
2. Run `go run . --namespace kube-system --name coredns` (or whatever exists).
3. Try a name that doesn't exist; confirm the `IsNotFound` branch fires.
4. Add a `--watch-status` flag that polls every 2s and prints when
   `ReadyReplicas == Spec.Replicas` (rough rollout-watch).
