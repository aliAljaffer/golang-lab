# Plan: projects/

Capstones combining multiple sections. Each should be a non-trivial mini-tool that could plausibly exist in a real environment.

## Scaffolding contract

Each capstone follows the same shape as the per-section mini-projects:

- **Claude scaffolds:** package layout, types, interfaces (defined at the consumption site per the section 07 doctrine), struct skeletons, function signatures with TODO bodies, a runnable `main.go` entry, and a `*_test.go` suite that **pins the contract** (happy path + failure modes + the load-bearing edge cases). Each capstone gets its own subfolder under `projects/` with its own `README.md`.
- **Student implements:** every function body marked `TODO`. The capstone is "done" when `go test -tags=exercise ./projects/<name>/...` is green.
- **Test gating:** all capstone tests carry `//go:build exercise`. Default `go test ./...` stays green; capstone tests are opt-in via `-tags=exercise`, same as the per-section mini-projects and exercises.
- **Interfaces over concretes:** external deps (k8s client, S3 client, docker daemon, GitHub API, HTTP destinations) are reached via narrow interfaces defined in the capstone package so tests can fake them. No test should require a real cluster / daemon / cloud creds to run.
- **Fakes, not mocks:** prefer hand-rolled fakes (channel-backed for streams, in-memory for stores) over mock libraries. Same convention as section mini-projects.

## Projects

### 1. `kube-events-to-slack`

- **Combines:** 04 (HTTP servers — for the Slack webhook client), 05 (concurrency — informer goroutine + dedup), 08 (kubernetes — informer)
- **What:** watches k8s events via an informer; filters by severity (Normal/Warning) + namespace allow-list + age cutoff; posts formatted alerts to a Slack incoming-webhook URL; dedupes within a cooldown window so a CrashLoopBackOff doesn't spam.
- **Why:** real ops use case; ties together informer cache, ResourceEventHandlerFuncs, outbound HTTP with retries, and dedup state.

**Scaffold intent:**

- `main.go` — CLI entry (flags: `--namespace`, `--webhook-url`, `--severities`, `--cooldown`, `--dry-run`, `--kubeconfig`); wires real `*kubernetes.Clientset` + `WebhookSink` (or `StdoutSink` if `--dry-run`) and calls `Run`.
- `filter.go` — `Filter` struct (severities set, namespaces set, min-age cutoff) + `ShouldAlert(*corev1.Event) bool`. Pure; trivially testable.
- `dedup.go` — `Deduper` with injectable `Now func() time.Time` + per-key `lastSeen map[string]time.Time` (mutex-guarded). Same shape as `08-kubernetes/mini-project/crashloop-alert`'s deduper.
- `sink.go` — `Sink` interface (`Send(ctx, Alert) error`) + `StdoutSink` (JSON-per-line) + `WebhookSink` (POSTs JSON, treats non-2xx as error, has retry/backoff TODO).
- `format.go` — `FormatSlackMessage(*corev1.Event) Alert` — pins the message shape the tests expect (text + fields: namespace, reason, count, age).
- `run.go` — `Run(ctx, kubernetes.Interface, Filter, *Deduper, Sink) error`; sets up `informers.NewSharedInformerFactoryWithOptions` on events; `AddFunc`/`UpdateFunc` close over filter+deduper+sink.
- `main_test.go` (`//go:build exercise`) — `Filter` (severity match, namespace match, age cutoff, empty filter = pass-all), `Deduper` (first-pass, blocks within cooldown, releases after, per-key isolation), `WebhookSink` (POSTs JSON, errors on 5xx, retries on transient — uses `httptest.Server`), `FormatSlackMessage` (output shape pinned), end-to-end with `fake.NewSimpleClientset` (a Warning event fires exactly one Send; a Normal event is silent when severities=[Warning]).

### 2. `s3-log-shipper`

- **Combines:** 02 (files/os — tailing + atomic offset writes), 03 (http-clients — S3 SDK uses these), 05 (concurrency — tailer-goroutine-per-file → batcher → uploader pipeline), 07 (AWS — S3 PutObject + retry/jitter)
- **What:** tails one or more local log files; batches lines by size or time; gzips the batch; uploads to S3 under a key derived from hostname + timestamp; persists the file offset so a restart doesn't re-ship the world. The "Fluentbit lite" pattern.
- **Why:** teaches batching state machines, backpressure across goroutines via channels, exponential-backoff-with-jitter for S3 5xx, and durable offset tracking.

**Scaffold intent:**

- `main.go` — CLI entry (flags: `--paths` repeated, `--bucket`, `--key-prefix`, `--max-batch-bytes`, `--max-batch-age`, `--offset-dir`, `--region`); wires real `s3.Client` + filesystem offset store + calls `Run`.
- `tail.go` — `Tailer` reads a file from a stored offset, emits lines on a `<-chan Line`, persists offset on each successful send. `OffsetStore` interface (`Load(path) int64` / `Save(path, off int64) error`); default impl writes a sibling `<path>.offset` file via the section-02 atomic-write pattern. Handles file truncation (offset > size → reset) and EOF-then-append.
- `batch.go` — `Batcher` accumulates lines; flushes when `len >= maxBytes` OR `age >= maxAge` (clock injected via `Now func() time.Time`); `Flush()` returns gzipped `[]byte` + sentinel key suffix. Pure state machine.
- `upload.go` — `Uploader` interface (`Put(ctx, key string, body []byte) error`); `S3Uploader` impl with exponential backoff + jitter for 5xx / `RequestTimeout`; gives up on permanent errors (403, 404 bucket).
- `run.go` — `Run(ctx, []Tailer, *Batcher, Uploader) error`; pipeline = tailers fan-in → batcher channel → uploader; clean shutdown on ctx with one final flush.
- `main_test.go` (`//go:build exercise`) — `Tailer` (reads from offset, persists on send, handles truncation, handles EOF-then-append, handles partial-line-at-EOF), `Batcher` with fake clock (size trigger, time trigger, size beats time when both ready, flush of empty batch is no-op, gzip round-trip), `S3Uploader` retry semantics (2 transient 5xx then success → 3 calls; permanent 403 → 1 call + error returned; backoff bounded by max retries), end-to-end with in-memory tailer + capture uploader (asserts N input lines arrive as 1+ uploaded gzipped objects, no data loss across batch boundaries).

### 3. `gcs-log-shipper`

- **Combines:** 02 (files/os — tailing + atomic offset writes), 05 (concurrency — tailer-fan-in → batcher → uploader pipeline), 07-gcp (GCS `Writer`/`Close` + retry on transient errors)
- **What:** the GCS sibling of `s3-log-shipper`. Same pipeline shape; differences are GCS-specific — `Writer.Close()` is the commit, retries classify gRPC codes (Unavailable/DeadlineExceeded transient vs PermissionDenied/NotFound permanent), and `Batch.CRC32C` is **Castagnoli**, not IEEE.
- **Why:** keeps the cloud-portable skeleton (tail/batch/upload, durable offsets, bounded retries) and surfaces only the GCS-specific decisions in the diff vs. `s3-log-shipper`. Practices the load-bearing GCS quirk from section 07-gcp — CRC32C with Castagnoli — as a pinned regression test.

**Scaffold intent:**

- `main.go` — cobra entry (flags: `--path` repeated, `--bucket`, `--key-prefix`, `--max-batch-bytes`, `--max-batch-age`, `--offset-dir`, `--max-retries`); wires real `*storage.Client` + adapter + calls `Run`.
- `tail.go` — `Tailer` reads a file from a stored offset, emits `Line` on a `chan<- Line`, persists each post-emit offset via `OffsetStore`. `FileOffsetStore` default writes a sibling `<path>.offset` file with the section-02 atomic-write pattern. Handles file truncation (offset > size → reset), EOF-then-append, and partial-line-at-EOF.
- `batch.go` — `Batcher` accumulates lines; flushes when `rawBytes >= MaxBytes` OR `now - firstAt >= MaxAge`; clock injected via `Now`. `Batch.CRC32C` MUST be computed with `crc32.MakeTable(crc32.Castagnoli)` (the GCS server-side polynomial, NOT IEEE); a regression test pins this.
- `upload.go` — `Uploader` interface (`Put(ctx, key, body) error`); `GCSUploader` wraps a `putter` with retry/backoff + jitter; `gcsClientAdapter` is the production `putter` driving a real `*storage.Client` (sets `ContentEncoding=gzip`; relies on `Writer.Close()` to commit). `IsTransient` classifies gRPC codes (`Unavailable`/`DeadlineExceeded`/`ResourceExhausted` → transient; `PermissionDenied`/`NotFound`/`InvalidArgument` → permanent; unknown-shape errors → transient).
- `run.go` — `Run(ctx, []*Tailer, *Batcher, Uploader, keyPrefix string, errOut io.Writer) error`; fan-in tailers → batcher loop with size + age triggers → uploader; one final `Flush()` on ctx-cancel; upload errors logged to errOut, do not kill the pipeline.
- `main_test.go` (`//go:build exercise`) — `Tailer` (offset resume, persist-on-emit, truncation, EOF-then-append, partial-line-at-EOF), `FileOffsetStore` (round-trip + missing == 0), `Batcher` (size trigger, age trigger, size beats time, empty flush is no-op, gzip round-trip, **CRC is Castagnoli not IEEE**), `GCSUploader` (transient-then-success → N+1 calls; permanent → 1 call + error; retries bounded by MaxRetries), `IsTransient` truth table for gRPC codes, end-to-end with real files + capture uploader (N lines arrive across ≥2 uploaded batches, no data loss across batch boundaries).

### 4. `deploy-bot`

- **Combines:** 01 (CLI — cobra), 03 (http-clients — GitHub API), 09 (docker — pull/build/run/inspect)
- **What:** a CLI that takes a `owner/repo` + release tag, fetches the corresponding artifact (a Dockerfile or a tarball-with-Dockerfile) from the GitHub Releases API, builds a local Docker image from it, runs the container with configured ports/env, then probes a health endpoint until it returns 200 (or times out). Reports the container ID + health status.
- **Why:** ties together GitHub API auth + paging, Docker SDK ImageBuild + ContainerRun + ContainerLogs, and a real human-facing CLI with subcommands and exit codes.

**Scaffold intent:**

- `main.go` / `cmd/` — cobra entry with `deploy <owner/repo> <tag>` + `--dry-run` + `--keep-container` + `--health-path` + `--health-timeout` + `--port`. Wires real `*github.Client` + docker `*client.Client` + calls `Run`.
- `github.go` — `ReleaseFetcher` interface (`Fetch(ctx, owner, repo, tag) (artifactURL string, err error)`); `GHReleaseFetcher` impl uses the GitHub API, handles 404 (release not found), picks the right asset when multiple are attached (TODO selection rule: prefer Dockerfile, then `.tar.gz`).
- `download.go` — `Downloader` interface (`Download(ctx, url) (io.ReadCloser, error)`); honors GH redirects; handles auth header pass-through.
- `build.go` — `Builder` interface (`Build(ctx, contextTar []byte, tag string) (imageID string, err error)`); impl wraps `cli.ImageBuild`; tag computation from `owner/repo:tag` → docker-safe form.
- `runctr.go` — `Runner` interface (`Run(ctx, RunOpts) (containerID string, err error)`); impl wraps `ContainerCreate`+`ContainerStart`+`ContainerInspect`; honors `RemoveOnExit` flag.
- `health.go` — `HealthChecker.Probe(ctx, url string) error` — polls until 200 or ctx-deadline; backoff between probes.
- `run.go` — `Run(ctx, ReleaseFetcher, Downloader, Builder, Runner, HealthChecker, Opts) (Report, error)` — orchestrates the pipeline + cleans up the container on health failure unless `--keep-container`.
- `main_test.go` (`//go:build exercise`) — `GHReleaseFetcher` (404 → typed error, multi-asset preference, missing asset → error; uses `httptest.Server`), `Builder` tag computation, `Runner` honors `RemoveOnExit`, `HealthChecker` (succeeds after N probes, times out on never-200, propagates ctx cancel), end-to-end with fake fetcher+downloader+builder+runner+health (asserts the pipeline calls each interface in order, cleans up on health failure, returns the right Report).

## Status

- [x] `kube-events-to-slack` scaffolded
- [ ] `kube-events-to-slack` built (tests green)
- [x] `s3-log-shipper` scaffolded
- [ ] `s3-log-shipper` built (tests green)
- [x] `gcs-log-shipper` scaffolded
- [ ] `gcs-log-shipper` built (tests green)
- [x] `deploy-bot` scaffolded
- [ ] `deploy-bot` built (tests green)

## Session Log

When a Claude session does work on a project in this folder, append an entry to the root [`SESSIONS.md`](../SESSIONS.md) before ending — do **not** log session history in this file. `PLAN.md` is the plan; `SESSIONS.md` is the history. Tick the Status boxes above as items complete.
