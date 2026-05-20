# Exercise 01 — Tidy experiment

A CLI-only exercise. No code to write, but pay close attention to what changes.

## Goal

Build intuition for what `go mod tidy` actually does by watching `go.mod` and `go.sum` change in real time.

## Steps

1. **Create a scratch module** (anywhere outside this repo):
   ```bash
   mkdir /tmp/tidy-demo && cd /tmp/tidy-demo
   go mod init example.com/tidy-demo
   ```

2. **Inspect the initial files:**
   ```bash
   ls -la       # should see go.mod, no go.sum yet
   cat go.mod
   ```
   Self-check: ☐ `go.mod` exists, ☐ `go.sum` does NOT exist (no deps yet)

3. **Write code that imports something:**
   ```bash
   cat > main.go <<'EOF'
   package main
   import (
       "fmt"
       "github.com/google/uuid"
   )
   func main() { fmt.Println(uuid.New()) }
   EOF
   ```

4. **Try to run it BEFORE tidying:**
   ```bash
   go run .
   ```
   Self-check: ☐ You see an error mentioning `github.com/google/uuid is not in std`

5. **Tidy:**
   ```bash
   go mod tidy
   ```

6. **Inspect what changed:**
   ```bash
   cat go.mod
   cat go.sum
   ```
   Self-check: ☐ `go.mod` now has a `require` line for `uuid`, ☐ `go.sum` exists with 2+ lines for `uuid`

7. **Run again:**
   ```bash
   go run .
   ```
   Self-check: ☐ Prints a UUID

8. **Remove the dependency from code:** comment out the import and the `uuid.New()` call. Then:
   ```bash
   go mod tidy
   cat go.mod
   cat go.sum
   ```
   Self-check: ☐ The `require` line is gone, ☐ `go.sum` entries are pruned

## Reflection

- What's the difference between `go.mod` (intent) and `go.sum` (cryptographic lock)?
- Why does `go mod tidy` need to *both* add and remove entries? (Hint: it makes the manifest match what your code actually imports.)
- What would happen if you committed `go.mod` but not `go.sum`?
