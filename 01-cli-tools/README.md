# 01 — CLI Tools

> Status: ☑ scaffolded — examples, mini-project, and exercises are all stubbed with failing tests. See [`PLAN.md`](./PLAN.md).

A CLI is the smallest useful Go program: read flags, do work, exit with a code. Most DevOps tooling never grows past this shape. `kubectl`, `terraform`, `helm`, `gh`, `docker`, `aws` — all CLIs, all written in Go for the same reason: one static binary, no runtime, fast startup.

This section walks from `os.Args` (the rawest interface) up to a `kubectl`-style nested-subcommand tool with config files and env-var overrides.

---

## What you'll learn

- Reading command-line args with `os.Args` and the stdlib `flag` package
- Building real CLIs with `cobra` (subcommands, help, completion) and `urfave/cli` (the lighter alternative)
- Subcommands, env vars, config files, precedence order
- Exit codes, stdout vs stderr, `log.Fatal` vs `os.Exit` (and why the difference matters)

---

## Mental model from other languages

| Concept             | Go                  | Python                 | Bash                | TS / Node                    |
| ------------------- | ------------------- | ---------------------- | ------------------- | ---------------------------- |
| Arg parsing (basic) | `flag`              | `argparse`             | `getopts`           | `process.argv` + `commander` |
| Arg parsing (rich)  | `cobra`             | `click`                | n/a                 | `commander` / `yargs`        |
| Env vars            | `os.Getenv`         | `os.environ.get`       | `${VAR:-default}`   | `process.env.VAR`            |
| Exit codes          | `os.Exit(1)`        | `sys.exit(1)`          | `exit 1`            | `process.exit(1)`            |
| Config files        | `viper`             | `pydantic-settings`    | `source`            | `dotenv` / `cosmiconfig`     |
| Stderr vs stdout    | `fmt.Fprintln(os.Stderr, ...)` | `print(..., file=sys.stderr)` | `>&2 echo ...` | `console.error(...)` |

**The cultural difference:** in Go, you almost never see a third-party "convenience" wrapper that abstracts argv away. `flag` is in the stdlib and is fine for ~80% of internal tools. Reach for `cobra` when you have ≥2 subcommands or want generated `--help` / shell completion. Reach for `viper` only when you're sourcing config from multiple places (file + env + flags) and need a precedence story.

---

## The DevOps angle

CLIs are the bread and butter of ops tooling. Go's killer feature here: **one static binary, no runtime deps.** Ship a single file to a fresh server, run it. No `pip install`, no `node_modules`, no JVM. This is why `kubectl`, `terraform`, `docker`, `helm`, `gh`, and basically every modern DevOps tool is written in Go.

**Exit codes are an API.** A CI step depends on `$?`. The shell's `&&` and `||` depend on `$?`. A `Makefile` rule's success depends on `$?`. If your CLI panics-and-prints instead of returning a clean non-zero code, you've broken every downstream consumer. Section 06 (`exit-codes`) is the most important example here for ops work.

---

## Walkthrough

Read these in order. Each is a runnable example demonstrating one specific concept.

1. [`01-os-args/`](./01-os-args/) — `os.Args` as a plain `[]string`. The raw interface; everything above is sugar on top.
2. [`02-flag-basics/`](./02-flag-basics/) — `flag.String`, `flag.Bool`, `flag.Int`, `flag.Parse()`. The stdlib answer; zero deps; built into every Go binary.
3. [`03-cobra-hello/`](./03-cobra-hello/) — the smallest `cobra` app with a single subcommand. Demonstrates the `Command{Use, Short, Run}` shape that every cobra binary uses.
4. [`04-cobra-nested/`](./04-cobra-nested/) — nested subcommands (`tool resource action`). The `kubectl get pods` / `gh pr create` pattern.
5. [`05-env-and-config/`](./05-env-and-config/) — env-var lookup with `os.LookupEnv` vs `os.Getenv` (the `_, ok :=` distinction matters when "set to empty string" ≠ "unset"), `viper`-backed config loading, and the canonical precedence order: **flag > env > config file > default**.
6. [`06-exit-codes/`](./06-exit-codes/) — exit-code discipline, `stderr` for errors and progress, `stdout` for parseable output. The `log.Fatal` footgun: it calls `os.Exit(1)` **without** running deferred functions, so any open files / DB connections / lock files leak.

---

## Mini-project: [`dirsize`](./mini-project/)

A CLI that walks a directory and prints sizes per subdirectory — like `du -sh *` but sortable, with `--top N` and `--json` flags.

The point: a real CLI has a filesystem walk + sort + two output formats + clean exit codes. Tests pin: exit 0 on a valid path, exit 1 on a missing path, `--json` produces parseable JSON, `--top 3` returns at most 3 entries.

Spec and starter in [`mini-project/`](./mini-project/).

---

## Exercises

See [`exercises/`](./exercises/):

1. **[`01-greplite`](./exercises/01-greplite/)** — a minimal `grep` with `-i` (case-insensitive) and `-n` (line numbers) flags. Practices `bufio.Scanner` line-by-line reading + flag composition.
2. **[`02-envdump`](./exercises/02-envdump/)** — print env vars matching a glob, with `--unset` to clear them. Practices `os.Environ()` + glob matching + side-effect flags.
3. **[`03-multi-subcommand`](./exercises/03-multi-subcommand/)** — a `kubectl`-style CLI (`tool get pods`, `tool delete pod X`). Practices cobra nesting + arg validation.

---

## Further reading

- [`flag` package docs](https://pkg.go.dev/flag) — the stdlib, dense but complete
- [`cobra` user guide](https://github.com/spf13/cobra/blob/main/user_guide.md) — the de-facto standard for non-trivial CLIs
- [`urfave/cli` docs](https://cli.urfave.org/) — the lighter alternative; one big builder, less ceremony
- [`viper` docs](https://github.com/spf13/viper) — config-from-anywhere; the natural pairing with cobra
- [12-factor: Config](https://12factor.net/config) — the env-var doctrine that informs `05-env-and-config`'s precedence order
