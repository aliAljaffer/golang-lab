# 03 — HTTP Clients

> Status: ☑ scaffolded — examples + mini-project + exercises have failing tests. See [`PLAN.md`](./PLAN.md).

Almost every ops tool eventually calls an HTTP API: cloud SDKs are HTTP clients underneath, webhook senders are HTTP clients, status pages and dashboards poll HTTP endpoints. The patterns in this section — timeouts, retries with jitter, streaming responses, context cancellation — are the difference between a tool that works in the demo and a tool that survives in production.

The single most important rule: **never use `http.Get` in production code.** Read on for why.

---

## What you'll learn

- `http.Get` for quick-and-dirty calls vs `http.Client{}` for anything that ships
- **Always set timeouts** — the default `http.Client` has none; one misbehaving server will hang your tool forever
- The request/response lifecycle, why `defer resp.Body.Close()` is on every example you've ever seen
- JSON with `json.Marshal` / `json.Unmarshal`, struct tags, `omitempty`, nested decoders
- Custom headers (auth, user-agent) and custom `http.Transport` (connection pooling, TLS)
- Retries with exponential backoff + jitter
- `context.Context` for cancellation — the only sound way to abort an in-flight request
- Streaming responses without loading the whole body into memory

---

## Mental model from other languages

| Concept              | Go                              | Python                          | TS / Node                      |
| -------------------- | ------------------------------- | ------------------------------- | ------------------------------ |
| HTTP GET             | `http.Get` (toy) / `client.Do`  | `requests.get`                  | `fetch` / `axios.get`          |
| JSON to struct       | `json.Unmarshal` + struct tags  | `json.loads` + dataclass        | `JSON.parse as T`              |
| Struct field tags    | `` `json:"name"` ``             | `dataclass(field_name)`         | `class-transformer` decorators |
| Timeouts             | `http.Client{Timeout: ...}`     | `requests.get(..., timeout=...)` | `AbortController`              |
| Cancellation         | `context.WithCancel/Timeout`    | manual / `asyncio.CancelledError` | `AbortController`            |
| Retries              | manual / `cenkalti/backoff`     | `tenacity`                      | `axios-retry`                  |
| Mock server (tests)  | `httptest.NewServer`            | `responses` / `httpx_mock`      | `nock` / `msw`                 |

**The cultural difference:** Go's stdlib `net/http` is a complete, production-grade HTTP client. There's no `requests`-equivalent third-party library to reach for; people build retry layers, OAuth helpers, and circuit breakers as thin wrappers around `*http.Client`. The cost of being stdlib-first is that the defaults are unforgiving — `http.Get` has no timeout, no retries, no context, and you don't realize until your tool wedges in production.

---

## The DevOps angle

Ops tools constantly call APIs — cloud SDKs underneath are HTTP clients, webhooks need to be sent, internal services need to be polled. The patterns matter:

- **Timeouts are non-optional.** A hung tool is worse than a failing tool; CI catches the failing one.
- **Retry only idempotent verbs** by default (GET/HEAD/PUT/DELETE). POSTs need an idempotency key or you'll create-three-resources from one button click.
- **Jitter prevents thundering herd.** When a downstream recovers, 1,000 retry-immediately clients will knock it back over. Add `±20%` to your backoff delay.
- **Stream responses you don't need in memory.** A daily DB-dump pull that buffers 4 GB before parsing will OOM the container; `json.Decoder(resp.Body)` walks the stream.

---

## Walkthrough

Read these in order. Each is a runnable example demonstrating one specific concept.

1. [`01-basic-get/`](./01-basic-get/) — the simplest GET, read body, print. Shows what `http.Get` actually does (it's `DefaultClient.Get` under the hood) and why production code never uses it.
2. [`02-json-decode/`](./02-json-decode/) — hit a JSON API, decode into a struct with field tags. `json:"login,omitempty"` vs no tag; nested objects; ignoring unknown fields (the default — `DisallowUnknownFields()` if you want strictness).
3. [`03-timeouts/`](./03-timeouts/) — two clients side by side, default vs `http.Client{Timeout: 5*time.Second}`. The default-client-against-slow-server case is the headline bug.
4. [`04-headers-auth/`](./04-headers-auth/) — `req.Header.Set("Authorization", "Bearer ...")`, custom `User-Agent`. Built from `http.NewRequest` rather than the convenience `http.Get`.
5. [`05-retry-backoff/`](./05-retry-backoff/) — retry on 5xx + transient network errors, exponential backoff (`base * 2^attempt`) with `±20%` jitter. Cap at a max delay; cap at a max attempt count.
6. [`06-context-cancel/`](./06-context-cancel/) — `context.WithTimeout(ctx, ...)` attached via `req.WithContext(ctx)`. Cancelling the parent ctx tears down the in-flight TCP connection — essential for `Ctrl-C` to work mid-request.
7. [`07-stream-response/`](./07-stream-response/) — process a large response without buffering it all. `json.NewDecoder(resp.Body).Decode(&item)` in a loop for JSONL/ND-JSON streams; `io.Copy(dst, resp.Body)` for raw passthrough.

---

## Mini-project: [`gh-repo-stats`](./mini-project/)

Fetch repo metadata (stars, forks, last commit date) for a list of GitHub repos. Cache responses to a local file to avoid burning rate limit. Output CSV.

The point: a real client combines almost everything in this section — JSON decoding, custom headers (the `User-Agent` GitHub requires), retries on 5xx and 429, etag-based conditional GETs (`If-None-Match`), and a file-backed cache. Tests verify retries-on-503/429, etag cache hits skip the network, and the CSV output schema matches spec.

Spec and starter in [`mini-project/`](./mini-project/).

---

## Exercises

See [`exercises/`](./exercises/):

1. **[`01-url-health-check`](./exercises/01-url-health-check/)** — given a list of URLs, return which are 2xx, with response times. The natural place to learn `time.Since(start)` + per-request timeout.
2. **[`02-pagination`](./exercises/02-pagination/)** — paginate through a GitHub endpoint that returns `Link: <next>; rel="next"` headers. The pattern generalizes to every paginated API on the public internet.
3. **[`03-mock-server-tests`](./exercises/03-mock-server-tests/)** — use `httptest.NewServer` to test retry logic without a real network. Foreshadows the testing patterns in section 06.

---

## Further reading

- [`net/http` docs](https://pkg.go.dev/net/http) — the stdlib client/server reference
- [`encoding/json` docs](https://pkg.go.dev/encoding/json) — struct tags, `RawMessage`, `Decoder` vs `Unmarshal`
- [Cloudflare: Go's `net/http` timeouts](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/) — the canonical "all the timeouts you didn't know existed" post
- [AWS Architecture Blog: Exponential backoff and jitter](https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/) — the math behind the jitter recommendation
- [`context` docs](https://pkg.go.dev/context) — the cancellation API that underpins example 06
