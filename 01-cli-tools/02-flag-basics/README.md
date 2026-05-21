# 02 — `flag` package

Stdlib argument parsing. No subcommands, no fancy help, but zero dependencies.

## Mental model

| You write | What it returns |
|---|---|
| `name := flag.String("name", "world", "...")` | `*string` (note: pointer) |
| `shout := flag.Bool("shout", false, "...")` | `*bool` |
| `count := flag.Int("count", 1, "...")` | `*int` |

After `flag.Parse()`, dereference with `*name`, `*shout`, `*count`.

## Things to learn the hard way

- `flag` accepts both `-name=Ali` and `--name Ali`, but **not** `--name=Ali` with combined short flags like `-vvv`.
- Positional args after flags are available via `flag.Args()`.
- Override `flag.Usage` to control `-h` / `--help` output.

## Comparison

- Python `argparse`: similar feel, more batteries (subparsers, type coercion).
- Bash `getopts`: positional-only, painful.
- Node `process.argv` + `commander`: closer to cobra than to `flag`.
