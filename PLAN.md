# Root Plan

High-level roadmap and conventions. Each section folder has its own `PLAN.md` with the detailed buildout for that domain — keep this file lean.

## Goal

Build a hands-on, DevOps-focused Go learning repo that:

1. Doubles as personal notes and a shareable resource
2. Opens as an Obsidian vault
3. Organizes content by **DevOps problem domain** (CLI, HTTP, AWS, k8s) rather than by Go language feature
4. Connects Go concepts to the user's existing languages (Python, Bash, TypeScript, Java)

## Architectural decisions

| Decision              | Choice                                              | Why                                                                                        |
| --------------------- | --------------------------------------------------- | ------------------------------------------------------------------------------------------ |
| Module strategy       | Single root `go.mod`                                | Simpler than multi-module workspaces; one `go test ./...` runs everything                  |
| Exercise solutions    | Test-driven, no solutions in repo                   | Forces hands-on practice; tests give objective "done" signal                               |
| Per-section structure | Notes + PLAN + examples + mini-project + exercises  | Consistent navigation; section PLAN keeps token cost low when loading into Claude sessions |
| CI                    | GitHub Actions, `go vet` + `go test` on push        | Keeps repo buildable as it grows; free for public repos                                    |
| Obsidian compat       | `.obsidian/` ignored; markdown links over wikilinks | Personal vault settings don't leak; links work on both Obsidian and GitHub                 |

## Per-section template

Every `NN-topic/` folder:

```bash
NN-topic/
├── README.md           # Notes (polished, public-facing teaching content)
├── PLAN.md             # Working roadmap for this section
├── 01-<concept>/       # Runnable example #1 (read-and-learn)
│   ├── main.go
│   └── README.md
├── 02-<concept>/
│   └── ...
├── mini-project/       # Section capstone
│   ├── README.md       # spec
│   ├── main.go         # reader implements
│   └── main_test.go    # tests that must pass
└── exercises/
    ├── README.md
    ├── 01-<exercise>/
    │   ├── starter.go        # skeleton with TODOs
    │   └── starter_test.go   # failing tests
    └── 02-<exercise>/
```

## Section README structure

Every section's `README.md` follows the same shape:

1. **What you'll learn** — bullet list
2. **Mental model from other languages** — explicit Python/Bash/TS/Java analogies
3. **The DevOps angle** — why ops engineers care
4. **Walkthrough** — annotated tour of the examples in this folder
5. **Mini-project** — what to build, what the tests verify
6. **Exercises** — prompts (no solutions; tests are the answer key)
7. **Further reading**

## Section PLAN structure

Every section's `PLAN.md` is brief and tactical:

1. **Concepts to cover** — checklist
2. **Examples to build** — list of subfolder names + 1-line spec
3. **Mini-project spec** — what the capstone tool does, what tests verify
4. **Exercises** — list of starter test files to create
5. **Status** — checkboxes per item; updated as work progresses

## Build sequence

1. ☑ Root foundation (go.mod, .gitignore, Makefile, LICENSE, CONTRIBUTING.md, \_assets/, SESSIONS.md, README.md, PLAN.md)
2. ☑ Bootcamp scaffolding (BOOTCAMP.md, issue templates)
3. ☑ Section skeletons (12 folders × stub PLAN/README/exercises/mini-project)
4. ☑ Flesh out `00-setup/` fully — validates the template end-to-end
5. ☑ CI workflow (`.github/workflows/ci.yml`)
6. ☐ User initializes git, pushes to GitHub (manual)
7. ☐ Iterate: fill in sections in any order, following the `00-setup/` template

## Capstones (`projects/`)

Cross-section capstones live under [`projects/`](./projects/) with their own [`PLAN.md`](./projects/PLAN.md) and [`SESSIONS.md`](./projects/SESSIONS.md). At-a-glance status:

- [x] `kube-events-to-slack` scaffolded · [ ] built (tests green)
- [x] `s3-log-shipper` scaffolded · [ ] built (tests green)
- [x] `gcs-log-shipper` scaffolded · [ ] built (tests green)
- [x] `deploy-bot` scaffolded · [ ] built (tests green)

## Session protocol

- Update [`SESSIONS.md`](./SESSIONS.md) at the end of every Claude session — what was done, what's open, any decisions
- Update each section's `PLAN.md` status checkboxes as items complete
- Update the roadmap table in `README.md` when a section moves from ☐ → ◐ → ☑
