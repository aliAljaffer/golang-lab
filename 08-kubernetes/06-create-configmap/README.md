# 06 — create a ConfigMap

`Create` is the universal write shape. Symmetric to Get and List:

```go
clientset.CoreV1().ConfigMaps(ns).Create(ctx, obj, metav1.CreateOptions{})
```

You build `obj` as a regular Go struct.

## Structuring a resource object

`corev1.ConfigMap` embeds `TypeMeta` and `ObjectMeta`:

```go
&corev1.ConfigMap{
    ObjectMeta: metav1.ObjectMeta{
        Name:      "demo-config",
        Namespace: "default",
        Labels:    map[string]string{"managed-by": "..."},
    },
    Data: map[string]string{"key": "value"},
}
```

Fields the server fills in for you (don't set them):

- `ResourceVersion` — the server's version stamp; used for optimistic concurrency
- `UID` — server-generated unique ID
- `CreationTimestamp`
- `SelfLink`, `Generation`

If you set them they're ignored on Create; on Update they're how you say
"I read version X; reject me if it's been changed since."

## CreateOptions

`metav1.CreateOptions{}` is usually empty. Useful fields:

| Field            | When you'd set it                                          |
|------------------|-------------------------------------------------------------|
| `DryRun`         | `[]string{metav1.DryRunAll}` — validate without persisting  |
| `FieldManager`   | Identify your tool for server-side apply attribution        |
| `FieldValidation`| `"Strict"` to reject unknown fields                         |

## Error: AlreadyExists

Creating a ConfigMap that already exists is a 409. `apierrors.IsAlreadyExists`
catches it. Idiomatic "upsert":

```go
existing, err := client.Get(ctx, name, metav1.GetOptions{})
if apierrors.IsNotFound(err) {
    _, err = client.Create(ctx, cm, metav1.CreateOptions{})
    return err
}
if err != nil { return err }
cm.ResourceVersion = existing.ResourceVersion  // optimistic concurrency
_, err = client.Update(ctx, cm, metav1.UpdateOptions{})
return err
```

Production code uses **server-side apply** instead: `client.Patch(ctx, name,
types.ApplyPatchType, data, metav1.PatchOptions{FieldManager: "my-tool"})`.
Same idempotent shape, no read-modify-write race. Out of scope for this
example, but worth knowing.

## Compare to other clients

|                       | Go (`client-go`)                       | Python (`kubernetes`)                |
|-----------------------|----------------------------------------|--------------------------------------|
| Build the object      | `&corev1.ConfigMap{...}`               | `client.V1ConfigMap(...)`            |
| Create call           | `ConfigMaps(ns).Create(ctx, cm, opts)` | `create_namespaced_config_map(ns, cm)` |
| AlreadyExists check   | `apierrors.IsAlreadyExists(err)`       | `e.status == 409`                     |

## TODO

1. Uncomment the TODO block.
2. Run `go run . --namespace default --name demo-config`. Verify with
   `kubectl get configmap demo-config -o yaml`.
3. Run it twice; observe the AlreadyExists path.
4. Add a `--update` flag that does the get-set-ResourceVersion-Update dance
   when the ConfigMap exists.
