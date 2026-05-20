# 01 — Hello World

The smallest possible Go program.

## Run it

```bash
go run ./00-setup/01-hello-world
```

You should see:
```
hello, devops
```

## What to notice

- **`package main`** is special. Any other package name would compile as a library (`go build` produces nothing executable).
- **Imports are explicit.** Unused imports are *compile errors*, not warnings. (Python/TS get away with unused imports; Go doesn't.)
- **`fmt.Println`** lives in the `fmt` package. The convention `package.Function` is everywhere — there's no `from fmt import Println`.
- **No semicolons.** The Go parser inserts them based on line endings. Don't add them.
- **The opening `{`** must be on the *same line* as the function declaration. This isn't a style preference; it's a parser rule. (Yes, this catches every Java developer once.)

## Try this

1. Comment out the `import "fmt"` line and try `go run` again — observe the compile error.
2. Rename `main` to `Main` — observe what happens. (Hint: Go is case-sensitive about exports.)
3. Add an unused variable inside `main()`: `x := 42` — observe the compile error.
