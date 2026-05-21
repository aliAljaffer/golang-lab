# Exercise 02 — `pagination`

Follow `Link: <next-url>; rel="next"` headers (RFC 5988) to paginate through an API.

## What to build

In `pagination.go`, implement:

```go
func ParseNextLink(linkHeader string) string
func FetchAll(client *http.Client, startURL string) ([][]byte, error)
```

## The Link header format

GitHub (and many others) return:

```
Link: <https://api.github.com/.../issues?page=2>; rel="next",
      <https://api.github.com/.../issues?page=10>; rel="last"
```

- Each comma-separated entry is one URL + one or more `; name="value"` params.
- We only care about the entry whose params include `rel="next"`.
- `<` and `>` are part of the format — strip them.

## Behaviour

- `ParseNextLink` returns `""` when there's no `rel="next"`.
- `FetchAll` GETs `startURL`, captures its body, looks at the `Link` header on the response, and loops.
- On the first non-2xx or transport error, return what you've got plus the error.

## Run

```
go test -tags=exercise ./03-http-clients/exercises/02-pagination/...
```

## Stretch

- Replace `[][]byte` with a callback `handle func(page []byte) error` so the caller can stream-process without buffering all pages.
- Decode JSON inline (`json.NewDecoder`) instead of reading raw bytes.
- Add a per-page count cap to bail out of malformed APIs that pretend to paginate forever.
