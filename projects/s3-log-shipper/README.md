# Capstone — `s3-log-shipper`

> Combines: **02** (files/os — tailing + atomic offset writes), **03**
> (http-clients — what the AWS SDK is built on), **05** (concurrency —
> tailer-fan-in + batcher + uploader pipeline), **07** (AWS — S3 `PutObject`
> + retry on transient errors with backoff/jitter).

The "Fluentbit lite" pattern, S3 edition. Sibling of [`gcs-log-shipper`](../gcs-log-shipper/):
same pipeline shape, S3-specific wrinkles instead of GCS ones.

Tails one or more local log files; batches lines by size **or** time; gzips
each batch; uploads to `s3://<bucket>/<key-prefix>/<hostname>/<unixnano>.gz`;
persists each file's offset so a restart doesn't re-ship the world.

## Spec

- Pipeline: `Tailer(s) ──linesCh──▶ Batcher ──Batch──▶ Uploader ──▶ S3`
- Each `Tailer` reads from a persisted offset; persists the new offset after
  every successful emit; survives file truncation (offset > size → reset)
  and EOF-then-append (read till EOF, sleep `PollInterval`, read again).
- `Batcher` flushes on **size** (`MaxBytes` of raw lines) OR **age**
  (`MaxAge` since the first line of the current batch).
- `S3Uploader` retries transient S3 errors (5xx, `RequestTimeout`, `SlowDown`,
  `Throttling`, transport errors) with exponential backoff + jitter;
  permanent errors (`AccessDenied`, `NoSuchBucket`) fail-fast.
- `Batch.MD5` is the hex md5 of the gzipped body — S3 reports the same md5
  back as the object's ETag for single-PUT uploads, so this is the only
  client-side hash that's comparable post-upload. (The S3 analogue of the
  Castagnoli quirk in `gcs-log-shipper`; pinned by a regression test.)
- Flags: `--path` (repeatable), `--bucket`, `--key-prefix`,
  `--max-batch-bytes`, `--max-batch-age`, `--offset-dir`, `--region`,
  `--max-retries`.

## Why S3 alongside GCS as a capstone

The two cloud SDKs are deliberately different on purpose:

| Thing                | S3 (SDK v2)                                | GCS (`cloud.google.com/go/storage`)             |
|---|---|---|
| Commit semantics     | `PutObject` returns when done              | `Writer.Close()` is what commits — defer hides errors |
| Retry hooks          | `aws.Retryer` middleware (built-in)        | client has retries; explicit retry needed for visibility |
| Error classification | `smithy.APIError.ErrorCode()` + HTTP status | gRPC `status.Code()`                            |
| Object integrity     | **ETag = md5** (single-part PUT)           | CRC32C with Castagnoli polynomial               |
| Transport            | HTTP/1.1 + REST                            | gRPC by default                                  |

The exercise reuses the same pipeline shape as `gcs-log-shipper` so the diff
between this and that capstone highlights *only* the cloud-specific
decisions — retry classification, integrity hash, commit semantics.

## Files

| File           | Purpose                                                                                       |
|---|---|
| `main.go`      | cobra entry; wires the real `*s3.Client` + adapter + calls `Run`. Full impl.                  |
| `tail.go`      | `Tailer` + `OffsetStore` + `FileOffsetStore` — the file-tailing state machine.                |
| `batch.go`     | `Batcher` — size/age flush triggers; gzip; hex md5 (S3 ETag parity).                          |
| `upload.go`    | `Uploader` interface + `S3Uploader` (retry/backoff) + `s3ClientAdapter` + `IsTransient`.       |
| `run.go`       | `Run` — fan-in tailers, drive batcher, hand batches to uploader. The orchestrator.            |
| `main_test.go` | `//go:build exercise` — pins the whole contract end-to-end.                                   |

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
| `TestBatcher_MD5MatchesGzippedBody`           | regression: MD5 equals the hex md5 of `Batch.Body` (S3 ETag parity) |
| `TestS3Uploader_TransientThenSuccess`         | 2× transient then OK → 3 attempts, no error      |
| `TestS3Uploader_PermanentNoRetry`             | permanent error → 1 attempt + error returned     |
| `TestS3Uploader_BackoffBounded`               | retries respect `MaxRetries`                     |
| `TestIsTransient_Classification`              | S3 error codes mapped correctly                  |
| `TestRun_EndToEnd_NoLoss`                     | N input lines arrive as ≥1 uploaded batches, all there |

All tests run against the local filesystem (real `os.File`) and a fake
`putter` — no real S3, no creds.

## How to run (once you've implemented it)

```bash
go run ./projects/s3-log-shipper \
  --path /var/log/myapp.log \
  --bucket my-log-archive \
  --key-prefix logs/myapp \
  --region us-east-1 \
  --max-batch-bytes 1048576 \
  --max-batch-age 30s \
  --offset-dir /var/lib/s3-log-shipper \
  --max-retries 3
```

## How to run the exercise tests

```bash
go test -tags=exercise ./projects/s3-log-shipper/...
```

Default `go test ./...` does **not** include these — they're gated behind
the `exercise` build tag, same as every other mini-project in the repo.
