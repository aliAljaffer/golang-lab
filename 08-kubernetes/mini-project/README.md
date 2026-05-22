# Mini-project — `crashloop-alert`

Watches every pod in the cluster via a SharedInformer; when a pod's container
enters `CrashLoopBackOff`, emits an alert (stdout line or POST to a webhook
URL). Deduplicates: the same pod can't re-alert within a cooldown window.

## Spec

- Builds a `kubernetes.Clientset` from in-cluster or kubeconfig (same loader
  as the examples).
- Builds an informer factory; registers an `Add` + `Update` handler on Pods.
- The handler:
  1. Type-asserts the obj to `*corev1.Pod`.
  2. Calls `IsCrashLooping(pod)`. Skip if not crashlooping.
  3. Builds a dedup key = `<namespace>/<name>`. Asks the `Deduper` if we
     should alert. Skip if rate-limited.
  4. Calls `sink.Send(alert)`. Logs send errors but keeps going.
- The `Sink` interface has one implementation each for stdout and HTTP-POST.
- Flags: `--namespace` (default all), `--cooldown` (default 5m), `--webhook`
  (default empty → stdout).

## Testable surface

```go
func IsCrashLooping(pod *corev1.Pod) bool             // pure
type Deduper struct{ ... }                            // injected clock
type Alert struct{ Namespace, Name, Reason string }
type Sink interface{ Send(context.Context, Alert) error }
type StdoutSink struct{ Out io.Writer }
type WebhookSink struct{ URL string; Client *http.Client }
```

## What the tests verify

| Test                                          | Concept                       |
|-----------------------------------------------|-------------------------------|
| `TestIsCrashLooping_DetectsWaitingReason`     | The detection contract        |
| `TestIsCrashLooping_RunningPodNotCrashLooping`| Negative path                 |
| `TestIsCrashLooping_PodWithMultipleContainers`| Any container crashing counts |
| `TestDeduper_FirstAlertPasses`                | Initial state                 |
| `TestDeduper_BlocksWithinCooldown`            | Cooldown enforcement          |
| `TestDeduper_AlertsAgainAfterCooldown`        | Window resets                 |
| `TestDeduper_PerPodIsolation`                 | Keys are independent          |
| `TestStdoutSink_WritesLine`                   | Stdout sink shape             |
| `TestWebhookSink_PostsJSON`                   | Webhook sink shape            |
| `TestRun_InformerFiresOnCrashLoopingPod`      | End-to-end with fake clientset|

All tests run against a fake clientset (`k8s.io/client-go/kubernetes/fake`)
or pure functions — no real cluster needed.

## How to run (once you've implemented it)

```bash
# stdout mode
go run ./08-kubernetes/mini-project --cooldown 30s

# webhook mode
go run ./08-kubernetes/mini-project --webhook https://hooks.example/alerts
```

## Notes on the fake clientset

`k8s.io/client-go/kubernetes/fake.NewSimpleClientset(...)` returns a
clientset that satisfies `kubernetes.Interface`. You build an informer
factory off it the same way as in production:

```go
clientset := fake.NewSimpleClientset(initialObjs...)
factory := informers.NewSharedInformerFactory(clientset, 0)
```

To inject an event during a test, just do `clientset.CoreV1().Pods(ns).Create(...)`
on the fake — the informer's ListWatcher (which is reading from the fake)
will deliver it to your handler.
