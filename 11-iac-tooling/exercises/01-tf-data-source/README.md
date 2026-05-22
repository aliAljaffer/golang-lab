# Exercise 01 — tf-data-source

Add a read-only **data source** to the provider.

## The deal

```hcl
data "fileops_file_info" "passwd" {
  path = "/etc/passwd"
}

output "passwd_size" { value = data.fileops_file_info.passwd.size }
```

Data sources are how a Terraform module reads existing-world state without managing it. Common examples in the wild: `data "aws_caller_identity"`, `data "aws_ami"`, `data "kubernetes_namespace"`.

## What you write

1. `ReadFileInfo(path string) (FileInfo, error)` — pure helper. Missing file → `Exists=false`, NOT an error.
2. `FileInfoDS.Metadata` — `<provider>_file_info`.
3. `FileInfoDS.Schema` — three attributes (`path`, `exists`, `size`).
4. `FileInfoDS.Read` — Config-in, State-out, calling `ReadFileInfo` in the middle.

## Tests

| Test                                          | What it pins                                        |
|-----------------------------------------------|-----------------------------------------------------|
| `TestReadFileInfo_ExistingFile`               | happy path: stat + size                             |
| `TestReadFileInfo_MissingFileIsNotAnError`    | the contract — `Exists=false` is the return shape    |
| `TestReadFileInfo_EmptyFile`                  | zero-size files are real files, not ghosts          |
| `TestMetadata_TypeNameUsesProviderPrefix`     | type naming convention                              |
| `TestSchema_HasThreeAttributes`               | schema completeness                                 |

Run:

```bash
go test -tags=exercise ./11-iac-tooling/exercises/01-tf-data-source/
```

## Why missing-file-is-not-an-error

A data source's job is to **report** state, not gate it. If `ReadFileInfo` returned an error for a non-existent file, the user would have to wrap their entire module in `count = ...` shenanigans just to handle "we may or may not have this file." Returning `Exists=false` lets them write:

```hcl
resource "fileops_templated_file" "ensure" {
  count = data.fileops_file_info.config.exists ? 0 : 1
  path  = "/etc/myapp/config.yaml"
  # ...
}
```

Real-world parallel: `data "aws_caller_identity"` returns the caller's identity; it doesn't error if you're using a different profile. The data source is a lens, not a guard.

## Data source vs resource — the symmetry

| Method     | Resource             | Data source         |
|------------|----------------------|---------------------|
| Schema     | yes (write-side)     | yes (read-side)     |
| Create     | yes                  | -                   |
| Read       | yes (refresh)        | yes (the only verb) |
| Update     | yes                  | -                   |
| Delete     | yes                  | -                   |
| Import     | optional             | -                   |

A data source's Read handler reads from `req.Config` (not `req.State` — there's no persistent state for a data source; it re-reads every plan).
