# 05 — EC2 list with tag filter

Same SDK v2 shape as S3, different service. Two things worth noting:

## Reservations vs Instances

`DescribeInstances` returns `[]Reservation`, and each Reservation has
`[]Instance`. A "reservation" is the original RunInstances call that
launched a batch — historically meaningful, today mostly noise. Code that
wants every instance must loop twice:

```go
for _, r := range page.Reservations {
    for _, inst := range r.Instances {
        // ...
    }
}
```

The exercise `02-find-untagged` flattens this for you.

## Filters: push, don't pull

`Filters` narrows the result *on the AWS side*. Always use it when you can —
listing 5,000 instances client-side just to keep 3 is wasteful and slow.

Filter names follow a string DSL:

| Filter name | Matches |
|---|---|
| `instance-state-name` | `pending`, `running`, `stopped`, … |
| `tag:Env` | instances with tag Key=Env, Value matching |
| `tag-key` | instances with any tag of this Key |
| `vpc-id` | instances in this VPC |

The full list is in the AWS API reference.

## Enum types

`inst.State.Name` is `ec2types.InstanceStateName` — a string enum:
`"pending"`, `"running"`, `"stopping"`, `"stopped"`, …

`inst.InstanceType` is `ec2types.InstanceType` — likewise. You can compare
with `== ec2types.InstanceStateNameRunning`.

## TODO

1. Fill in the TODO blocks.
2. Run `go run .` — list every instance.
3. Tag one instance with `Env=test` and run `go run . Env test`.
4. Add a second filter for `instance-state-name=running` and combine — AND
   semantics across filters, OR within `Values`.
