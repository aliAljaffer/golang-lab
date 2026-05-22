# 02 — tf-provider-crud

A real, fully-functional `echo_value` resource. End-to-end: Create / Read / Update / Delete / Replace / Refresh.

## What's new vs. 01

- **A `length` computed attribute** the provider derives from `input` — first taste of "schema-computed, not user-set."
- **A validator** (`stringvalidator.LengthAtLeast(1)`) — the framework rejects empty input *during plan*, before Create runs.
- **A plan modifier** (`stringplanmodifier.RequiresReplace()`) on `input` — changing it triggers Destroy+Create instead of in-place Update.
- **A stable `id`** derived from input (sha1 prefix). Stable IDs mean `terraform refresh` is a noop when nothing changed.

## State, Plan, and Config — the three values

| Value  | Where it comes from         | When it's read       | When it's null/unknown          |
|--------|------------------------------|----------------------|----------------------------------|
| Config | the user's `.tf` file        | always               | optional/unset attributes        |
| Plan   | Terraform's diff engine      | Create/Update only   | computed-by-provider attributes  |
| State  | the `terraform.tfstate` file | Read/Update/Delete   | resource not yet created         |

The pattern in this example:

```go
// Create / Update
var data echoModel
req.Plan.Get(ctx, &data)         // Plan is the source of truth
populate(&data)                  // provider computes the computed fields
resp.State.Set(ctx, &data)       // State is the sink

// Read
var data echoModel
req.State.Get(ctx, &data)        // State is the source
// (in a real provider: call remote API, overwrite data)
resp.State.Set(ctx, &data)       // State is also the sink
```

## Plan modifiers — the "what kind of change is this" knob

| Modifier                                | What it does                                              |
|-----------------------------------------|-----------------------------------------------------------|
| `stringplanmodifier.RequiresReplace()`  | changing this attr → Destroy+Create                       |
| `stringplanmodifier.UseStateForUnknown()` | keep the prior state value through plans (no "(known after apply)") |
| `stringplanmodifier.RequiresReplaceIf(fn, ...)` | conditional replace based on Old vs New                |

Plan modifiers run **after** validators but **before** the CRUD handler. They shape what the diff *looks like* to the user.

## Validators — the "is this even valid" knob

Validators live on the attribute schema. The framework runs them during `terraform plan` — so the user sees a useful error message instead of the Create call panicking.

Common ones (from `terraform-plugin-framework-validators`):

```go
stringvalidator.LengthAtLeast(1)
stringvalidator.LengthBetween(1, 64)
stringvalidator.RegexMatches(regexp.MustCompile(`^[a-z]+$`), "lowercase only")
stringvalidator.OneOf("us-east-1", "us-west-2", "eu-west-1")
```

For numbers/bools/lists there are matching packages (`int64validator`, `boolvalidator`, `listvalidator`).

## Compare to writing a Pulumi provider

Pulumi providers do the same CRUD dance — Create/Diff/Update/Delete — but the diff engine lives in the provider, not in a separate `terraform plan`. So Pulumi providers do more work; Terraform providers are simpler because Terraform already computed the diff for you.

## Running it

```bash
go build -o terraform-provider-echo
BIN_PATH=~/.terraform.d/plugins/registry.terraform.io/examples/echo/0.2.0/$(go env GOOS)_$(go env GOARCH)
mkdir -p "$BIN_PATH"
mv terraform-provider-echo "$BIN_PATH/"
```

Then in a scratch dir:

```hcl
terraform {
  required_providers {
    echo = { source = "registry.terraform.io/examples/echo", version = "0.2.0" }
  }
}
resource "echo_value" "x" {
  input = "hello"
}
output "out"    { value = echo_value.x.output }
output "length" { value = echo_value.x.length }
output "id"     { value = echo_value.x.id }
```

```bash
terraform init
terraform apply              # observe Create
terraform apply              # noop — Read+plan see no drift
terraform apply -var ...     # change input via the .tf file → Replace
terraform destroy
```

## TODO

1. Uncomment the marked blocks in `Create` and `Update` (Read is already filled).
2. `go build ./...` — confirm it builds.
3. Try installing locally and running `terraform plan` against a `.tf` that uses `echo_value`.
4. Add a `terraform_data` lifecycle: set `input = ""` in the .tf and observe the validator firing during plan.
5. Read the mini-project for what a non-toy provider looks like (real filesystem operations, acceptance tests).
