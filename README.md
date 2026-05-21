# golang-lab — A Self-Paced Go-for-DevOps Bootcamp

A hands-on, self-paced curriculum to take you from "I've heard of Go" to "I can ship Go-based DevOps tooling." Free, public, and community-supported.

> **New here?** Start with [`BOOTCAMP.md`](./BOOTCAMP.md) for how the bootcamp works and how to ask for help.

Also designed to open cleanly as an [Obsidian](https://obsidian.md/) vault for personal note-taking.

## Who this is for

- Engineers comfortable with **at least one** of Python, Bash, TypeScript, or Java who want to add Go.
- Specifically for **DevOps / Platform / SRE / Cloud** work — CLIs, HTTP clients & servers, AWS/k8s/Docker SDKs, concurrency for parallel infra tasks, observability.
- People who learn by doing. Every section has runnable examples + a mini-project + exercises that ship as failing tests; you write code until they pass.

## Who this isn't for

- People wanting a comprehensive Go language reference (use [the Go Programming Language Specification](https://go.dev/ref/spec) and ["The Go Programming Language" by Donovan & Kernighan](https://www.gopl.io/)).
- People building games, ML systems, or low-latency trading code in Go (the language is great for those — this repo just doesn't focus on them).

## How to participate

- **Work through the sections** at your own pace (suggested: 1 section per week, no due dates).
- **Ask for help in [Discussions](https://github.com/alialjaffer/golang-lab/discussions)** when stuck. Open-ended questions live here.
- **File an [Issue](https://github.com/alialjaffer/golang-lab/issues)** when you hit a bug, typo, or have a concrete improvement. Three templates are provided.
- **Help others** if you've already finished a section. That's how this becomes more than a personal repo.

See [`BOOTCAMP.md`](./BOOTCAMP.md) for the full participation guide and how to ask a good question.

## Roadmap

Each section is a folder. Status legend: ☐ not started · ◐ in progress · ☑ done.

| #   | Section                                       | Status |
| --- | --------------------------------------------- | ------ |
| 00  | [Setup & toolchain](./00-setup/README.md)     | ☑      |
| 01  | [CLI tools](./01-cli-tools/README.md)         | ☐      |
| 02  | [Files & OS](./02-files-and-os/README.md)     | ☐      |
| 03  | [HTTP clients](./03-http-clients/README.md)   | ☐      |
| 04  | [HTTP servers](./04-http-servers/README.md)   | ☐      |
| 05  | [Concurrency](./05-concurrency/README.md)     | ☐      |
| 06  | [Testing](./06-testing/README.md)             | ☐      |
| 07  | [AWS SDK](./07-aws/README.md)                 | ☐      |
| 08  | [Kubernetes](./08-kubernetes/README.md)       | ☐      |
| 09  | [Docker](./09-docker/README.md)               | ☐      |
| 10  | [Observability](./10-observability/README.md) | ☐      |
| 11  | [IaC tooling](./11-iac-tooling/README.md)     | ☐      |
| ★   | [End-to-end projects](./projects/README.md)   | ☐      |

## How this repo is organized

```bash
NN-topic/
├── README.md         polished notes, the public-facing teaching content
├── PLAN.md           working roadmap — what's planned, what's done, what's next
├── 01-<concept>/     runnable read-and-learn examples
├── mini-project/     section capstone: real DevOps tool, verified by tests
└── exercises/        starter skeletons with failing tests; reader writes code until they pass
```

See [`PLAN.md`](./PLAN.md) for the overall plan and conventions, and [`SESSIONS.md`](./SESSIONS.md) for the running log of work.

## Using this as an Obsidian vault

1. In Obsidian, **Open folder as vault** → select this repo's root.
2. Obsidian creates a `.obsidian/` folder for vault settings — this is gitignored, so it stays personal.
3. Recommended community plugins: **Dataview** (query notes), **Templater** (note templates).
4. Cross-references inside notes use standard markdown links (`[text](./path.md)`) so they work in both Obsidian and on GitHub.

## Running examples

Prerequisites: Go 1.22+.

```bash
# Run a specific example
go run ./01-cli-tools/01-flag-basics

# Run all tests
go test ./...

# Vet everything
go vet ./...

# Or use the Makefile
make test
make run SECTION=01-cli-tools EX=01-flag-basics
```

## License

MIT — see [LICENSE](./LICENSE).

## Acknowledgements

- [Go by Example](https://gobyexample.com/) — the syntax-reference baseline
- ["The Go Programming Language"](https://www.gopl.io/) by Donovan & Kernighan
- [Effective Go](https://go.dev/doc/effective_go)
