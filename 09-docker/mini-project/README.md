# Mini-project — `image-pruner`

Finds and removes Docker images by policy. The three policies (OR'd):

| Flag | Removes |
|---|---|
| `--untagged` | dangling images (`<none>:<none>` or empty `RepoTags`) |
| `--max-age <duration>` | images whose `Created` is older than this |
| `--no-containers` | images with no container references (running OR stopped) |

Plus `--dry-run` (print, don't mutate) and `--force` (pass `Force=true` to
`ImageRemove` — needed when a stopped container is using the image).

## Testable surface

```go
type DockerAPI interface {
    ImageList(ctx, opts) ([]image.Summary, error)
    ImageRemove(ctx, id, opts) ([]image.DeleteResponse, error)
    ContainerList(ctx, opts) ([]container.Summary, error)
}

type Policy struct {
    RemoveUntagged   bool
    RemoveOlderThan  time.Duration  // zero == ignore
    RemoveUnused     bool
    DryRun           bool
    Force            bool
}

func Plan(images, containers, policy, now) []string           // pure
func Sync(ctx, api, policy, now, out) ([]string, error)        // does the work
func isUntagged(image.Summary) bool                            // unexported helper
```

A real `*client.Client` satisfies `DockerAPI` naturally — it has those three
methods with matching signatures. Tests pass a `fakeDocker`.

## What the tests verify

| Test | Concept |
|---|---|
| `TestIsUntagged_*` (3) | Nil tags / `<none>:<none>` / real tag |
| `TestPlan_RemovesUntagged` | Dangling-image policy |
| `TestPlan_RemovesOldImages` | Age cutoff (clock injected) |
| `TestPlan_KeepsImagesWithContainerReferences` | Referenced images are protected |
| `TestPlan_PoliciesAreOred` | Multiple policies combine with OR, not AND |
| `TestPlan_EmptyPolicyEmptyResult` | No policy = no-op success |
| `TestPlan_OutputIsSorted` | Deterministic order |
| `TestSync_HappyPathRemovesImages` | End-to-end with fake API |
| `TestSync_DryRunMakesNoRemoveCalls` | `--dry-run` must not mutate |
| `TestSync_ListErrorPropagates` | List failure aborts |
| `TestSync_RemoveErrorPropagates` | Remove failure aborts |
| `TestSync_ForceFlagThreadsThrough` | `--force` actually reaches `RemoveOptions.Force` |

All tests run against a `fakeDocker` — no Docker daemon needed.

## How to run (once you've implemented it)

```bash
# dry run, untagged only
go run ./09-docker/mini-project --untagged --dry-run

# images older than a week, with no container refs
go run ./09-docker/mini-project --max-age 168h --no-containers

# actually remove (forces removal even if stopped containers reference it)
go run ./09-docker/mini-project --untagged --force
```

## Notes on policy OR-ing

If you set `--untagged --max-age 168h`, an image is flagged when it's
**untagged OR older than 7d** (not both). That matches `docker image prune
--filter "until=168h"` semantics. The `RemoveUnused` policy is a separate
axis — when set, even a still-referenced image survives.

## Why `--force` exists

The Docker daemon refuses to remove an image that any container — even a
stopped one — references, unless you pass `Force=true`. The flag tests
catch the regression where `Force` is wired into `Policy` but forgotten in
`RemoveOptions`. Production code should generally NOT default to `--force`;
the failed-removal error is informative.
