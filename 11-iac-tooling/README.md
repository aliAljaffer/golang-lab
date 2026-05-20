# 11 — IaC Tooling

> Status: ☐ not started — see [`PLAN.md`](./PLAN.md)

## What you'll learn

- The shape of a custom Terraform provider (using `terraform-plugin-framework`)
- Pulumi-Go basics
- Why so many IaC tools are written in Go

## Mental model from other languages

There's no clean analog here — IaC tools have their own model (declarative state, plan/apply, drift). What you're learning is *how to extend* IaC tools in Go, not the IaC concepts themselves (you likely know those from your DevOps background).

| Concept | Where it lives |
|---|---|
| Terraform provider plugin | grpc-served Go binary that Terraform invokes |
| Pulumi resource provider | similar pattern — grpc plugin |
| OpenTofu / Terragrunt | both written in Go, same plugin model as Terraform |

## The DevOps angle

If you ever need to manage a resource Terraform doesn't already know about, you write a provider. If you want first-class loops and conditionals in your IaC, you use Pulumi. Either way, Go is the language.

See [`PLAN.md`](./PLAN.md).
