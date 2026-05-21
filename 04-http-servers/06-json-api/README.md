# 06 — JSON API

POST endpoint with a JSON body. Four things to get right.

## 1. Cap the body size

```go
r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MiB
```

A malicious client can stream gigabytes into your decoder otherwise. `MaxBytesReader` short-circuits to an error once the limit is hit; the matching error response will be a 413.

## 2. Reject unknown fields

```go
dec := json.NewDecoder(r.Body)
dec.DisallowUnknownFields()
```

By default `encoding/json` silently ignores unknown fields. That hides typos on the client (`{"emial":"..."}` becomes a User with no email). Strict mode catches them.

## 3. Return JSON errors, not text

A JSON API should fail in JSON:

```json
{ "error": "name is required" }
```

Not `name is required\n` with `Content-Type: text/plain`. Consistency matters when the client is parsing responses.

## 4. Validate explicitly

Decoding only checks structure. After Decode, check business rules: required fields, format, ranges. Return 400 with a descriptive message.

## Comparison

| Concept          | Go                                   | Express                                   | FastAPI                        |
| ---------------- | ------------------------------------ | ----------------------------------------- | ------------------------------ |
| Decode JSON body | `json.NewDecoder(r.Body).Decode(&v)` | `app.use(express.json())` then `req.body` | typed via Pydantic params      |
| Strict mode      | `DisallowUnknownFields()`            | `strict: true` on body-parser             | Pydantic `extra="forbid"`      |
| Body size limit  | `http.MaxBytesReader`                | `express.json({limit: '1mb'})`            | uvicorn `--limit-request-line` |
| Auto validation  | hand-rolled                          | hand-rolled                               | Pydantic models                |

Go's stdlib is deliberately low-level. For larger projects with lots of endpoints, libraries like `go-playground/validator` add tag-based validation.

## Run

```bash
go run .
curl -i -X POST http://localhost:8080/users -H 'Content-Type: application/json' \
     -d '{"name":"ali","email":"ali@example.com"}'
curl -i -X POST http://localhost:8080/users -d '{"name":""}'
curl -i -X POST http://localhost:8080/users -d '{"username":"ali"}'   # unknown field
```
