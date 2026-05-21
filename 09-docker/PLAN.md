# Plan: 09-docker

## Prerequisites

- Docker daemon running locally (`docker info` should succeed)

## Concepts to cover

- [ ] Connecting to the daemon: `client.NewClientWithOpts(client.FromEnv)`
- [ ] API version negotiation
- [ ] Container lifecycle: create, start, stop, remove
- [ ] Image operations: pull, list, build, push
- [ ] Streaming output: logs, build output, exec output (all `io.Reader`)
- [ ] Events stream (the `docker events` equivalent)
- [ ] Filters (`filters.Args`)

## Examples to build

| Folder | Demonstrates |
|---|---|
| `01-connect/` | Connect to daemon, print server version |
| `02-list-containers/` | List running + all containers |
| `03-pull-and-run/` | Pull an image, create+start a container, stop+remove |
| `04-logs-stream/` | Follow a container's logs |
| `05-exec/` | Exec a command in a running container, capture output |
| `06-events/` | Subscribe to the events stream |

## Mini-project

**`image-pruner`** — finds and removes Docker images matching policy: untagged, older than N days, or with no associated containers. `--dry-run` and `--force` flags.

Tests verify:
- Correctly identifies prunable images (mock the client)
- `--dry-run` doesn't delete

## Exercises

1. **`01-container-stats`** — stream live CPU/memory stats for all running containers
2. **`02-buildkit-tar`** — build an image from an in-memory tar (no Dockerfile on disk)
3. **`03-restart-on-exit`** — watch events, restart any container that exits non-zero

## Status

- [ ] Concepts in README walkthrough
- [ ] Examples 01-06 built
- [ ] Mini-project `image-pruner` built + tested
- [ ] Exercises scaffolded

## Session Log

When a Claude session does work in this section, append an entry to the root [`SESSIONS.md`](../SESSIONS.md) before ending — do **not** log session history in this file. `PLAN.md` is the plan; `SESSIONS.md` is the history. Tick the Status boxes above as items complete.
