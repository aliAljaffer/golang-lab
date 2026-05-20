# Plan: 03-http-clients

## Concepts to cover

- [ ] `http.Get` quick-and-dirty vs `http.Client{}` for production
- [ ] **Always set timeouts** — the default client has none, will hang forever
- [ ] Request/response lifecycle, `Body.Close()`, why `defer resp.Body.Close()` is on every example you've ever seen
- [ ] JSON: `json.Marshal`/`Unmarshal`, struct tags, `omitempty`, nested structs
- [ ] Custom headers (auth, user-agent)
- [ ] Custom `http.Transport` (connection pooling, TLS config)
- [ ] Retries with exponential backoff + jitter
- [ ] `context.Context` for cancellation
- [ ] Streaming responses (don't load whole body into memory)

## Examples to build

| Folder | Demonstrates |
|---|---|
| `01-basic-get/` | Simplest GET, read body, print |
| `02-json-decode/` | Hit a JSON API, decode into struct with tags |
| `03-timeouts/` | Two clients side by side: default vs timeout-set; show the bug |
| `04-headers-auth/` | Setting Authorization header, custom User-Agent |
| `05-retry-backoff/` | Retry on 5xx with exponential backoff |
| `06-context-cancel/` | Cancel a request mid-flight with `context.WithTimeout` |
| `07-stream-response/` | Process a large response without buffering it all |

## Mini-project

**`gh-repo-stats`** — fetches repo metadata for a list of GitHub repos (stars, forks, last commit date), caches responses to a local file to avoid rate limiting, outputs CSV.

Tests verify:
- Retries on 503/429
- Honors `If-None-Match` (etag) cache hits
- CSV output schema matches spec

## Exercises

1. **`01-url-health-check`** — given a list of URLs, return which are 2xx, with response times
2. **`02-pagination`** — paginate through a GitHub endpoint that returns Link headers
3. **`03-mock-server-tests`** — use `httptest.NewServer` to test your retry logic without a real network

## Status

- [ ] Concepts in README walkthrough
- [ ] Examples 01-07 built
- [ ] Mini-project `gh-repo-stats` built + tested
- [ ] Exercises scaffolded
