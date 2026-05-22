# Exercise 03 — rolling restart

Trigger a rolling restart of a Deployment by patching its pod template's
`kubectl.kubernetes.io/restartedAt` annotation with the current timestamp.

This is exactly what `kubectl rollout restart deployment/<name>` does
under the hood. By using the same annotation key, your tool's restarts are
visible to `kubectl rollout status` and `kubectl rollout history` like any
other restart.

## Why this works

Kubernetes controllers reconcile based on `.spec.template`. Any change to
the pod template — even a label or annotation deep inside it — bumps the
Deployment's generation, which the Deployment controller treats as "the
template changed" and rolls out new pods.

The annotation is on `spec.template.metadata.annotations`, not on the
Deployment's top-level metadata. Putting it on the top level would NOT
cause a roll — that's pod template territory.

## What you implement

```go
type DeploymentAPI interface {
    Patch(ctx, name string, pt types.PatchType, data []byte,
          opts metav1.PatchOptions, subresources ...string) (*appsv1.Deployment, error)
}

func RollingRestart(ctx, api DeploymentAPI, ns, name string, now time.Time) error
```

`now` is injected so tests can pin a known timestamp. In production code,
callers pass `time.Now()`.

## Patch shape

Use `types.StrategicMergePatchType`. The body is a tiny JSON document:

```json
{ "spec": { "template": { "metadata": { "annotations":
  { "kubectl.kubernetes.io/restartedAt": "2026-05-22T12:30:45Z" }
} } } }
```

The strategic merge patch tells the API server "merge these fields in";
fields you don't mention are left alone.

## Contract pinned by tests

- `Patch` is called exactly once.
- Patch type is `StrategicMergePatchType`.
- The `name` argument matches what you passed in.
- The body parses as JSON and the `restartedAt` annotation equals `now.Format(time.RFC3339)`.
- Errors propagate (`errors.Is` predicate).
- Two calls with different `now` produce different bodies.

## Run the failing suite

```bash
go test -tags=exercise ./08-kubernetes/exercises/03-rolling-restart/
```
