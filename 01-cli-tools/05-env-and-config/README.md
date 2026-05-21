# 05 — Env vars + config files

Real tools layer config from multiple sources. `viper` standardizes the precedence so you don't reinvent it:

```md
flag > env > config file > default
```

## `os.Getenv` vs `os.LookupEnv`

|                     | Returns                   | Distinguishes unset from empty? |
| ------------------- | ------------------------- | ------------------------------- |
| `os.Getenv("X")`    | `string` (empty if unset) | ❌                              |
| `os.LookupEnv("X")` | `(string, bool)`          | ✅                              |

Use `LookupEnv` when "unset" means something different from "empty string" (e.g. allowing explicit `FOO=""` to override a default).

## Viper gotchas

- `viper.AutomaticEnv()` + `SetEnvPrefix("APP")` makes `viper.GetString("log_level")` look for `APP_LOG_LEVEL`.
- Add `SetEnvKeyReplacer(strings.NewReplacer("-", "_"))` so `--log-level` ↔ `APP_LOG_LEVEL` work.
- `ReadInConfig()` returns an error if the file isn't found — usually fine to ignore for optional config.

## Try it

```bash
go run .                              # default: info
echo "log_level: debug" > config.yaml
go run .                              # debug (from file)
APP_LOG_LEVEL=warn go run .           # warn (env wins)
APP_LOG_LEVEL=warn go run . --log-level=error   # error (flag wins)
```
