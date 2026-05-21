# 02 — JSON Decode

Decode a JSON response into a typed struct.

## Things to notice

- **Struct tags** like `` `json:"full_name"` `` map Go field names to JSON keys. Without them, Go would look for a key called `FullName` and find nothing.
- `omitempty` on the tag means "skip this field when marshaling if it's the zero value." It has no effect on unmarshaling.
- `json.NewDecoder(r).Decode(&v)` streams directly from the response body — preferred over `json.Unmarshal(io.ReadAll(...))` because it avoids buffering the whole response.
- Unknown fields are silently dropped. Use `decoder.DisallowUnknownFields()` if you want to catch typos in your struct tags.

## Comparison

| Concept            | Go                              | Python                               | TS                         |
| ------------------ | ------------------------------- | ------------------------------------ | -------------------------- |
| Parse JSON to type | `json.NewDecoder(r).Decode(&v)` | `pydantic.parse_obj_as(T, r.json())` | `JSON.parse(text) as T`    |
| Field rename       | `` `json:"name"` ``             | `Field(alias="name")`                | n/a (or class-transformer) |
| Optional field     | `omitempty` (marshal-only)      | `Optional[...]`                      | `?`                        |
| Unknown fields     | silently dropped                | strict by default in pydantic        | n/a                        |

## Run

```bash
go run .
```
