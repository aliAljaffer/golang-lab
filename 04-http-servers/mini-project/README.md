# `webhook-runner` — mini-project

A tiny CI runner. Receives signed webhooks, looks up a named job in a YAML config, runs the command, and replies with the exit code + (truncated) output.

## Spec

```bash
WEBHOOK_SECRET=swordfish webhook-runner -c webhooks.yaml --addr :8080
```

Config (`webhooks.yaml`):

```yaml
jobs:
  deploy:
    command: ["sh", "-c", "./scripts/deploy.sh prod"]
  smoke:
    command: ["./bin/smoke-test"]
```

Send a signed request:

```bash
SECRET=swordfish
BODY='{"job":"deploy"}'
SIG=$(printf '%s' "$BODY" | openssl dgst -sha256 -hmac "$SECRET" | sed 's/^.* //')
curl -i -X POST http://localhost:8080/webhook \
  -H "X-Hub-Signature-256: sha256=$SIG" \
  -H "Content-Type: application/json" \
  -d "$BODY"
```

Response:

```json
{ "job": "deploy", "exit_code": 0, "output": "..." }
```

## How it's split for testability

| Function                             | Job                                                                           |
| ------------------------------------ | ----------------------------------------------------------------------------- |
| `LoadConfig(path)`                   | Parse the YAML, validate non-empty jobs + non-empty commands.                 |
| `VerifyHMAC(secret, body, header)`   | Constant-time check of `sha256=<hex>`.                                        |
| `runJob(ctx, job, maxOutput)`        | `exec.CommandContext` + capture stdout/stderr + truncate + extract exit code. |
| `newHandler(cfg, secret, maxOutput)` | Wire everything into `POST /webhook`.                                         |
| `newRootCmd()`                       | cobra wiring + signal-driven graceful shutdown.                               |

Splitting `runJob` out of the handler is what makes the **graceful-shutdown-drains** test possible: the handler runs synchronously, so `srv.Shutdown(ctx)` blocks until the in-flight job returns. No background goroutines = no drain logic to write.

## Errors and status codes

| Condition                       | Status                                                                   |
| ------------------------------- | ------------------------------------------------------------------------ |
| Missing / bad signature         | `401`                                                                    |
| Unknown job name                | `404`                                                                    |
| Job ran but exited non-zero     | `200` with `exit_code != 0` (the _request_ succeeded — the _job_ failed) |
| Couldn't even spawn the process | `500`                                                                    |

## Run the tests

```bash
go test -tags=exercise ./04-http-servers/mini-project/...
```

All tests fail until you implement the TODOs. The graceful-shutdown test is timing-sensitive (it does a real `sleep 0.3`) — should be reliable on a normal laptop, may need bumping in a starved CI.

## Stretch ideas

- Per-job timeout in the YAML, applied to `runJob`'s context.
- Concurrent job slots with a semaphore — reject extras with `429`.
- Audit log: append each request's `{job, signature_ok, exit_code, ts}` to a JSONL file.
- Honor an `X-Hub-Signature` (the older SHA-1 variant) as a fallback to match real GitHub deliveries.
- Use `chi` + `middleware.Recoverer` so a panicking command doesn't kill the server.
