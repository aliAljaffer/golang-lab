# 05 — informers

The production pattern for "react to cluster state." Every Kubernetes
controller (Deployment, ReplicaSet, your own operator) is built on this.

## Three pieces

```text
+-----------------+      +----------------+      +-----------------+
|   ListWatcher   | ---> |  DeltaFIFO     | ---> | event handlers  |
| (initial List + |      |  + local Store |      | (your code)     |
|   Watch loop)   |      | (the "cache")  |      |                 |
+-----------------+      +----------------+      +-----------------+
```

1. **ListWatcher** runs the initial `List(...)`, then `Watch(...)` from the
   returned resourceVersion. On disconnect or `Error` event, it re-Lists and
   re-Watches. You never write this loop.
2. **DeltaFIFO + Store** = a thread-safe local cache keyed by `<namespace>/<name>`.
   Lookups are local — no API call to ask "what was that pod's labels?"
3. **Event handlers** fire AFTER the cache has been updated with the change.
   Handler ordering: `Add` (initial pods + new ones), `Update` (any change,
   including periodic resyncs), `Delete` (final removal).

## resync vs real updates

`NewSharedInformerFactory(clientset, 30*time.Second)` sets a *resync period*.
Every 30s the informer re-fires `UpdateFunc` for every cached object — even
if nothing changed. This is a feature, not a bug: it gives your handler a
chance to retry work it failed on last time without you having to track
retries.

This means **your UpdateFunc must be idempotent**, and it should early-exit
on no-op updates (compare `oldObj` to `newObj`). Spamming work on every
resync is the most common informer bug.

## The DeleteFunc tombstone case

Sometimes the informer misses the actual delete event and only knows the
object is gone after a re-List. In that case `obj` in `DeleteFunc` is
`*cache.DeletedFinalStateUnknown`, not your resource type. Handle both:

```go
DeleteFunc: func(obj interface{}) {
    pod, ok := obj.(*corev1.Pod)
    if !ok {
        tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
        if !ok { return }
        pod, ok = tombstone.Obj.(*corev1.Pod)
        if !ok { return }
    }
    // ... handle pod
}
```

For the example we skip the tombstone path — it's a thing to know exists.

## WaitForCacheSync

`factory.Start(stopCh)` returns immediately. `cache.WaitForCacheSync(...)`
blocks until each informer's initial List has populated the store. **Never
read from the cache or fire your handlers before this returns.**

## Compare to raw Watch (example 04)

|                                | Raw Watch       | Informer                       |
|--------------------------------|-----------------|--------------------------------|
| Initial state                  | Not delivered   | Delivered as Add events        |
| Reconnect on disconnect        | Manual          | Automatic                      |
| Local cache lookup             | None            | `informer.GetStore().GetByKey` |
| Multiple consumers, one stream | No              | Yes (shared factory)           |
| Periodic re-fire of state      | No              | Yes (resync period)            |

Use informers for everything except prototypes.

## TODO

1. Uncomment the TODO block.
2. Run `go run . --namespace ""` (all namespaces) against a real cluster.
3. Add a counter (atomic) that increments on every Add/Update/Delete and
   prints every 10 seconds.
4. Read the cache directly via `podInformer.Informer().GetStore().List()` —
   confirm it matches `kubectl get pods -A` length.
