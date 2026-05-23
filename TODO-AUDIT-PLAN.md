# TODO Audit Plan

## Why this exists

`CONTRIBUTING.md` already states the rule: **"Exercises ship as failing test files. Readers write code until `go test` passes. No solutions in the repo."**

In practice some files violate the spirit of that rule by leaving the full solution as commented-out `// TODO:` lines that the student "completes" by deleting `//`. Reference smell:
[`11-iac-tooling/exercises/03-provider-error-handling/diagnostics.go`](./11-iac-tooling/exercises/03-provider-error-handling/diagnostics.go) — 17 `// TODO:` lines that, taken together, are the complete answer.

This plan defines what TODOs should be, where they're allowed, and how to audit existing files.

---

## The principle

A TODO should point at a **decision the student must make**, not at the keystrokes that implement it.

A good TODO answers: *"What do I need to think about here?"*
A bad TODO answers: *"What do I type next?"*

The constraint: rewritten TODOs still have to converge on an implementation that passes the existing test in the same file. They guide; they don't teach by dictation, but they must not mislead.

---

## Anti-patterns to remove

1. **Commented-out solution**
   ```go
   // TODO: if err == nil: return nil.
   // TODO: switch {
   // TODO: case errors.Is(err, os.ErrPermission):
   // TODO:     diags.AddError("permission denied", ...)
   ```
   Student "solves" by uncommenting. Zero thinking required.

2. **Variable-name spoilers**
   ```go
   // TODO: declare var diags diag.Diagnostics
   ```
   Names the exact identifier and type. Student isn't choosing the data structure — they're transcribing.

3. **Sequenced micro-steps that enumerate every line**
   Five TODOs for a five-line function is a recipe, not guidance.

4. **TODOs that repeat what the docstring already says**
   If the package doc and function signature already specify the behavior, the TODO should not restate it line-by-line.

---

## Good patterns

1. **One TODO that names the decision, not the code**
   ```go
   // TODO: return early on nil err.
   // TODO: build diag.Diagnostics, dispatching on the error kind
   //   (permission / missing parent dir / other). Each case AddErrors
   //   with a Summary and an actionable Detail — see package doc for
   //   the required wording.
   ```
   Student still has to pick `errors.Is` vs `==`, write Detail strings, and structure the dispatch.

2. **Pointer to the relevant standard-library doc, not the call**
   ```go
   // TODO: read lines from input. See bufio.Scanner.
   ```
   Better than `// TODO: scanner := bufio.NewScanner(input)`.

3. **Question form for non-obvious design choices**
   ```go
   // TODO: how should this behave when pattern is empty?
   //   The test in greplite_test.go pins the answer — read it first.
   ```

4. **Single-line "// TODO: implement."** when the docstring + signature + tests fully specify the contract.
   This is the Section-01 baseline and it works.

---

## Where TODOs are allowed, and at what density

| Location                    | Allowed TODO style                                 | Notes |
| --------------------------- | -------------------------------------------------- | ----- |
| `examples/*`                | Free-form; examples are *meant* to be read         | Examples are not exercises. They can show full code with explanatory comments. |
| `exercises/*` (failing)     | Decision-level only; ≤ 1–2 TODOs per function      | The test file is the spec. The TODO names what to *think about*, not what to type. |
| `mini-project/*`            | Decision-level; can have a few more given scope    | Mini-projects are larger. A TODO per logical milestone is fine; per-line is not. |
| `projects/*` (capstones)    | High-level only; one TODO per component at most    | Capstones are integration work. Students should be designing, not filling blanks. |

---

## Audit scope

Every Go file under:

- `00-setup/` through `11-iac-tooling/` — subdirs `exercises/` and `mini-project/`
- `projects/` (capstones: `deploy-bot/`, `gcs-log-shipper/`, `kube-events-to-slack/`, `s3-log-shipper/`)

Excluded: `examples/` subdirs (TODOs there are fine if they exist), `_assets/`, top-level scaffolding.

Quick triage command:

```bash
# Files with > 3 TODOs are the prime suspects.
grep -rlc "// TODO" --include="*.go" \
  00-setup 01-cli-tools 02-files-and-os 03-http-clients 04-http-servers \
  05-concurrency 06-testing 07-aws 07-gcp 08-kubernetes 09-docker \
  10-observability 11-iac-tooling projects \
  | awk -F: '$2 > 3 {print}' | sort -t: -k2 -n -r
```

---

## Per-file checklist

For each file flagged by the triage above, decide:

1. **Does the docstring + function signature + test file already specify the behavior?**
   If yes → collapse to `// TODO: implement.` (or a one-line nudge).

2. **If multiple TODOs remain, does each one name a *decision* (which API, which error kind, which data structure) rather than a *keystroke* (which variable name, which exact call)?**
   If no → rewrite at the decision level.

3. **Would a student who uncomments every TODO have a passing solution without thinking?**
   If yes → the TODOs are too literal; reduce them.

4. **Do the rewritten TODOs still steer toward a solution the existing tests accept?**
   The test file is authoritative — TODOs must not contradict it or lead the student down a path the tests reject.

---

## Process

1. Run the triage command. Produce a ranked list of offenders.
2. Pick the worst N files (start with > 5 TODOs).
3. For each: read the test file first, then rewrite the TODOs to decision-level. Keep the failing skeleton failing.
4. Verify each rewritten exercise still compiles and its tests still fail in the expected way:
   ```bash
   go test -tags=exercise ./path/to/exercise/
   ```
5. Commit per section, not per file — easier to review.

---

## Out of scope

- Rewriting the actual solutions. The student writes those. This plan only touches the *guidance* left in the failing skeleton.
- Examples in `examples/` subdirs. Those are reference material, not exercises.
- `README.md` / `PLAN.md` content — separate concern.
