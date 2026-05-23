# Capstone — `kube-events-to-slack`

> Combines: **04** (HTTP servers — webhook client), **05** (concurrency — informer
> goroutine + dedup), **08** (kubernetes — informer + event API).

Watches Kubernetes `Event` objects via a SharedInformer; filters by severity
(Normal/Warning), a namespace allow-list, and an age cutoff; posts a formatted
alert to a Slack incoming-webhook URL; **dedupes** within a cooldown so a
`CrashLoopBackOff` event doesn't spam.

## Spec

- Builds a `kubernetes.Clientset` from in-cluster or kubeconfig.
- Builds an informer factory; registers an `Add` + `Update` handler on
  `corev1.Event`.
- The handler runs the pipeline: **filter → dedup → format → sink**.
- Sink is one of:
  - `StdoutSink` — JSON-per-line (default / `--dry-run`).
  - `WebhookSink` — POSTs JSON to a Slack incoming-webhook URL, treats non-2xx
    as an error, retries transient failures with backoff (TODO).
- Flags: `--namespace` (repeated; empty = all), `--webhook-url`, `--severities`
  (comma-separated; default `Warning`), `--cooldown` (default `5m`),
  `--max-age` (default `1h`; events older are skipped), `--dry-run`,
  `--kubeconfig`.

## Files

| File | Purpose |
|---|---|
| `main.go` | cobra entry; wires the real `*kubernetes.Clientset` + sink + calls `Run`. |
| `filter.go` | `Filter` struct + `ShouldAlert(*corev1.Event) bool` — pure. |
| `dedup.go` | `Deduper` with injectable clock + per-key `lastSeen`. |
| `sink.go` | `Sink` interface + `StdoutSink` + `WebhookSink`. |
| `format.go` | `FormatSlackMessage(*corev1.Event, now time.Time) Alert`. |
| `run.go` | `Run(ctx, kubernetes.Interface, Filter, *Deduper, Sink, errOut) error`. |
| `main_test.go` | `//go:build exercise` — pins the whole contract. |

## What the tests verify

| Test | Concept |
|---|---|
| `TestFilter_SeverityMatch` | only listed severities pass |
| `TestFilter_NamespaceAllowList` | only allow-listed namespaces pass |
| `TestFilter_AgeCutoff` | events older than `MaxAge` are skipped |
| `TestFilter_EmptyMeansPassAll` | zero-value filter passes everything |
| `TestDeduper_FirstAlertPasses` | initial state |
| `TestDeduper_BlocksWithinCooldown` | cooldown enforcement |
| `TestDeduper_AlertsAgainAfterCooldown` | window resets |
| `TestDeduper_PerKeyIsolation` | keys are independent |
| `TestStdoutSink_WritesLine` | stdout sink shape (JSON-per-line round-trip) |
| `TestWebhookSink_PostsJSON` | webhook sink shape |
| `TestWebhookSink_NonOKIsError` | non-2xx surfaces as error |
| `TestFormatSlackMessage_Shape` | Alert fields pinned from an Event |
| `TestRun_WarningFires_NormalIsSilent` | end-to-end w/ `fake.NewSimpleClientset` |

All tests run against pure functions or a fake clientset
(`k8s.io/client-go/kubernetes/fake`) — no real cluster needed.

## How to run (once you've implemented it)

```bash
# dry-run stdout mode
go run ./projects/kube-events-to-slack --dry-run --severities Warning

# webhook mode (Slack incoming-webhook URL)
go run ./projects/kube-events-to-slack \
  --webhook-url https://hooks.slack.com/services/T000/B000/XXXX \
  --namespace kube-system --namespace default \
  --severities Warning --cooldown 5m --max-age 1h
```

## How to run the exercise tests

```bash
go test -tags=exercise ./projects/kube-events-to-slack/...
```

Default `go test ./...` does **not** include these — they're gated behind
the `exercise` build tag, same as every other mini-project in the repo.
