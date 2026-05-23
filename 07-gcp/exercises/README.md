# Exercises — 07-gcp

Each subfolder is an exercise with failing tests under
`//go:build exercise`. Default `go test ./...` stays green; opt in with
`go test -tags=exercise ./07-gcp/exercises/...`.

Tests use hand-rolled fakes — no exercise needs real GCP credentials. Tests
that would hit real GCP would be tagged `//go:build integration` (none of
these are).

See parent [`PLAN.md`](../PLAN.md).
