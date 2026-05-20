# 03 — Modules & Dependencies

This one is a guided CLI walkthrough — no `main.go`. The point is to feel how `go mod` works.

## Setup

In any scratch folder (NOT inside this repo):

```bash
mkdir /tmp/go-mod-demo && cd /tmp/go-mod-demo
go mod init example.com/demo
```

Inspect `go.mod`:
```
module example.com/demo

go 1.22
```

Just the module path and Go version. No `go.sum` yet — no deps.

## Add a dependency

Create a tiny program that uses an external library:

```bash
cat > main.go <<'EOF'
package main

import (
	"fmt"
	"github.com/google/uuid"
)

func main() {
	fmt.Println("new id:", uuid.New().String())
}
EOF
```

Try to run:
```bash
go run .
```

You'll see Go complain — the import isn't in `go.mod`. Fix it:
```bash
go mod tidy
```

Now inspect `go.mod`:
```
require github.com/google/uuid v1.6.0
```

And inspect `go.sum`:
```bash
cat go.sum
# two lines per dep: zip hash + go.mod hash
```

Now `go run .` works:
```bash
go run .
# new id: <some-uuid>
```

## Remove the dependency

Comment out the `import "github.com/google/uuid"` line and the `uuid.New()` call. Then:
```bash
go mod tidy
```

The `require` line disappears from `go.mod`. The `go.sum` entries get pruned. `go mod tidy` is the **truth-bringer** — it makes the manifest reflect what's actually imported.

## Pin a specific version

```bash
go get github.com/google/uuid@v1.5.0
cat go.mod   # now pinned to v1.5.0
```

## What this means

- `go.mod` is the manifest (declarative: "we want these deps")
- `go.sum` is the lock (cryptographic: "this is the exact bytes we trust")
- `go mod tidy` reconciles `go.mod` with what your code actually imports
- `go get pkg@version` is for explicit version changes

## Analogies

| Action | Go | Python (poetry) | TypeScript (npm) |
|---|---|---|---|
| Init project | `go mod init <path>` | `poetry init` | `npm init` |
| Add a dep | `go get pkg@v` | `poetry add pkg` | `npm install pkg` |
| Sync manifest ↔ imports | `go mod tidy` | (manual) | `npm prune` + reinstall |
| Lock file | `go.sum` | `poetry.lock` | `package-lock.json` |
| Commit lock? | **yes** | **yes** | **yes** |
