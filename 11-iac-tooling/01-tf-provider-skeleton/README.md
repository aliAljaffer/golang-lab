# 01 — tf-provider-skeleton

The minimum-viable Terraform provider, written against `terraform-plugin-framework`.

## What's in the box

- A provider named `echo` (no provider-level config)
- One resource: `echo_value` with two attributes (`input` required, `output` computed)
- Create copies `input` → `output`; Read/Update/Delete are stubs (see example 02 for real CRUD)

## The shape of a provider

```
provider.Provider          // top-level — the binary registers this
├── Metadata               // provider name + version
├── Schema                 // provider-level config attributes
├── Configure              // called once, after Schema (use to build clients)
├── Resources              // []func() resource.Resource — factories
└── DataSources            // []func() datasource.DataSource — factories

resource.Resource          // one per resource type
├── Metadata               // resource type name (`echo_value`)
├── Schema                 // attributes the user can set / will read
├── Create                 // plan → state, write to remote
├── Read                   // remote → state (drift detection / refresh)
├── Update                 // plan → state, mutate remote
└── Delete                 // state → tombstone, delete remote
```

Each method takes a `<Verb>Request` and writes to a `*<Verb>Response`. The response carries a `Diagnostics` slice — that's how you surface warnings/errors back to Terraform.

## Why `tfsdk:"..."` tags?

The framework uses reflection to marshal between Terraform's typed value system (`types.String`, `types.Bool`, ...) and your Go struct. The tag tells the marshaller which schema attribute maps to which field. Same idea as `json:"..."` but for tfvalues.

## Why `types.String` and not `string`?

Terraform values have a third state besides "value" and "zero" — they can also be **unknown** (known-after-apply) or **null** (not set). A plain Go `string` can't represent those. `types.String` carries that ternary information.

```go
v.IsUnknown()  // true during plan, before the value is computed
v.IsNull()     // true if the attribute wasn't set
v.ValueString() // the underlying string (panic-free; "" if null/unknown)
```

## How the plugin talks to Terraform

`providerserver.Serve` starts a gRPC server on stdin/stdout (the plugin protocol). When you run `terraform apply`, the CLI exec()s your binary and connects. The `Address` field is the registry source — locally you install at `~/.terraform.d/plugins/<address>/<version>/<os>_<arch>/`.

## Compare to other ecosystems

|                  | Terraform provider (Go)     | Pulumi provider              | AWS CDK construct           |
|------------------|------------------------------|------------------------------|------------------------------|
| Plugin lives in  | grpc-served Go binary        | grpc-served per-lang binary  | in-process TypeScript/Java   |
| State            | tfstate JSON file            | Pulumi state backend         | CloudFormation (downstream)  |
| Lifecycle method | Create/Read/Update/Delete    | Create/Diff/Update/Delete    | none (just a struct)         |
| Extension lang   | Go (`-plugin-framework`)     | Go / TS / Python / .NET      | TS / Python / Java / Go      |

## TODO

1. Uncomment the schema + Create blocks. Notice: Update is already filled in — it mirrors Create.
2. `go build ./...` — confirm it compiles.
3. Read the framework docs on attribute validators (`stringvalidator.LengthAtLeast(1)`) — you'll add one in example 02.
4. Read example 02 to see what a fully filled-out CRUD looks like.

You won't actually run the binary in this example — example 02 includes the local-install + `terraform apply` walkthrough.
