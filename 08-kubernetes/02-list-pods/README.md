# 02 — list pods

`clientset.CoreV1().Pods(ns).List(ctx, opts)` is the shape every "list things"
call follows in client-go:

```go
clientset.<GroupVersion>().<ResourceType>(<namespace>).List(<ctx>, <opts>)
```

Swap `CoreV1` for `AppsV1` and `Pods` for `Deployments` and you've got the
same call for deployments. The shape is the lesson.

## Label selectors filter server-side

`metav1.ListOptions{LabelSelector: "app=nginx"}` is passed in the URL query
(`?labelSelector=app%3Dnginx`). The API server applies it before returning.
This matters when listing hundreds of pods — you don't want them all on the
wire just to drop most of them client-side.

The selector grammar is the standard k8s one:

| Selector              | Meaning                          |
|-----------------------|----------------------------------|
| `app=nginx`           | label `app` equals `nginx`       |
| `app!=nginx`          | label `app` not equal to `nginx` |
| `env in (prod, qa)`   | label `env` is one of            |
| `tier`                | label `tier` exists              |
| `!tier`               | label `tier` does NOT exist      |

You can combine with commas: `app=nginx,env=prod`.

## All-namespaces trick

`Pods("")` (empty namespace string) returns pods from every namespace.
That's a documented client-go convention. Compare to `kubectl get pods -A`.

## What's in a Pod object

`corev1.Pod` is a big struct, but the fields you'll touch 90% of the time:

- `pod.Name`, `pod.Namespace` (from embedded `ObjectMeta`)
- `pod.Labels`, `pod.Annotations`
- `pod.Status.Phase` — "Pending", "Running", "Succeeded", "Failed", "Unknown"
- `pod.Status.ContainerStatuses[i].State.Waiting.Reason` — where you find
  "CrashLoopBackOff", "ImagePullBackOff", etc.
- `pod.Spec.NodeName` — which node it's scheduled on (empty if Pending)

## TODO

1. Uncomment the TODO block.
2. Run `go run . --namespace kube-system` against your cluster.
3. Add `-o json`-style output: marshal each pod to JSON and print.
4. Make `--selector` accept multiple comma-joined key=value pairs and verify
   they all apply (`app=nginx,env=prod`).
