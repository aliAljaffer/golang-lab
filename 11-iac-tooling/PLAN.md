# Plan: 11-iac-tooling

## Concepts to cover

### Terraform provider (intro)
- [x] What a provider *is* (a grpc-served Go binary)
- [x] `terraform-plugin-framework` (modern) vs `terraform-plugin-sdk/v2` (legacy)
- [x] Resource lifecycle: Create, Read, Update, Delete (CRUD)
- [x] Schema definition
- [x] State vs plan
- [x] How Terraform invokes the provider (`providerserver.Serve`)

### Pulumi-Go

Pulumi content is intentionally skipped — Terraform providers are the focus.

## Examples to build

| Folder | Demonstrates |
|---|---|
| `01-tf-provider-skeleton/` | Minimal provider with one resource (`echo_value`) |
| `02-tf-provider-crud/` | Full CRUD + validator + plan modifier on the `echo_value` resource |

## Mini-project

**`tf-provider-fileops`** — a custom Terraform provider with a `fileops_templated_file` resource that renders a Go template + vars into a local file. Drift detection via `Read` + idempotent destroy.

Tests verify (against pure helpers — `RenderTemplate`, `WriteTemplatedFile`, `ReadTemplatedFile`, `DeleteTemplatedFile`):
- Template rendering uses `missingkey=error` (no silent `<no value>` substitution)
- File round-trip on disk
- `os.ErrNotExist` propagates for drift detection
- Delete is idempotent (re-runnable)

Includes a `TestAcc*` stub that skips without `TF_ACC=1` — extension exercise: wire `terraform-plugin-testing` to drive a real `terraform apply` cycle.

## Exercises

1. **`01-tf-data-source`** — add a data source (read-only) — `fileops_file_info` reporting `exists` + `size`
2. **`03-provider-error-handling`** — translate raw `os` errors into actionable provider diagnostics (chmod/mkdir hints)

## Status

- [x] Concepts in README walkthrough
- [x] Examples 01-02 built
- [x] Mini-project `tf-provider-fileops` built + tested
- [x] Exercises scaffolded

## Session Log

When a Claude session does work in this section, append an entry to the root [`SESSIONS.md`](../SESSIONS.md) before ending — do **not** log session history in this file. `PLAN.md` is the plan; `SESSIONS.md` is the history. Tick the Status boxes above as items complete.
