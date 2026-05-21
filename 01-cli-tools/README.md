# 01 — CLI Tools

> Status: ☑ scaffolded — examples, mini-project, and exercises are all stubbed with failing tests. See [`PLAN.md`](./PLAN.md).

## What you'll learn

- Reading command-line args with `os.Args` and the stdlib `flag` package
- Building real CLIs with `cobra` and `urfave/cli`
- Subcommands, env vars, config files
- Exit codes, stdout vs stderr

## Mental model from other languages

| Concept | Go | Python | Bash | TS / Node |
|---|---|---|---|---|
| Arg parsing (basic) | `flag` | `argparse` | `getopts` | `process.argv` + `commander` |
| Arg parsing (rich) | `cobra` | `click` | n/a | `commander` / `yargs` |
| Env vars | `os.Getenv` | `os.environ.get` | `${VAR:-default}` | `process.env.VAR` |
| Exit codes | `os.Exit(1)` | `sys.exit(1)` | `exit 1` | `process.exit(1)` |
| Config files | `viper` | `pydantic-settings` | `source` | `dotenv` / `cosmiconfig` |

## The DevOps angle

CLIs are the bread and butter of ops tooling. Go's killer feature here: **one static binary, no runtime deps**. Ship a single file to a fresh server, run it. No `pip install`, no `node_modules`, no JVM. This is why `kubectl`, `terraform`, `docker`, `helm`, `gh`, and basically every modern DevOps tool is written in Go.

## Walkthrough

See [`PLAN.md`](./PLAN.md) for the planned examples and mini-project.
