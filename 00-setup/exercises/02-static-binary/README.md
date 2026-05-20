# Exercise 02 — Static binary

Confirm with your own eyes that `go build` produces a single, statically-linked binary that doesn't depend on a Go installation.

## Steps

1. **Create a scratch program:**
   ```bash
   mkdir /tmp/static-demo && cd /tmp/static-demo
   go mod init example.com/static-demo
   cat > main.go <<'EOF'
   package main
   import "fmt"
   func main() { fmt.Println("from a binary near you") }
   EOF
   ```

2. **Build it:**
   ```bash
   go build -o hello
   ls -lh hello
   ```
   Self-check: ☐ The binary is 1-2 MB

3. **Verify it's statically linked:**
   ```bash
   file hello
   ```
   Self-check (Linux): output should contain `statically linked`
   Self-check (macOS): output is `Mach-O 64-bit executable arm64`. On Darwin you can also run `otool -L hello` — it should list very few (or no) dynamic library dependencies relative to a typical C binary.

4. **Run it without Go installed (simulate via Docker):**
   ```bash
   docker run --rm -v "$PWD:/app" alpine /app/hello
   ```
   On Linux/Mac amd64: this works. On macOS arm64 (Apple Silicon) you built a darwin binary, so cross-compile first:
   ```bash
   GOOS=linux GOARCH=amd64 go build -o hello-linux
   docker run --rm -v "$PWD:/app" alpine /app/hello-linux
   ```
   Self-check: ☐ The binary runs inside Alpine (which has no Go installed)

5. **Cross-compile for ARM64 Linux** (e.g. a Raspberry Pi):
   ```bash
   GOOS=linux GOARCH=arm64 go build -o hello-arm64
   file hello-arm64
   ```
   Self-check: ☐ Output mentions `aarch64` or `ARM aarch64`

## Reflection

- Why is the binary larger than the source code? (Hint: the Go runtime is embedded.)
- Why is "no runtime dependency" a big deal for DevOps tooling?
- What's the C equivalent of what just happened? (Hint: `gcc -static`, but cross-compiling C is famously painful — Go made it trivial.)
