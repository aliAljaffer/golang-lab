# 08 — Kubernetes

> Status: ☐ not started — see [`PLAN.md`](./PLAN.md)

## What you'll learn

- Loading a kubeconfig with `client-go`
- Querying resources (pods, deployments, namespaces)
- Watching resources via informers (the k8s way of "subscribe to changes")
- A taste of operator/controller patterns

## Mental model from other languages

| Concept | Go (`client-go`) | Python (`kubernetes` client) |
|---|---|---|
| Client construction | `kubernetes.NewForConfig(cfg)` | `kubernetes.client.CoreV1Api()` |
| Get a pod | `clientset.CoreV1().Pods(ns).Get(ctx, name, opts)` | `api.read_namespaced_pod(name, ns)` |
| Watch (basic) | `Watch()` returns event channel | `watch.Watch().stream(...)` |
| Watch (production) | **informers** (caching, deltas) | none — manually polled |

**Go advantage:** informers are unique to client-go — they maintain a local cache of the cluster state and emit add/update/delete events. This is how every controller in k8s works internally.

## The DevOps angle

This is the SDK for everything cluster-adjacent: cluster admin tools, custom operators, drift detectors, automated remediation bots. The verbosity of client-go pays off when you write controllers — the patterns are highly composable.

See [`PLAN.md`](./PLAN.md).
