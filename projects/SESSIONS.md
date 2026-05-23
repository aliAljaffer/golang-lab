# Projects — Session Log

History of Claude sessions working under `projects/`. The plan lives in
[`PLAN.md`](./PLAN.md); this file is the running log of what each session
actually did. One short bullet per session.

> Scope: only entries touching files under `projects/` belong here. Repo-wide
> work goes in the root `SESSIONS.md`.

## 2026-05-22 — scaffolded `kube-events-to-slack`

- Created `projects/kube-events-to-slack/` with `README.md`, `main.go`,
  `filter.go`, `dedup.go`, `sink.go`, `format.go`, `run.go`, and an
  `//go:build exercise`-tagged `main_test.go`.
- All function bodies are `TODO` for the student; default `go build ./...`
  and `go test ./...` stay green; `go test -tags=exercise ./projects/kube-events-to-slack/...`
  fails as expected (the contract the student needs to make pass).
- Ticked `kube-events-to-slack scaffolded` in `PLAN.md` and updated the
  status cell in `projects/README.md`.

## 2026-05-23 — scaffolded `s3-log-shipper`

- Created `projects/s3-log-shipper/` with `README.md`, `main.go`, `tail.go`,
  `batch.go`, `upload.go`, `run.go`, and an `//go:build exercise`-tagged
  `main_test.go`. Mirrors the `gcs-log-shipper` shape so the diff between
  the two capstones surfaces only the S3-vs-GCS-specific decisions:
  `IsTransient` classifies via `smithy.APIError.ErrorCode()` instead of
  gRPC codes, and `Batch.MD5` (hex md5 of the gzipped body, S3-ETag parity)
  is the load-bearing regression test in place of the GCS Castagnoli CRC.
- All function bodies are `TODO`; default `go build ./...` and
  `go test ./...` stay green; `go test -tags=exercise ./projects/s3-log-shipper/...`
  fails on every TODO as designed.
- Ticked `s3-log-shipper scaffolded` in `PLAN.md` and flipped the status
  cell in `projects/README.md`.

## 2026-05-23 — audited all four capstone scaffolds

- Read every `.go` file and `README.md` under `projects/kube-events-to-slack/`,
  `projects/s3-log-shipper/`, `projects/gcs-log-shipper/`, and
  `projects/deploy-bot/`; cross-checked against the per-capstone
  "Scaffold intent" bullets in `projects/PLAN.md`. All declared files,
  types, interfaces, function signatures, and pinned test cases present;
  every function body still `TODO` (correct — student work).
- Confirmed `//go:build exercise` gate on all four `main_test.go` files,
  `go vet ./projects/...` clean, and `go test ./projects/...` reports
  "no test files" (default builds stay green; tests only run under
  `-tags=exercise`).
- Added a `Capstones (projects/)` at-a-glance status block to root
  `PLAN.md`. Did not touch the four `scaffolded` checkboxes in
  `projects/PLAN.md` (already correctly ticked).
- Removed four stray Mach-O binaries (9–47 MB each) at repo root —
  `deploy-bot`, `gcs-log-shipper`, `kube-events-to-slack`,
  `s3-log-shipper` — accidental `go build` outputs, untracked.

## 2026-05-23 — scaffolded `deploy-bot`

- Created `projects/deploy-bot/` with `README.md`, `main.go`, `github.go`,
  `download.go`, `build.go`, `runctr.go`, `health.go`, `run.go`, and an
  `//go:build exercise`-tagged `main_test.go`.
- Capstone shape: `Fetch → Download → Build → Run → HealthCheck`.
  Interfaces (`ReleaseFetcher`, `Downloader`, `Builder`, `Runner`,
  `HealthChecker`) defined at the consumption site; production adapters
  (`GHReleaseFetcher`, `HTTPDownloader`, `dockerBuildAdapter`,
  `dockerRunAdapter`, `HTTPHealthChecker`) live in the same package; tests
  wire hand-rolled fakes against the interfaces.
- Load-bearing tests pinned: GitHub asset selection (Dockerfile preferred,
  `.tar.gz` fallback, `ErrReleaseNotFound` on 404, `ErrNoSuitableAsset`
  when neither matches), `DockerSafeTag` charset/leading-char rules,
  `DockerRunner` propagating `RunOpts.RemoveOnExit` to the SDK as
  `autoRemove` (parallel to the s3-shipper retry quirk), and pipeline
  cleanup-on-health-failure honoring `KeepContainer`.
- All function bodies are `TODO`; default `go build ./...` and
  `go test ./...` stay green; `go test -tags=exercise ./projects/deploy-bot/...`
  fails on every TODO as designed.
- Ticked `deploy-bot scaffolded` in `PLAN.md` and flipped the status cell
  in `projects/README.md`.
