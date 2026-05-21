# 07 — Webhook receiver

Verify HMAC-SHA256 signatures the way GitHub / Slack / Stripe send them.

## The pattern

The sender computes:

```
sig = HMAC-SHA256(secret, raw_body)
```

…and sends it in a header like `X-Hub-Signature-256: sha256=<hex>`. The receiver re-computes from the raw body it received and compares — **in constant time**, because timing differences leak the secret.

```go
mac := hmac.New(sha256.New, secret)
mac.Write(body)
got := mac.Sum(nil)

want, _ := hex.DecodeString(strings.TrimPrefix(header, "sha256="))
ok := hmac.Equal(got, want) // constant time
```

## Rules

1. **Verify against the raw body, not the parsed JSON.** Re-marshalling produces different bytes (key order, whitespace). Read the body once into a `[]byte`, verify, then parse.
2. **Use `hmac.Equal`, not `bytes.Equal` or `==`.** Constant-time comparison.
3. **Cap the body size.** `http.MaxBytesReader` — a webhook receiver is exposed to the internet.
4. **Return 401 on bad signature, not 400.** It's an auth failure.

## Generate a signed request locally

```bash
SECRET=swordfish
BODY='{"event":"push"}'
SIG=$(printf '%s' "$BODY" | openssl dgst -sha256 -hmac "$SECRET" | sed 's/^.* //')
curl -i -X POST http://localhost:8080/webhook \
  -H "X-Hub-Signature-256: sha256=$SIG" \
  -H "Content-Type: application/json" \
  -d "$BODY"
```

A different `BODY` with the same `SIG` will be rejected.

## What real providers send

| Provider | Header | Algorithm |
|---|---|---|
| GitHub | `X-Hub-Signature-256: sha256=<hex>` | HMAC-SHA256 over raw body |
| Slack | `X-Slack-Signature: v0=<hex>` (also `X-Slack-Request-Timestamp`) | HMAC-SHA256 over `v0:<ts>:<body>` |
| Stripe | `Stripe-Signature: t=<ts>,v1=<hex>` | HMAC-SHA256 over `<ts>.<body>` |

The mini-project `webhook-runner` uses the GitHub flavor.

## Run

```
WEBHOOK_SECRET=swordfish go run .
# then use the curl above
```
