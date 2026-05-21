# 03 — Cobra, hello world

[`spf13/cobra`](https://github.com/spf13/cobra) is the CLI framework behind `kubectl`, `helm`, `gh`, `docker`, `hugo`. If you're going to build DevOps tooling in Go, you'll meet it constantly.

## Why cobra over `flag`

- Subcommands as a first-class concept (`tool resource action`).
- Auto-generated `--help` for every command at every level.
- Shell completion generation built in (`bash`, `zsh`, `fish`, `pwsh`).
- Required flags, mutually exclusive flags, custom validators.

## Core types

- `cobra.Command` — one node in the tree. Has `Use`, `Short`, `Long`, `Run`/`RunE`, flags.
- **`Run` vs `RunE`**: prefer `RunE` so errors bubble out and `Execute()` can set the exit code.
- **`Flags()` vs `PersistentFlags()`**: persistent flags inherit to child commands.

## After implementing

```bash
go run . --help
go run . greet --help
go run . greet --name Ali
```
