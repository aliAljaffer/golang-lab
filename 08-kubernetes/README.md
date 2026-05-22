# 08 — Kubernetes

> Status: ☑ scaffolded — see [`PLAN.md`](./PLAN.md)

`client-go` is the SDK for everything cluster-adjacent: admin tools, custom operators, drift detectors, automated remediation bots. The surface is large and the verbosity high, but the patterns are tight — once you've written one informer, you've essentially seen the architecture of every controller in Kubernetes.

This section walks from "load a kubeconfig and list pods" up to "watch the cluster with an informer and react to changes," which is the core loop of every controller and operator.

---

## What you'll learn

- Loading a kubeconfig: `rest.InClusterConfig()` falls back to `clientcmd.BuildConfigFromFlags("", kubeconfig)` (the canonical "in-pod OR outside" pattern)
- Querying typed resources: `clientset.CoreV1().Pods(ns).List/Get`, `clientset.AppsV1().Deployments(ns).Get`
- Label selectors and `Pods("")` for all-namespaces
- The Spec vs Status doctrine — your intent vs what the controller reconciled
- Raw `Watch()` and why production code uses informers instead
- **Informers**: shared cache + ResourceEventHandlerFuncs + resync semantics + the `DeletedFinalStateUnknown` tombstone
- `Create` / `Update` / `Delete` against typed objects, with `apierrors.IsNotFound` / `IsAlreadyExists` for idempotency

---

## Mental model from other languages

| Concept              | Go (`client-go`)                                          | Python (`kubernetes` client)                     |
| -------------------- | --------------------------------------------------------- | ------------------------------------------------ |
| Client construction  | `kubernetes.NewForConfig(cfg)`                            | `kubernetes.client.CoreV1Api()`                  |
| Get a pod            | `clientset.CoreV1().Pods(ns).Get(ctx, name, opts)`        | `api.read_namespaced_pod(name, ns)`              |
| List with selector   | `List(ctx, metav1.ListOptions{LabelSelector: "app=x"})`   | `api.list_namespaced_pod(ns, label_selector=...)` |
| All namespaces       | `Pods("")` (empty string convention)                      | `api.list_pod_for_all_namespaces()`              |
| Watch (basic)        | `Watch()` returns `watch.Interface` with `ResultChan`     | `watch.Watch().stream(...)`                      |
| Watch (production)   | **shared informers** (caching, deltas, replay)            | none — manually polled                           |
| Test fakes           | `fake.NewSimpleClientset(initialObjs...)`                 | no first-party equivalent                        |

**Go's twist:** informers are unique to `client-go`. They keep a local in-memory cache of cluster state, sync once at startup via List, then stay current via Watch with automatic reconnect + replay across etcd compactions. This is how every controller in Kubernetes — kubelet, scheduler, kube-controller-manager — works internally. The Python client has no equivalent; long-running controllers in Python typically lean on `kopf`, which is itself a port of informer concepts.

---

## The DevOps angle

The verbosity of `client-go` pays off when you write controllers. The composition (typed clientset + informer factory + workqueue + reconcile loop) is highly reusable; kubebuilder and Operator SDK both code-generate on top of it. Most "I need a small Kubernetes tool" jobs don't need the full machinery — they just need a typed client and maybe one informer, which is exactly what this section covers.

The non-obvious production details:

- **Always wire a context with cancellation through to informers.** `factory.Start(ctx.Done())` will run forever otherwise; SIGTERM gives you ~30 seconds to drain.
- **`UpdateFunc` must be idempotent.** Informer factories re-fire `UpdateFunc` on every resync (default 30 seconds), not only on real changes — if it isn't idempotent, you create N copies of work every resync.
- **`fake.NewSimpleClientset(initialObjs...)` is the gateway drug for testing.** No real cluster needed; you pre-seed objects, then the informer factory built on top delivers them as Add events. The mini-project's tests run entirely against this.

---

## Walkthrough

Read these in order. Each is a runnable example demonstrating one specific concept.

1. [`01-load-config/`](./01-load-config/) — `rest.InClusterConfig()` → fall back to `clientcmd.BuildConfigFromFlags("", kubeconfig)` → `Discovery().ServerVersion()` smoke probe. The "this binary runs both inside the cluster and from a laptop" recipe.
2. [`02-list-pods/`](./02-list-pods/) — `CoreV1().Pods(ns).List(ctx, ListOptions{LabelSelector: "app=x"})`. Empty-namespace = all-namespaces. The most-touched fields on `corev1.Pod`.
3. [`03-get-deployment/`](./03-get-deployment/) — `AppsV1().Deployments(ns).Get`, the Spec vs Status split (`Spec.Replicas` is *intent*; `Status.ReadyReplicas` is what the controller actually achieved), `apierrors.IsNotFound`.
4. [`04-watch-basic/`](./04-watch-basic/) — `Watch(ctx, opts) watch.Interface` + consuming `ResultChan`. The "raw watch dies on every connection rotation and you lose events" footgun motivates 05.
5. [`05-informer/`](./05-informer/) — `NewSharedInformerFactoryWithOptions` + `ResourceEventHandlerFuncs{AddFunc, UpdateFunc, DeleteFunc}` + `WaitForCacheSync`. The `DeletedFinalStateUnknown` tombstone is documented but the example skips handling it (one heavy concept at a time).
6. [`06-create-configmap/`](./06-create-configmap/) — `Create(ctx, obj, CreateOptions{})`, the `ObjectMeta` vs server-filled fields (`UID`, `ResourceVersion`), `IsAlreadyExists` for idempotency. Server-side apply is mentioned as the production-grade alternative.

---

## Mini-project: [`crashloop-alert`](./mini-project/)

Informer-based watcher for pods in `CrashLoopBackOff`, with dedup and a pluggable `Sink` (stdout or webhook).

The point: this is the smallest realistic Kubernetes operator pattern — watch pods, detect a condition, fire a side effect, suppress duplicates. The 11 tests cover the crash-loop detection logic, the clock-injectable `Deduper` (`Now func() time.Time` from `06-testing/02-fake-clock`), both Sink implementations (stdout writes JSON, webhook POSTs and errors on non-2xx), and a full end-to-end run via `fake.NewSimpleClientset` — no real cluster needed.

Spec and starter in [`mini-project/`](./mini-project/).

---

## Exercises

See [`exercises/`](./exercises/):

1. **[`01-namespace-audit`](./exercises/01-namespace-audit/)** — list namespaces missing a required label key. Practices `NamespaceAPI` interface (the 1-method consumption-site pattern) + the kubectl-equivalent contract that an empty *value* counts as the label being present.
2. **[`02-resource-counter`](./exercises/02-resource-counter/)** — count pods/deployments/services per namespace. Practices `ClusterAPI` with 4 methods — a judgment call vs four 1-method interfaces.
3. **[`03-rolling-restart`](./exercises/03-rolling-restart/)** — patch a deployment's pod-template annotation `kubectl.kubernetes.io/restartedAt` to trigger a rollout. Same annotation key `kubectl rollout restart` uses, so the tool's restarts are visible to `kubectl rollout status` and `kubectl rollout history`.

---

## Further reading

- [`client-go` GitHub](https://github.com/kubernetes/client-go) — source of truth; `examples/` covers patterns the docs don't
- [Kubernetes API reference](https://kubernetes.io/docs/reference/kubernetes-api/) — the typed structs in `k8s.io/api` mirror this
- [Kubernetes "Sample Controller"](https://github.com/kubernetes/sample-controller) — the canonical informer + workqueue + reconcile pattern at minimal scale
- [The Kubernetes API Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md) — why Spec/Status, why ObjectMeta, why server-side defaults
- [Kubebuilder Book](https://book.kubebuilder.io/) — the next step after this section for building real CRD-backed operators
