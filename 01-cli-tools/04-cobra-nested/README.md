# 04 — Cobra, nested subcommands

Real DevOps CLIs have a tree shape: `kubectl get pods`, `gh pr list`, `aws s3 cp`. Cobra calls every node a `Command`. Parents typically have **no** `Run` — they only group children and contribute persistent flags.

## Patterns

- **Persistent flags** on a parent become flags on every descendant. Use this for `--namespace`, `--region`, `--profile`.
- **`cobra.ExactArgs(1)`** / `cobra.MinimumNArgs(1)` validate positional args before `RunE` runs.
- Use `Use: "pod NAME"` so the auto-generated help shows the positional arg.

## Try it after implementing

```
go run . --help
go run . get --help
go run . get pods --namespace kube-system
go run . delete pod my-pod-123
```

## When you'll reach for this

Any time your CLI has more than one "verb" — start with cobra, not `flag`. Migrating later is annoying.
