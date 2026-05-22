# Exercise 03 — provider-error-handling

Translate raw `os` errors into actionable provider diagnostics.

## What you write

`WriteFileDiagnostics(path string, err error) diag.Diagnostics` — see `diagnostics.go` for the contract.

## Tests

| Test                                                  | What it pins                                                  |
|-------------------------------------------------------|---------------------------------------------------------------|
| `TestWriteFileDiagnostics_NilErrReturnsEmpty`         | nil err → no diags                                            |
| `TestWriteFileDiagnostics_PermissionDeniedHintsChmod` | mentions chmod/chown/privileges + the path                    |
| `TestWriteFileDiagnostics_MissingParentHintsCreateDir`| says "create the dir" / mkdir                                 |
| `TestWriteFileDiagnostics_GenericErrorIncludesUnderlying` | passes raw err string + path through                      |
| `TestWriteFileDiagnostics_AllSeveritiesAreError`      | nothing degraded to a Warning                                 |

Run:

```bash
go test -tags=exercise ./11-iac-tooling/exercises/03-provider-error-handling/
```

## Why this matters

Terraform diagnostics show up to the user like this:

```
╷
│ Error: <summary here>
│
│   with fileops_templated_file.config,
│   on main.tf line 3, in resource "fileops_templated_file" "config":
│    3: resource "fileops_templated_file" "config" {
│
│ <detail here>
╵
```

A summary like `write failed` and a detail like `open /etc/myapp/config.yaml: permission denied` is technically accurate but leaves the user to figure out what to do. A diagnostic with `Summary: "permission denied"` and `Detail: "cannot write to /etc/myapp/config.yaml: check file mode (chmod) and ownership (chown), or run terraform with sufficient privileges."` is the same information shaped for action.

## The framework `diag` package

| Constructor              | When                                                  |
|--------------------------|-------------------------------------------------------|
| `diags.AddError(s, d)`   | the operation failed; the user must act               |
| `diags.AddWarning(s, d)` | something looked off; the operation continued anyway  |
| `diags.AddAttributeError(path, s, d)` | error is about a specific attribute (`path` is a path.Path) |

Almost all diagnostics in a provider are `AddError`. Warnings exist for things like "deprecated attribute X is still set" or "we silently coerced your type." For an exercise that's specifically about *failed writes*, anything other than `SeverityError` is wrong — the dedicated severity test pins that.

## Tip: attribute-scoped errors

When the error is "about" a specific attribute (bad regex, invalid value), use `AddAttributeError(path, summary, detail)` with a `path.Root("attribute_name")`. Terraform highlights that attribute in the source `.tf` file output. The exercise's "write failed" errors aren't attribute-scoped (they're about the action, not a single attribute), so plain `AddError` is correct.
