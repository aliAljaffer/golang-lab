# Capstone — `deploy-bot`

> Combines: **01** (CLI — cobra), **03** (http-clients — GitHub Releases API),
> **09** (docker — pull/build/run/inspect).

A CLI that takes `owner/repo` + a release tag, fetches the artifact attached
to that release, builds a local Docker image from it, runs the container with
the configured port mapping, and probes a health endpoint until it returns
200 (or times out). Reports the container ID + health status; tears the
container down on failure unless `--keep-container` was passed.

## Spec

- Pipeline: `Fetch ──URL──▶ Download ──tar bytes──▶ Build ──imageID──▶ Run ──ctrID──▶ HealthCheck`
- `ReleaseFetcher` hits `GET https://api.github.com/repos/{owner}/{repo}/releases/tags/{tag}`
  and picks **one** asset using a deterministic preference order
  (Dockerfile → `.tar.gz` → error). A 404 from the API surfaces as a typed
  `ErrReleaseNotFound`.
- `Downloader` GETs the asset URL, following GitHub's 302 to the
  S3-presigned download URL. Strip the `Authorization` header on redirect —
  the S3 presigned URL signs its own auth and a stray `Authorization: token`
  header will collide.
- `Builder` accepts a tarball (`[]byte`) and an image tag; the tag is derived
  from `DockerSafeTag(owner, repo, tag)` — lowercase, slashes replaced,
  registry-safe.
- `Runner` honors `RunOpts.RemoveOnExit` → sets the docker `AutoRemove`
  flag on `ContainerCreate`. On health-check failure, `Run` calls
  `Runner.Remove(ctx, id)` unless the caller set `KeepContainer`.
- `HealthChecker` polls the configured URL until either a 2xx (success) or
  the ctx deadline fires (failure). Backoff between probes is fixed
  (`Interval`); ctx cancellation aborts the loop immediately.
- Flags: `deploy <owner/repo> <tag>` plus `--dry-run`, `--keep-container`,
  `--health-path`, `--health-timeout`, `--host-port`, `--container-port`,
  `--env` (repeatable), `--gh-token`.

## Files

| File           | Purpose                                                                                       |
|---|---|
| `main.go`      | cobra entry; wires the real `*client.Client` + `*http.Client` + calls `Run`. Full impl.       |
| `github.go`    | `ReleaseFetcher` interface + `GHReleaseFetcher` (uses `net/http` against the Releases API).   |
| `download.go`  | `Downloader` interface + `HTTPDownloader` (follows 302, strips auth on redirect).             |
| `build.go`     | `Builder` interface + `DockerBuilder` (wraps `cli.ImageBuild`) + `DockerSafeTag` helper.      |
| `runctr.go`    | `Runner` interface + `DockerRunner` (wraps `ContainerCreate`+`Start`+`Remove`).               |
| `health.go`    | `HealthChecker` interface + `HTTPHealthChecker` (poll-until-2xx-or-deadline).                 |
| `run.go`       | `Run` — orchestrates the pipeline + cleans up on health failure. Returns a `Report`.          |
| `main_test.go` | `//go:build exercise` — pins the whole contract end-to-end.                                   |

## What the tests verify

| Test                                            | Concept                                          |
|---|---|
| `TestGHReleaseFetcher_NotFound`                 | API 404 → typed `ErrReleaseNotFound`             |
| `TestGHReleaseFetcher_PrefersDockerfileAsset`   | multi-asset release: Dockerfile beats tarball    |
| `TestGHReleaseFetcher_FallsBackToTarball`       | no Dockerfile, `.tar.gz` present → tarball       |
| `TestGHReleaseFetcher_NoSuitableAsset`          | neither Dockerfile nor `.tar.gz` → error         |
| `TestGHReleaseFetcher_SendsToken`               | `Authorization: Bearer ${Token}` header sent     |
| `TestDockerSafeTag`                             | `owner/repo:tag` → lowercase, slashes replaced   |
| `TestDockerRunner_PassesAutoRemove`             | `RunOpts.RemoveOnExit` → SDK `AutoRemove=true`   |
| `TestHTTPHealthChecker_SucceedsAfterNProbes`    | retries until first 2xx                          |
| `TestHTTPHealthChecker_TimesOut`                | never-200 server + short ctx deadline → error    |
| `TestHTTPHealthChecker_PropagatesCtxCancel`     | cancel during sleep → return ctx.Err immediately |
| `TestRun_HappyPath`                             | fakes invoked in order, `Report.Healthy == true` |
| `TestRun_HealthFailureRemovesContainer`         | health err → `Runner.Remove` called once         |
| `TestRun_KeepContainerSkipsRemove`              | `KeepContainer=true` → no `Remove` on failure    |
| `TestRun_FetcherError_FailsFast`                | fetcher err → pipeline stops; later stages unhit |

All tests run against `httptest.Server` and hand-rolled fakes — no real
GitHub, no real Docker daemon, no creds.

## How to run (once you've implemented it)

```bash
go run ./projects/deploy-bot deploy alialjaffer/example v1.2.3 \
  --gh-token "$GH_TOKEN" \
  --host-port 8080 \
  --container-port 8080 \
  --health-path /healthz \
  --health-timeout 30s
```

## How to run the exercise tests

```bash
go test -tags=exercise ./projects/deploy-bot/...
```

Default `go test ./...` does **not** include these — they're gated behind
the `exercise` build tag, same as every other mini-project in the repo.
