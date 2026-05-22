# 02 — list containers

`docker ps` and `docker ps -a` in Go. The single endpoint is
`cli.ContainerList(ctx, opts)`; the `All` bool flips between "running only"
and "everything the daemon knows about."

## The Names slice prefix

Every name in `container.Names` is prefixed with `/`. Historical: container
names live in the daemon's local "namespace" the same way DNS hostnames do, so
the canonical form is `/my-container`. The Docker CLI strips it before
display — your tool should too.

```go
name := strings.TrimPrefix(c.Names[0], "/")
```

## Filters — the `filters.Args` DSL

```go
filters.NewArgs(
    filters.Arg("status", "running"),
    filters.Arg("label",  "owner=platform"),
)
```

Same key/value pairs that `docker ps -f status=running -f label=owner=platform`
uses. Filter keys are endpoint-specific — `status` is valid on `ContainerList`
but not on `ImageList`. The daemon will return a clear error if you pass a key
it doesn't understand.

## What's on a `types.Container`?

| Field | Notes |
|---|---|
| `ID` | Full 64-char SHA. Slice `[:12]` for `docker ps`-style display. |
| `Names` | `[]string`, "/"-prefixed. |
| `Image` | The image *reference* used to start it — e.g. `nginx:latest`, not the resolved sha256. |
| `ImageID` | The resolved `sha256:...` digest. |
| `State` | `running`, `exited`, `created`, `paused`, `restarting`, `dead`. |
| `Status` | Human string: `Up 3 hours`, `Exited (0) 2 days ago`. |
| `Ports` | `[]types.Port` (host:container mappings). |
| `Labels` | `map[string]string`. |

If you need more — env, mounts, network details — `cli.ContainerInspect(ctx, id)`
returns the full record.

## TODO

1. Uncomment the TODO blocks.
2. `docker run -d --name pinger alpine sleep 3600` and re-run; observe.
3. Add a filter by `label=mylabel=x`, run with `docker run -l mylabel=x ...`.
4. Print `c.Ports` as `host:container/proto` strings.
