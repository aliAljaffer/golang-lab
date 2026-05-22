# Mini-project — `tf-provider-fileops`

A real Terraform provider that manages templated local files. End-to-end CRUD with drift detection.

## What it does

```hcl
resource "fileops_templated_file" "config" {
  path     = "/etc/myapp/config.yaml"
  template = <<-EOT
    listen: ":{{.port}}"
    debug:  {{.debug}}
  EOT
  vars = {
    port  = "8080"
    debug = "true"
  }
}
```

`terraform apply` renders the template, writes the file, and tracks it in state. Subsequent applies re-render only if `template` or `vars` change. Manual deletion of the file is detected as drift on the next refresh. `terraform destroy` removes the file.

## File layout

| File           | What's in it                                                   |
|----------------|----------------------------------------------------------------|
| `main.go`      | provider + resource (`fileops_templated_file`) — TODO CRUD     |
| `helpers.go`   | the four pure I/O helpers — TODO bodies                        |
| `main_test.go` | unit tests for the helpers + provider smoke test + TF_ACC stub |

The split is deliberate. The CRUD handlers in `main.go` are framework glue (read plan, set state). The actual work — template rendering, file I/O — lives in `helpers.go`. **All the interesting tests target the helpers.** This is how production Terraform providers are usually structured: keep the framework surface thin.

## What to fill in

1. `helpers.go` — the four functions. `RenderTemplate` is the only non-obvious one (`Option("missingkey=error")` is the production stance).
2. `main.go` — six CRUD blocks (Create/Read/Update/Delete + the model wiring). Each is a 5-7 line dance: `Get → call helper → Set` (with diagnostics).

When `go test -tags=exercise ./11-iac-tooling/mini-project/` is green, you're done.

## Why `missingkey=error`

Default `text/template` substitutes the string `<no value>` for a missing variable:

```go
template.Must(template.New("").Parse("hi {{.name}}")).Execute(os.Stdout, map[string]string{})
// → "hi <no value>"
```

That's fine for log messages. For a tool that **writes the result to a file Terraform will track forever**, it's a footgun:

```yaml
listen: ":<no value>"
debug:  <no value>
```

…now lives in `/etc/myapp/config.yaml`. The pinned stance: any missing var is a plan-time error. The user can decide whether to add it, remove the placeholder, or accept the gap (e.g., default-empty via `{{or .name ""}}`).

## Why DeleteTemplatedFile is idempotent

`terraform destroy` is re-runnable. If the previous destroy was interrupted, or someone manually removed the file, the second destroy must succeed. The pattern is universal in providers — any "remove the underlying resource" handler should treat 404 as success.

## Drift detection — what makes it work

When you `terraform refresh` (or implicitly during `plan`/`apply`), the provider's `Read` handler runs. It:

1. Reads the resource's state (the path).
2. Reads the actual file from disk (`ReadTemplatedFile`).
3. If the file is gone → `resp.State.RemoveResource(ctx)`. Terraform plans a recreate.
4. If the contents changed → state's `content` updates, plan shows the drift.

The `errors.Is(err, os.ErrNotExist)` check is what makes drift work. If `ReadTemplatedFile` lost the sentinel error (e.g., by wrapping it in a custom message), drift detection silently breaks. There's a regression test for that — `TestReadTemplatedFile_NotFoundReturnsErrNotExist`.

## Extending to a real acceptance test

The skeleton has `TestAccTemplatedFile_SkipsWithoutTF_ACC` as a placeholder. A full acceptance test:

```go
import (
    "github.com/hashicorp/terraform-plugin-framework/providerserver"
    "github.com/hashicorp/terraform-plugin-go/tfprotov6"
    "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var testProtoV6 = map[string]func() (tfprotov6.ProviderServer, error){
    "fileops": providerserver.NewProtocol6WithError(NewProvider()),
}

func TestAccTemplatedFile_Lifecycle(t *testing.T) {
    if os.Getenv("TF_ACC") == "" { t.Skip("TF_ACC not set") }
    tmpFile := filepath.Join(t.TempDir(), "out.txt")
    resource.Test(t, resource.TestCase{
        ProtoV6ProviderFactories: testProtoV6,
        Steps: []resource.TestStep{
            { // Create
                Config: fmt.Sprintf(`resource "fileops_templated_file" "x" {
                    path = %q
                    template = "hello {{.name}}"
                    vars = { name = "world" }
                }`, tmpFile),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("fileops_templated_file.x", "content", "hello world"),
                ),
            },
            { // Update (in-place)
                Config: fmt.Sprintf(`resource "fileops_templated_file" "x" {
                    path = %q
                    template = "hi {{.name}}"
                    vars = { name = "world" }
                }`, tmpFile),
                Check: resource.TestCheckResourceAttr("fileops_templated_file.x", "content", "hi world"),
            },
        },
    })
}
```

Requires `terraform` (or `tofu`) on `PATH`. The test framework exec()s the binary, points it at the in-process provider, and drives plan/apply/destroy for you.

## Why this isn't a `null_resource` or `local_file`

- `null_resource` runs a `local-exec` provisioner — but provisioners are a [last-resort feature](https://developer.hashicorp.com/terraform/language/resources/provisioners/syntax#provisioners-are-a-last-resort) per HashiCorp. They don't track drift, can't be refreshed, and run shell commands.
- `local_file` from the `hashicorp/local` provider does write files — but it doesn't do template rendering. You'd compose `templatefile()` (a Terraform function) with `local_file`, which loses validation of the template at plan time and has clunkier diff messages.
- This provider gets you typed schema for `template` and `vars`, atomic write semantics, and drift detection of the file's actual contents — all visible in `terraform plan` output.
