# Capstone — `gcs-log-shipper`

> Combines: **02** (files/os — tailing + atomic offset writes), **05**
> (concurrency — tailer-fan-in + batcher + uploader pipeline), **07-gcp**
> (GCS `Writer`/`Close` + retry on transient errors).

The "Fluentbit lite" pattern, GCS edition. Sibling of [`s3-log-shipper`](../s3-log-shipper/):
same problem shape, GCS-specific wrinkles instead of S3 ones.

Tails one or more local log files; batches lines by size **or** time; gzips
each batch; uploads to `gs://<bucket>/<key-prefix>/<hostname>/<unixnano>.gz`;
persists each file's offset so a restart doesn't re-ship the world.

## Spec

- Pipeline: `Tailer(s) ──linesCh──▶ Batcher ──Batch──▶ Uploader ──▶ GCS`
- Each `Tailer` reads from a persisted offset; persists the new offset after
  every successful emit; survives file truncation (offset > size → reset)
  and EOF-then-append (read till EOF, sleep `PollInterval`, read again).
- `Batcher` flushes on **size** (`MaxBytes` of raw lines) OR **age**
  (`MaxAge` since the first line of the current batch).
- `GCSUploader` retries transient GCS errors (gRPC `Unavailable`,
  `DeadlineExceeded`, transport errors) with exponential backoff + jitter;
  permanent errors (e.g. `PermissionDenied`, missing bucket) fail-fast.
- `Batch.CRC32C` is the **Castagnoli** CRC of the gzipped body — GCS uses
  Castagnoli, not the IEEE default, so this CRC is comparable to the one
  GCS reports back. This is the load-bearing GCS quirk from section 07-gcp;
  there's a regression test that pins it.
- Flags: `--path` (repeatable), `--bucket`, `--key-prefix`,
  `--max-batch-bytes`, `--max-batch-age`, `--offset-dir`, `--max-retries`.

## Why GCS, not S3, as a second capstone

The two cloud SDKs are deliberately different on purpose:

| Thing                | S3 (SDK v2)                         | GCS (`cloud.google.com/go/storage`)             |
|---|---|---|
| Commit semantics     | `PutObject` returns when done       | `Writer.Close()` is what commits — defer hides errors |
| Retry hooks          | `aws.Retryer` middleware            | client has retries; explicit retry needed for visibility |
| Object integrity     | ETag (md5 of single-part)           | **CRC32C with Castagnoli polynomial** (not IEEE) |
| Transport            | HTTP/1.1 + JSON or REST             | gRPC by default                                  |
| Mockability          | Interface that matches `*s3.Client` | Chained concrete types — needs a thin adapter    |

The exercise reuses the same pipeline structure so the diff between this and
`s3-log-shipper` highlights *only* the cloud-specific decisions.

## Files

| File           | Purpose                                                                                  |
|---|---|
| `main.go`      | cobra entry; wires the real `*storage.Client` + adapter + calls `Run`. Full impl.        |
| `tail.go`      | `Tailer` + `OffsetStore` + `FileOffsetStore` — the file-tailing state machine.           |
| `batch.go`     | `Batcher` — size/age flush triggers; gzip; Castagnoli CRC.                               |
| `upload.go`    | `Uploader` interface + `GCSUploader` (retry/backoff) + `gcsClientAdapter` + `IsTransient`. |
| `run.go`       | `Run` — fan-in tailers, drive batcher, hand batches to uploader. The orchestrator.        |
| `main_test.go` | `//go:build exercise` — pins the whole contract end-to-end.                              |

## What the tests verify

| Test                                          | Concept                                          |
|---|---|
| `TestTailer_ReadsFromOffset`                  | resume reads from stored offset                  |
| `TestTailer_PersistsOffsetOnEmit`             | offset advances after a successful emit          |
| `TestTailer_HandlesTruncation`                | offset > size → reset to 0                       |
| `TestTailer_HandlesEOFThenAppend`             | sleeps at EOF, then picks up new bytes           |
| `TestTailer_HoldsPartialLineAtEOF`            | a line without trailing `\n` is buffered, not lost |
| `TestFileOffsetStore_RoundTrip`               | Load returns what Save wrote (atomic)            |
| `TestBatcher_SizeTrigger`                     | flush when raw bytes ≥ MaxBytes                  |
| `TestBatcher_AgeTrigger`                      | flush when first-line age ≥ MaxAge               |
| `TestBatcher_SizeBeatsTime`                   | size trigger wins if both conditions hit at once |
| `TestBatcher_EmptyFlushIsNoop`                | `Flush()` on empty state returns `(_, false)`    |
| `TestBatcher_GzipRoundTrip`                   | `Batch.Body` decompresses back to the input      |
| `TestBatcher_CRCIsCastagnoli`                 | regression: CRC uses Castagnoli, not IEEE        |
| `TestGCSUploader_TransientThenSuccess`        | 2× transient then OK → 3 attempts, no error      |
| `TestGCSUploader_PermanentNoRetry`            | permanent error → 1 attempt + error returned     |
| `TestGCSUploader_BackoffBounded`              | retries respect `MaxRetries`                     |
| `TestIsTransient_Classification`              | gRPC codes mapped correctly                      |
| `TestRun_EndToEnd_NoLoss`                     | N input lines arrive as ≥1 uploaded batches, all there |

All tests run against the local filesystem (real `os.File`) and a fake
`putter` — no real GCS, no creds.

## How to run (once you've implemented it)

```bash
# point it at a log file; dry-run by passing an invalid bucket and inspecting
# stderr for the retry/error path — or use a real bucket.
go run ./projects/gcs-log-shipper \
  --path /var/log/myapp.log \
  --bucket my-log-archive \
  --key-prefix logs/myapp \
  --max-batch-bytes 1048576 \
  --max-batch-age 30s \
  --offset-dir /var/lib/gcs-log-shipper \
  --max-retries 3
```

## How to run the exercise tests

```bash
go test -tags=exercise ./projects/gcs-log-shipper/...
```

Default `go test ./...` does **not** include these — they're gated behind
the `exercise` build tag, same as every other mini-project in the repo.
