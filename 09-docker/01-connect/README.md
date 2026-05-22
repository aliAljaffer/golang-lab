# 01 — connect

Every tool that talks to the Docker daemon starts the same way: build a
`*client.Client`, defer-close it, make calls against it. This example is the
"hello daemon" smoke test.

## The two `Opt`s you almost always want

| Opt | Why |
|---|---|
| `client.FromEnv` | Read `DOCKER_HOST`, `DOCKER_TLS_VERIFY`, `DOCKER_CERT_PATH` from the env — same vars the `docker` CLI honours. |
| `client.WithAPIVersionNegotiation()` | Ask the daemon what API version it speaks, then downgrade the client to match. Without this, you ship with whatever API version the SDK was compiled against — and any daemon older than that rejects your calls. |

There are others (`WithHost`, `WithVersion`, `WithHTTPClient`) but you almost
never need them for a daemon on the same machine.

## API version negotiation, in one paragraph

The Docker daemon's HTTP API is versioned (`/v1.43/containers/...`). The SDK
defaults to whatever was current when the SDK was released. If your daemon is
older, you get `client version X is too new for daemon API version Y`.
`WithAPIVersionNegotiation()` makes a one-shot `GET /version` call up front
and stores the daemon's max version on the client. Cheap, painless, do it.

## Compare to other clients

|                       | Go (`docker/docker/client`)            | Python (`docker`)                | Bash (`docker` CLI)     |
|-----------------------|----------------------------------------|----------------------------------|-------------------------|
| Build client          | `client.NewClientWithOpts(FromEnv)`    | `docker.from_env()`              | (implicit)              |
| Server version        | `cli.ServerVersion(ctx)`               | `client.version()`               | `docker version`        |
| Close                 | `cli.Close()`                          | `client.close()`                 | (n/a)                   |

## TODO

1. Uncomment the TODO block.
2. Run `go run .` — print the daemon version + API version.
3. Try `DOCKER_HOST=tcp://127.0.0.1:1` go run . — observe the error path.
4. Also call `cli.Info(ctx)` and print `info.Name`, `info.Containers`, `info.Images`.
