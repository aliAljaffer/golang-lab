# 04 — Headers & Auth

Most real APIs need at least one custom header (auth) and benefit from a polite `User-Agent`.

## The pattern

```go
req, err := http.NewRequest(http.MethodGet, url, nil)
if err != nil { return err }
req.Header.Set("Authorization", "Bearer " + token)
req.Header.Set("User-Agent",   "myapp/1.0 (contact: x@y)")
req.Header.Set("Accept",       "application/json")
resp, err := client.Do(req)
```

`http.Get` / `http.Post` are shortcuts; the moment you need a header, switch to `NewRequest` + `client.Do`.

## Things to notice

- `Header.Set` replaces; `Header.Add` appends (relevant for multi-valued headers like `Accept-Encoding`).
- Canonical header names: `req.Header.Set("authorization", ...)` works — Go canonicalises to `Authorization` on the wire.
- **Never log a token.** When debugging request dumps, strip the `Authorization` value. `httputil.DumpRequest` will show it raw.
- The `User-Agent` defaults to `Go-http-client/1.1`, which many APIs treat as bot traffic. Set your own.

## Comparison

| Concept | Go | Python | TS |
|---|---|---|---|
| Build request | `http.NewRequest(...)` | `requests.Request(...)` | `new Request(...)` |
| Set header | `req.Header.Set(k, v)` | `headers={k: v}` | `req.headers.set(k, v)` |
| Bearer auth | manual | `requests.get(..., headers={"Authorization": "Bearer ..."})` | manual |

## Run

```
GH_TOKEN=fake go run .
```
