# 05 — `httptest.NewServer`

The right answer for testing HTTP clients in Go: don't mock the client, mock the upstream.

## Things to notice

- `httptest.NewServer(handler)` boots a real server on a random localhost port. Listener, TLS-or-not, parser, connection reuse — all real.
- `srv.URL` is the random base URL (e.g. `http://127.0.0.1:54321`). Inject it into your client. The reason `Client.BaseURL` is a field on the struct, not a hardcoded const, is exactly so tests can do this.
- `srv.Client()` returns a `*http.Client` that knows how to talk to `srv` (matters for `NewTLSServer` — sets up cert trust).
- `defer srv.Close()` shuts down the listener at end of test.
- The handler is a real `http.Handler`. You write the canned response, the bad-status path, the slow path — all by writing handler code, not by stubbing methods on a fake client.

## Why this over mocking `http.Client`

- The real network stack runs. You catch URL escaping bugs, Content-Type negotiation bugs, body-not-closed leaks.
- The handler is plain Go. No DSL like `gock.New("...").Reply(200).JSON(...)`.
- Switching from real network to fake is a one-line URL change — the rest of the client code is identical in production and tests.

## When NOT to use httptest

- Pure transport-layer bugs (DNS, real TLS handshake). For those you need an integration test against a real environment.
- Code that does NOT make HTTP calls. Don't reach for httptest if a plain function call suffices.

## Run

```
go test ./06-testing/05-httptest/...
go test -v ./06-testing/05-httptest/...
```

## Related

`gh-repo-stats` (`03-http-clients/mini-project`) uses this same pattern across its whole test file — same `httptest.NewServer` recipe, just with more elaborate handler scripting (retry-after-N-failures, etc.).
