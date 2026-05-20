# 03 — HTTP Clients

> Status: ☐ not started — see [`PLAN.md`](./PLAN.md)

## What you'll learn

- Making HTTP requests with `net/http`
- JSON marshaling/unmarshaling with struct tags
- **Timeouts** (the single most common Go networking bug is forgetting to set them)
- Custom transports, headers, retries with backoff
- Consuming a real API (GitHub)

## Mental model from other languages

| Concept | Go | Python | TS |
|---|---|---|---|
| HTTP GET | `http.Get` | `requests.get` | `fetch` / `axios.get` |
| JSON to struct | `json.Unmarshal` | `json.loads` + dataclass | `JSON.parse as T` |
| Struct tags | `` `json:"name"` `` field tag | `dataclass(field_name)` | `class-transformer` decorators |
| Timeouts | `http.Client{Timeout: ...}` | `requests.get(..., timeout=...)` | `AbortController` |
| Retries | manual / `cenkalti/backoff` | `tenacity` | `axios-retry` |

## The DevOps angle

Ops tools constantly call APIs — cloud SDKs underneath are HTTP clients, webhooks need to be sent, internal services need to be polled. Knowing the patterns (timeouts, retries with jitter, structured errors) is essential.

See [`PLAN.md`](./PLAN.md).
