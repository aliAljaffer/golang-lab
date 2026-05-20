# 09 — Docker

> Status: ☐ not started — see [`PLAN.md`](./PLAN.md)

## What you'll learn

- Using the Docker SDK (`docker/docker/client`) to talk to the local daemon
- Listing, inspecting, exec-ing into containers
- Building images programmatically
- Streaming logs

## Mental model from other languages

| Concept | Go (`docker/docker/client`) | Python (`docker-py`) | Bash (`docker` CLI) |
|---|---|---|---|
| List containers | `cli.ContainerList(ctx, opts)` | `client.containers.list()` | `docker ps` |
| Pull image | `cli.ImagePull(ctx, ref, opts)` | `client.images.pull(...)` | `docker pull` |
| Exec into | `cli.ContainerExecCreate` + `ExecStart` | `container.exec_run(...)` | `docker exec` |
| Build image | `cli.ImageBuild(ctx, ...)` | `client.images.build(...)` | `docker build` |

## The DevOps angle

Useful for: image janitors, container introspection tools, build pipelines that need fine-grained control. For most use cases you'd just shell out to `docker` — but when you need a long-running daemon or fine event filtering, the SDK is the right call.

See [`PLAN.md`](./PLAN.md).
