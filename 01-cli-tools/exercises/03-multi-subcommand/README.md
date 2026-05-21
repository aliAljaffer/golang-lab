# Exercise 03 — multi-subcommand (kubectl-shape)

Build a small CLI shaped like `kubectl`:

```
tool get pods
tool get pod NAME
tool create pod NAME --image=nginx
tool delete pod NAME
```

## What's scaffolded

Two files split the work along a common Go boundary:

| File | What goes here |
|---|---|
| `store.go` | Pure logic: `Store`, `Pod`, `CreatePod` / `GetPod` / `ListPods` / `DeletePod`. **Tested.** |
| `cmd.go` | Cobra wiring. Builds the command tree, calls into `Store`. Not unit-tested. |

This split is the pattern: keep business logic in a plain package with no CLI imports; the CLI is a thin shell over it. You can swap cobra for a different framework, or expose the same logic as an HTTP API, without rewriting `store.go`.

## Run the tests

```
go test ./01-cli-tools/exercises/03-multi-subcommand/...
```

All six `TestStore_*` tests fail until you implement the methods.

## Stretch

- Add a `Node` resource and `tool get nodes` to exercise the multi-resource pattern.
- Add a `--namespace` persistent flag on `root` and partition the Store by namespace.
- Replace the in-memory map with JSON-backed storage (`./state.json`) — preview of section 02.
