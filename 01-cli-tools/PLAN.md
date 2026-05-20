# Plan: 01-cli-tools

## Concepts to cover

- [ ] `os.Args` — the lowest level
- [ ] `flag` package — stdlib arg parsing
- [ ] `cobra` — subcommands, help text, generation
- [ ] `urfave/cli` — alternative to cobra (lighter)
- [ ] Env vars + `os.LookupEnv` vs `os.Getenv`
- [ ] Config files with `viper`
- [ ] Exit codes, stderr vs stdout, `log.Fatal` vs `os.Exit`

## Examples to build

| Folder | Demonstrates |
|---|---|
| `01-os-args/` | Raw `os.Args` slice; manual parsing |
| `02-flag-basics/` | Stdlib `flag.String`, `flag.Bool`, `flag.Parse` |
| `03-cobra-hello/` | Smallest cobra app with one subcommand |
| `04-cobra-nested/` | Nested subcommands (`tool resource action`) |
| `05-env-and-config/` | Env var precedence, viper-based config loading |
| `06-exit-codes/` | Proper exit codes, stderr usage, `log.Fatal` pitfalls |

## Mini-project

**`dirsize`** — CLI that walks a directory and prints sizes per subdirectory (like `du -sh *` but sortable, with `--top N` and `--json` flags).

Tests verify:
- Exits 0 on a valid path, 1 on a missing path
- `--json` produces parseable JSON
- `--top 3` returns at most 3 entries

## Exercises

1. **`01-greplite`** — implement a minimal `grep` (case-insensitive flag, line numbers flag)
2. **`02-envdump`** — print env vars matching a glob, with `--unset` to clear them
3. **`03-multi-subcommand`** — build a `kubectl`-style CLI shell (`tool get pods`, `tool delete pod X`)

## Status

- [ ] Concepts documented in README.md walkthrough
- [ ] Examples 01-06 built
- [ ] Mini-project `dirsize` built + tested
- [ ] Exercises scaffolded
