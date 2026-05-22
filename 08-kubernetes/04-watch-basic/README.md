# 04 — raw Watch()

`Pods(ns).Watch(ctx, opts)` opens a long-lived HTTP stream to the API server.
The server pushes events as JSON objects on that stream until something
breaks the connection.

## The event shape

Each event is a `watch.Event`:

```go
type Event struct {
    Type   EventType   // Added, Modified, Deleted, Bookmark, Error
    Object runtime.Object
}
```

`Object` is the resource type you're watching — `*corev1.Pod` here. You have
to type-assert because `Watch` is generic over `runtime.Object`.

## Why raw watches are dangerous

The connection WILL die. The k8s API server:

- Rotates connections on a timeout (default ~5-10 min)
- Sends `Bookmark` events to mark a resource version
- Sends `Error` events when the resource version you started from has been
  compacted (etcd garbage collects history)

Naïve code that loops on `ResultChan()` until close, then quits, will miss
every event after the first connection drop — which on a busy cluster is
"most of them, very quickly."

## What informers do better (see example 05)

`cache.NewListWatcher` + `cache.NewSharedInformer`:

1. **Initial List** — fetch every existing pod, populate a local cache
2. **Watch from the resource version** the List returned
3. On error / disconnect, automatically **re-List + re-Watch**
4. Emit `Add`/`Update`/`Delete` events from the cache, never from the wire
5. Multiple consumers share one watch stream (the "shared" in shared informer)

That's the production pattern. Raw Watch is for prototyping or one-shot
tools.

## When to use raw Watch anyway

- "Wait until pod X reaches Running, then exit." (Single event, short-lived.)
- Tests where you control the cluster.
- `kubectl get pod -w` is implemented with raw Watch — it's not wrong, it's
  just narrow.

## TODO

1. Uncomment the TODO block.
2. Run `go run . --namespace default`.
3. In another terminal: `kubectl run pinger --image=busybox --restart=Never -- sh -c 'sleep 3'`. Watch the events fly past.
4. Wrap the body in a `for { ... }` outer loop that reopens the watch when
   the channel closes — that's the first step toward an informer. Compare
   your wrapper to what example 05 gives you for free.
