# 11 — IaC Tooling

> Status: ☑ examples + mini-project + exercises scaffolded; walkthrough in this README — see [`PLAN.md`](./PLAN.md).

If you ever need Terraform to manage a resource it doesn't natively know about — an internal service's API, a SaaS without an existing provider, a custom config-store — you write a provider. Terraform providers are Go binaries that speak gRPC over stdin/stdout to the `terraform` CLI. The official toolkit is `github.com/hashicorp/terraform-plugin-framework` (v1+); the older `terraform-plugin-sdk/v2` still exists for legacy providers but new code should default to the framework.

This section walks the minimum-viable provider, then full CRUD with validators and plan modifiers, then a mini-project that templates files. **Pulumi is out of scope** per the session-level decision — the section focuses on Terraform/OpenTofu, which use the same plugin model.

---

## What you'll learn

- The shape of a Terraform provider: `provider.Provider` interface (`Metadata`/`Schema`/`Configure`/`Resources`/`DataSources`) + `resource.Resource` interface (`Metadata`/`Schema`/`Create`/`Read`/`Update`/`Delete`)
- `providerserver.Serve` — the gRPC entry point that the `terraform` CLI exec()s and connects to over stdin/stdout
- Typed values with ternary semantics: `types.String` / `types.Int64` / `types.Bool` carry **value / null / unknown** (a plain `string` can't represent "known after apply")
- The `tfsdk:"..."` reflection tag that bridges Go structs and Terraform values
- Validators (`stringvalidator.LengthAtLeast(1)`) run during `terraform plan` — error before Create is ever called
- Plan modifiers (`RequiresReplace()`, `UseStateForUnknown()`) shape what the diff *looks like* to the user
- The three-values doctrine: **Config** (raw HCL the user wrote) / **Plan** (intended next state) / **State** (what currently exists). Read from / write to whichever is the source of truth for the step.
- The canonical Refresh pattern: `Read` = state → model → remote API → state
- Data sources as read-only siblings of resources: same Schema, only `Read`
- Diagnostics with `Summary` (short noun phrase) + `Detail` (path + remediation hint) — always `SeverityError`, never `Warning` if work was prevented

---

## Mental model from other languages

There's no clean analog — IaC tools have their own model (declarative state, plan/apply cycle, drift detection). What you're learning is *how to extend* Terraform in Go, not IaC concepts themselves.

| Concept                      | Where it lives                                              |
| ---------------------------- | ----------------------------------------------------------- |
| Terraform provider plugin    | gRPC-served Go binary the `terraform` CLI invokes           |
| OpenTofu provider plugin     | same protocol — providers built for Terraform work directly |
| Pulumi resource provider     | similar pattern, separate ecosystem (skipped here)          |
| Crossplane provider          | Kubernetes-controller-style — `client-go` from section 08   |
| Custom resource (k8s CRD)    | A different model — see kubebuilder, not this section       |

**The framework's typed value model** is the part that surprises people coming from other ecosystems. `Path types.String` (not `Path string`) because every Terraform attribute can be value / null / unknown. "Unknown" = "known after apply" — the placeholder the CLI shows in plan output when the value depends on something that hasn't been created yet. A plain Go string has no way to represent that, so the framework wraps every primitive.

---

## The DevOps angle

You write a Terraform provider when:

- Your team has an internal service (a config store, a feature flag system, an internal SaaS) and you want infra-as-code coverage of it.
- An existing third-party SaaS has an API but no community provider, or the community provider doesn't cover the resource you need.
- You're consolidating "the shell script that runs after `terraform apply`" into the apply itself so it's part of the plan, the state, and the drift detection.

The non-obvious production details:

- **`terraform-plugin-framework` is the modern API.** The older `terraform-plugin-sdk/v2` still works but new providers should default to the framework — cleaner ergonomics, proper null/unknown semantics, structured diagnostics.
- **Local install path: `~/.terraform.d/plugins/registry.terraform.io/<namespace>/<name>/<version>/<os>_<arch>/`.** Build the provider binary, drop it there, declare the address in your provider config block, and `terraform apply` finds it. Documented in every example's README.
- **Validators run during `terraform plan`,** before Create/Update. The user sees a useful error at plan time instead of the Create call panicking on a bad value.
- **Plan modifiers run AFTER validators but BEFORE the CRUD handler.** `RequiresReplace()` turns an in-place Update into Destroy+Create. `UseStateForUnknown()` prevents the "(known after apply)" placeholder for computed attributes that are stable across applies.
- **`Read` must handle the "remote was deleted out-of-band" case** by calling `resp.State.RemoveResource(ctx)` — that's how Terraform learns the resource drifted; the next plan will offer to recreate it.
- **Acceptance tests gate on `TF_ACC=1`** by HashiCorp convention. The full wiring uses `terraform-plugin-testing` and a real `terraform` binary; the mini-project ships a skip-stub so the suite runs without it.

---

## Walkthrough

Read these in order. Each is a runnable example demonstrating one specific concept.

1. [`01-tf-provider-skeleton/`](./01-tf-provider-skeleton/) — the minimum-viable provider. One resource (`echo_value`) that copies `input` → `output`. `Read`/`Update`/`Delete` are stubs because there's no remote state to refresh. The TODO walks you through `providerserver.Serve`, the `Provider` + `Resource` interfaces, the model struct with `tfsdk:"..."` tags, and the local install path.
2. [`02-tf-provider-crud/`](./02-tf-provider-crud/) — full CRUD on the same `echo_value` resource. Adds: `stringvalidator.LengthAtLeast(1)` validators, `stringplanmodifier.RequiresReplace()` and `UseStateForUnknown()` plan modifiers, a computed `length` attribute the provider derives from `input`, a stable sha1-prefix `id`, and the canonical state→model→remote→state Refresh shape in `Read`. The README walks through `terraform init && terraform apply` against a locally-installed build.

---

## Mini-project: [`tf-provider-fileops`](./mini-project/)

A `fileops_templated_file` resource: render a Go `text/template` against a `vars` map, write the result to a local path. The Terraform-equivalent of a `local_file` resource with template rendering folded in.

The scaffold splits into `main.go` (provider + resource + model + CRUD TODO stubs) and `helpers.go` (4 pure I/O helpers: `RenderTemplate` / `WriteTemplatedFile` / `ReadTemplatedFile` / `DeleteTemplatedFile`). The 10 tests pin the production-grade contracts:

- **`Option("missingkey=error")`** so a typo'd template variable fails loudly. Default `text/template` substitutes `<no value>`, which is fine for log messages but disastrous for a config file Terraform tracks forever.
- **`errors.Is(err, os.ErrNotExist)` survives `os.ReadFile`'s wrapping** — `Read` depends on this sentinel to call `resp.State.RemoveResource(ctx)` for drift detection; if the sentinel disappears in a future refactor, drift detection silently breaks.
- **`Delete` is idempotent on missing files** — `terraform destroy` must be re-runnable.

There's also a `TestProviderSchemaCompiles` smoke test and a `TestAccTemplatedFile_SkipsWithoutTF_ACC` acceptance-test stub (full `terraform-plugin-testing` wiring is documented in the mini-project README as an extension exercise).

Spec and starter in [`mini-project/`](./mini-project/).

---

## Exercises

See [`exercises/`](./exercises/):

1. **[`01-tf-data-source`](./exercises/01-tf-data-source/)** — implement a `fileops_file_info` data source: read-only, `path` Required, `exists` Computed bool, `size` Computed int64. The pinned contract: a missing file → `Exists=false`, **NOT** an error — data sources are lenses, not guards. The user composes them into `count = data.x.exists ? 1 : 0` patterns.
2. **[`03-provider-error-handling`](./exercises/03-provider-error-handling/)** — `WriteFileDiagnostics(path, err) diag.Diagnostics` translates raw `os` errors into actionable user-facing diagnostics. Pinned: `os.ErrPermission` → mentions chmod/chown/privileges + path; `os.ErrNotExist` → mentions create-dir/mkdir; generic err → includes underlying error string + path; **all severities are `SeverityError`** (no warnings — if work was prevented, that's an error).

(Exercise 02 in the original PLAN was Pulumi-Go and was skipped per the session-level decision.)

---

## Further reading

- [`terraform-plugin-framework` docs](https://developer.hashicorp.com/terraform/plugin/framework) — the canonical reference for the modern API
- [`terraform-plugin-framework-validators`](https://github.com/hashicorp/terraform-plugin-framework-validators) — the bundled validators (string/int/list/...)
- [`terraform-plugin-testing`](https://developer.hashicorp.com/terraform/plugin/testing) — the acceptance-test framework; the mini-project's TF_ACC stub is the entry point
- [Terraform Plugin Protocol](https://developer.hashicorp.com/terraform/plugin/terraform-plugin-protocol) — the gRPC plane the framework abstracts away
- [HashiCorp tutorial: Implement Create](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-create) — the official "build a provider" tutorial; mirrors what example 02 does
