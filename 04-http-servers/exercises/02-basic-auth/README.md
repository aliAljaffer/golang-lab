# Exercise 02 — Basic Auth middleware

HTTP Basic Auth with constant-time comparison.

## What to implement

`BasicAuth(username, password string) func(http.Handler) http.Handler`:

1. Use `r.BasicAuth()` — the stdlib decodes the `Authorization: Basic ...` header for you.
2. Compare both username and password with `subtle.ConstantTimeCompare` (returns 1 on match, 0 on mismatch).
3. On any mismatch, set `WWW-Authenticate: Basic realm="restricted"` and return 401.

## Why constant-time

`==` short-circuits the moment a byte differs. An attacker who can measure response times learns *where* the first mismatch is — and can guess one byte at a time. Constant-time compare always takes the same time regardless of input.

For passwords specifically, the real-world answer is "don't compare plaintext at all — store a bcrypt/argon2 hash and verify against that." For this exercise we're comparing in-memory constants from CLI flags, where constant-time string compare is the right tool.

## Run the tests

```
go test -tags=exercise ./04-http-servers/exercises/02-basic-auth/...
```

## Stretch

- Support multiple users (load `htpasswd`-style file at startup).
- Add bcrypt password hashing: store `$2a$...` hashes, verify with `golang.org/x/crypto/bcrypt`.
- Move to a route group: only `/admin/*` is protected; `/healthz` is public.
