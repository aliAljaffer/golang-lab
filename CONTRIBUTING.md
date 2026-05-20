# Contributing

This repo is primarily a personal learning journal, but contributions and corrections are welcome.

## Conventions

- Each numbered section (`NN-topic/`) has a fixed structure — see [`PLAN.md`](./PLAN.md) for the template.
- `README.md` files are public-facing notes. `PLAN.md` files are working roadmaps for what to build in that section.
- Code examples should be **minimal but runnable**. Each example sits in its own subfolder with a `main.go` and a short README.
- Exercises ship as **failing test files**. Readers write code until `go test` passes. No solutions in the repo.

## Style

- Run `gofmt -w .` before committing — no exceptions.
- Run `go vet ./...` — must be clean.
- Run `go test ./...` — must pass. Exercises that ship with failing tests use the `//go:build exercise` build tag, so they're excluded from default test runs. To run an exercise's tests, pass `-tags=exercise` (e.g. `go test -tags=exercise ./00-setup/exercises/03-env-explorer/`).

## Obsidian notes

If you use Obsidian as a vault:
- Don't commit `.obsidian/workspace.json` — it churns every time you open the vault. `.gitignore` already excludes the whole `.obsidian/` folder.
- For cross-references inside notes, prefer `[label](./path/to/file.md)` over `[[wikilinks]]` so links render on both GitHub and Obsidian.

## PRs

- One concept per PR — easier to review and learn from.
- Link to the section's `PLAN.md` if your PR completes a planned item.
