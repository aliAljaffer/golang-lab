# Exercises — 00-setup

Each subfolder is one exercise.

## Exercises

1. **[`01-tidy-experiment/`](./01-tidy-experiment/)** — CLI walkthrough: observe how `go mod tidy` modifies `go.mod` and `go.sum`
2. **[`02-static-binary/`](./02-static-binary/)** — CLI walkthrough: verify Go produces statically-linked binaries
3. **[`03-env-explorer/`](./03-env-explorer/)** — code exercise: implement `GoEnv(key)`; tests behind build tag `exercise`

## How code exercises work

Code exercises ship as **failing tests** behind the `exercise` build tag — so `go test ./...` stays green by default. To work on an exercise:

```bash
go test -tags=exercise ./00-setup/exercises/03-env-explorer/
```

The tests fail until you implement the function. When all tests pass, you're done.

See parent [`PLAN.md`](../PLAN.md) for what else is planned.
