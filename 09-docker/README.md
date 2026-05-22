# 09 — Docker

> Status: ☑ scaffolded — examples + mini-project + exercises ready; implementation + walkthrough pending. See [`PLAN.md`](./PLAN.md).

The Docker Engine API is HTTP-over-Unix-socket; the official Go SDK (`github.com/docker/docker/client`) is a typed client on top of it. Most days you should just shell out to `docker`; the SDK earns its keep when you need a long-running daemon, fine-grained event filtering, or programmatic control of container lifecycle (image janitors, build farms, container-based test harnesses, autorestart sidecars).

This section walks the connection, container listing, pull-and-run sequence, log streaming, exec, and the event stream — the shape every Docker-driving Go program shares.

---

## What you'll learn

- `client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())` — and why API negotiation is non-optional
- `ContainerList` with `filters.NewArgs(filters.Arg("status", "running"))` — the SDK's filter DSL
- The 6-step run sequence: `ImagePull` → drain the reader → `ContainerCreate` → `ContainerStart` → `ContainerWait` (two-channel) → `stdcopy.StdCopy` → `ContainerRemove`
- Log streaming with `ContainerLogs(..., {Follow: true})` and the multiplex frame format (the `[stream code 1B][len 4B][payload]` layout that `StdCopy` decodes; TTY containers skip the multiplex)
- The two-phase exec API: `ContainerExecCreate` → `ContainerExecAttach` (this is what starts it) → `StdCopy` → `ContainerExecInspect` for the exit code
- The event stream: `cli.Events(ctx, events.ListOptions{Filters: f})` two-channel return + filter-at-source

---

## Mental model from other languages

| Concept            | Go (`docker/docker/client`)                          | Python (`docker-py`)             | Bash (`docker` CLI)        |
| ------------------ | ---------------------------------------------------- | -------------------------------- | -------------------------- |
| Client             | `client.NewClientWithOpts(...)`                      | `docker.from_env()`              | (implicit)                 |
| List containers    | `cli.ContainerList(ctx, container.ListOptions{...})` | `client.containers.list()`       | `docker ps`                |
| Pull image         | `cli.ImagePull(ctx, ref, ImagePullOptions{})`        | `client.images.pull(...)`        | `docker pull`              |
| Run container      | `ContainerCreate` + `ContainerStart`                 | `client.containers.run(...)`     | `docker run`               |
| Exec into          | `ContainerExecCreate` + `ContainerExecAttach`        | `container.exec_run(...)`        | `docker exec`              |
| Build image        | `cli.ImageBuild(ctx, tarReader, opts)`               | `client.images.build(...)`       | `docker build`             |
| Event stream       | `cli.Events(ctx, opts)` — two channels               | `client.events(decode=True)`     | `docker events`            |

**The cultural difference:** the SDK does not provide a one-shot `containers.run(...)` like `docker-py`. You build it yourself out of the 4-5 primitive calls; that's the price of admission for fine control over each step (e.g., "pull, but only if not present"; "start, but stream logs concurrently from a goroutine"). The mini-project and example 03 are the canonical walkthroughs of that sequence.

---

## The DevOps angle

The most common production use cases:

- **Image cleanup ("janitor") tools**, like the mini-project — clusters of build hosts run out of disk if nobody prunes; tag-by-policy is more nuanced than `docker image prune`.
- **Long-running event consumers**, like the `restart-on-exit` exercise — watch `events.Type=container, Action=die, exitCode≠0`, restart, alert.
- **Test harnesses** that spin up real services in containers (Testcontainers is essentially this pattern productized).
- **Build pipelines** that need to build, push, scan, and tag in one orchestrated sequence.

The non-obvious production details:

- **`client.WithAPIVersionNegotiation()` is non-optional.** Without it, your tool pins to the SDK's compile-time max API version, and any older daemon rejects your calls with "client too new."
- **You MUST drain the `ImagePull` reader.** The daemon considers the pull "in progress" until the response body is read to EOF. `io.Copy(io.Discard, rc)` is mandatory even if you don't care about the bytes — the #1 source of "my pull silently doesn't finish" bugs.
- **Use `stdcopy.StdCopy`, not `io.Copy`, on multiplexed log streams.** Without a TTY, Docker multiplexes stdout+stderr over one stream with an 8-byte frame header; plain `io.Copy` writes the header bytes into your output and makes the logs unreadable.

---

## Walkthrough

Read these in order. Each is a runnable example demonstrating one specific concept. **All require a running Docker daemon** (`docker info` must succeed); they compile and lint clean without one.

1. [`01-connect/`](./01-connect/) — `client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())` + `cli.ServerVersion(ctx)` smoke probe. The minimum "am I talking to Docker?" test.
2. [`02-list-containers/`](./02-list-containers/) — `ContainerList` with the `filters.Args` DSL. The `Names` field's leading `/` is a quirk to know about.
3. [`03-pull-and-run/`](./03-pull-and-run/) — the full 6-step run sequence. The drain-the-reader rule is the headline footgun.
4. [`04-logs-stream/`](./04-logs-stream/) — `ContainerLogs` + `{Follow: true}` + the multiplex frame format + the TTY exception that breaks `StdCopy` (Inspect first to find out).
5. [`05-exec/`](./05-exec/) — the two-phase exec API. The exit code lives in a separate `ContainerExecInspect` call after the stream finishes, which is the part everyone forgets.
6. [`06-events/`](./06-events/) — `cli.Events(ctx, opts)` two-channel return + the per-event `Type`/`Action`/`Actor` shape. Filter at the daemon ("only `container die`") rather than client-side — saves bandwidth and CPU.

---

## Mini-project: [`image-pruner`](./mini-project/)

Policy-driven image cleanup. Three policies OR'd: `--untagged` (remove `<none>:<none>`), `--max-age <duration>` (remove images older than X), `--no-containers` (remove images with no running container reference). Plus `--dry-run` and `--force`.

The point: a real cleanup tool combines listing, plan generation (deterministic), and execution with safety rails (dry-run, force opt-in, error logging without aborting the run). The 14 tests use a `DockerAPI` interface + a `captureRemoveFake` to assert the `--force` flag threads through to `RemoveOptions.Force` — a real regression-catcher.

Spec and starter in [`mini-project/`](./mini-project/).

---

## Exercises

See [`exercises/`](./exercises/):

1. **[`01-container-stats`](./exercises/01-container-stats/)** — implement `CPUPercent(curr, prev Snapshot) float64` + `MemoryPercent(s Snapshot) float64`. The math (cumulative-counter → delta → percentage with `onlineCPUs`) is where most people get it wrong; tests pin counter-reset, first-sample NaN, and division-by-zero edge cases.
2. **[`02-buildkit-tar`](./exercises/02-buildkit-tar/)** — assemble an in-memory tar suitable for `cli.ImageBuild`. Tests pin tar-terminator-block-present (catches missing `tw.Close()`) and binary-data-survives.
3. **[`03-restart-on-exit`](./exercises/03-restart-on-exit/)** — `ShouldRestart(events.Message) bool` + `Run(ctx, DockerAPI)`. Watches the event stream for `container die` with non-zero exit code, restarts the container, continues on transient errors. Practices the channel-backed fake from section 06.

---

## Further reading

- [Docker Engine API reference](https://docs.docker.com/reference/api/engine/) — the underlying HTTP API; the SDK is a typed wrapper
- [`docker/docker/client` Go docs](https://pkg.go.dev/github.com/docker/docker/client) — the Go-level reference; the godoc on each method is the primary doc
- [Testcontainers for Go](https://golang.testcontainers.org/) — the most-used production consumer of this SDK; reads as a worked tutorial
- [`stdcopy` package](https://pkg.go.dev/github.com/docker/docker/pkg/stdcopy) — the multiplex decoder used in examples 03/04
- [The `+incompatible` go.mod suffix](https://go.dev/ref/mod#non-module-compat) — why `docker/docker` imports as `v28.5.2+incompatible`
