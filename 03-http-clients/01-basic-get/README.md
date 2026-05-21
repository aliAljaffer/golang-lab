# 01 — Basic GET

The 30-second version of an HTTP request in Go.

## Things to notice

- `http.Get(url)` is a convenience wrapper around the package-level `http.DefaultClient`. It works for one-shot scripts; **do not use it in production code** because it has no timeout.
- The returned `*http.Response` owns a `Body` that you must close — `defer resp.Body.Close()` is on every example you have ever seen.
- A non-2xx response is **not** an error. `err` is only non-nil for transport failures (DNS, refused connection, mid-flight cancel). Always check `resp.StatusCode` yourself.
- `resp.Body` is an `io.ReadCloser` — anything that consumes an `io.Reader` (`json.NewDecoder`, `bufio.NewScanner`, `io.Copy`) will accept it.

## Comparison

| Concept | Go | Python | TS |
|---|---|---|---|
| GET request | `http.Get(url)` | `requests.get(url)` | `await fetch(url)` |
| Read body | `io.ReadAll(resp.Body)` | `resp.text` | `await resp.text()` |
| Close body | `defer resp.Body.Close()` | context-manager handles it | auto |
| Status code | `resp.StatusCode` | `resp.status_code` | `resp.status` |

## Run

```
go run . https://httpbin.org/get
```
