# 01 — `os.Args`

The lowest-level CLI input. `os.Args` is a `[]string` where `os.Args[0]` is the program name and the rest are user-supplied tokens — already split on whitespace by the shell.

## What to notice

- It's just a slice. No parsing, no types, no help text.
- This is the layer every other parser is built on top of.
- In Python this is `sys.argv`; in Node `process.argv`; in Bash `$@`.

## Why doing this manually matters

You'll appreciate `flag` and `cobra` more after you've felt the pain of `if os.Args[i] == "--upper"`.
