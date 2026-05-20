# Plan: 11-iac-tooling

## Concepts to cover

### Terraform provider (intro)
- [ ] What a provider *is* (a grpc-served Go binary)
- [ ] `terraform-plugin-framework` (modern) vs `terraform-plugin-sdk/v2` (legacy)
- [ ] Resource lifecycle: Create, Read, Update, Delete (CRUD)
- [ ] Schema definition
- [ ] State vs plan
- [ ] How Terraform invokes the provider (`terraform-plugin-go` debug mode)

### Pulumi-Go
- [ ] Pulumi's model: code-as-infra vs Terraform's declarative DSL
- [ ] `pulumi.RunFunc`
- [ ] Reading config, defining resources
- [ ] When Pulumi-Go is better than Terraform (loops, conditionals, library reuse)

## Examples to build

| Folder | Demonstrates |
|---|---|
| `01-tf-provider-skeleton/` | Minimal provider with one resource (`echo`) |
| `02-tf-provider-crud/` | Add Create/Read/Update/Delete to the `echo` resource |
| `03-pulumi-hello/` | Smallest Pulumi-Go program (creates a local file resource) |

## Mini-project

**`tf-provider-fileops`** — a custom Terraform provider that manages local filesystem operations as Terraform resources (a `file_with_template` resource that templates a file based on Terraform variables). Includes acceptance tests.

Tests verify:
- Provider passes `terraform-plugin-framework` validation
- Acceptance tests (run with `TF_ACC=1`) create/read/update/destroy successfully

## Exercises

1. **`01-tf-data-source`** — add a data source (read-only) to the provider
2. **`02-pulumi-loop`** — use a Pulumi-Go loop to create N resources from a config list
3. **`03-provider-error-handling`** — make the provider gracefully handle missing files (return useful diagnostics)

## Status

- [ ] Concepts in README walkthrough
- [ ] Examples 01-03 built
- [ ] Mini-project `tf-provider-fileops` built + tested
- [ ] Exercises scaffolded
