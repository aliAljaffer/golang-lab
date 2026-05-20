# Plan: 08-kubernetes

## Prerequisites

- A k8s cluster: `minikube`, `kind`, or `colima --kubernetes` for local
- `kubectl` configured (`~/.kube/config` populated)

## Concepts to cover

- [ ] `client-go` packages: `clientset`, `dynamic`, `discovery`, `informers`
- [ ] Loading kubeconfig: in-cluster vs out-of-cluster
- [ ] The API groups + versions structure (`CoreV1()`, `AppsV1()`, etc.)
- [ ] List, Get, Create, Update, Delete — the basic verbs
- [ ] `metav1.ListOptions` — label/field selectors
- [ ] Watching: raw `Watch()` vs informers (and why informers win in production)
- [ ] Informer pattern: ListAndWatch + cache + event handlers
- [ ] Brief mention of building controllers/operators (don't implement; just orient)

## Examples to build

| Folder | Demonstrates |
|---|---|
| `01-load-config/` | Load kubeconfig out-of-cluster, print server version |
| `02-list-pods/` | List pods in a namespace; filter by label |
| `03-get-deployment/` | Get a specific deployment, print its replica count |
| `04-watch-basic/` | Raw Watch() — observe pod events |
| `05-informer/` | SharedInformer with add/update/delete handlers |
| `06-create-configmap/` | Create a configmap programmatically |

## Mini-project

**`crashloop-alert`** — watches pods cluster-wide via an informer; when a pod enters CrashLoopBackOff, prints to stdout (or POSTs to a configurable webhook URL). Dedups alerts (don't re-fire for the same pod within N minutes).

Tests verify:
- Detects the CrashLoopBackOff state correctly (use a fake clientset)
- Dedup logic works (table tests with synthetic events)

## Exercises

1. **`01-namespace-audit`** — list all namespaces missing a required label (e.g. `owner`)
2. **`02-resource-counter`** — count pods, deployments, services per namespace; output table
3. **`03-rolling-restart`** — trigger a rolling restart of a deployment (patch the pod template annotation)

## Status

- [ ] Concepts in README walkthrough
- [ ] Examples 01-06 built
- [ ] Mini-project `crashloop-alert` built + tested
- [ ] Exercises scaffolded
