# Session Log

A running log of what each Claude Code session accomplished. Newest entries on top. Each session should append a new entry before ending.

## Entry format

```md
## YYYY-MM-DD — <one-line summary>

**Goal:** what the session set out to do
**Done:**

- bullet list of concrete outputs (files created, sections fleshed out, decisions made)
  **Files touched:** comma-separated list (or "see git log <commit>")
  **Open / next:**
- bullet list of follow-ups for the next session
  **Notes:** any decisions, gotchas, or context worth carrying forward
```

---

## 2026-05-21 — `06-testing/` scaffolded

**Goal:** Flesh out `06-testing/` following the `05-concurrency/` pattern: examples with TODOs, mini-project + tests, exercises with failing tests. Spec is in `06-testing/PLAN.md`.

**Done:**

- 8 example folders, each as a small package + working `_test.go` + concept README (deliberate deviation from earlier sections — for a testing topic, the user reads the canonical form first, then extends via in-file TODOs):
  - `01-basic-test` (Add/Sub + t.Errorf vs t.Fatalf)
  - `02-table-driven` (Reverse + inline slice-of-struct)
  - `03-subtests` (Repeat + t.Run + -run filtering + t.Parallel hint)
  - `04-mock-interface` (Notifier + hand-rolled fake with recording slice + failOn)
  - `05-httptest` (WeatherClient + httptest.NewServer + srv.Client() rationale)
  - `06-testdata` (CSV parser + testdata/good.csv, badscore.csv + t.Helper + t.Cleanup)
  - `07-benchmark` (SumLoop vs SumRange + b.ResetTimer + -benchmem + benchstat note)
  - `08-fuzz` (ParseInt with intentional `"-"` bug + FuzzParseInt vs strconv.Atoi invariant)
- Mini-project `logstats`: kitchen-sink that exercises every example pattern in one place. `Parse` + `FormatRate` (pure, fuzz/bench targets), `Aggregator` (stateful), `Source` interface with `FileSource` + `HTTPSource` impls, `Summarize` composition. `main_test.go` has 14 tests + 1 benchmark + 1 fuzz target + a `TestMain` for suite-level setup, covering all 8 example concepts inline-annotated by example number. Includes `testdata/lines.log` fixture (10 lines, mixed levels).
- 3 exercises with failing tests under `//go:build exercise`:
  - `01-table-tests` (`classify` package): `Classify(score) string` with 5 separate passing Test\* funcs as "before" state; exercise file has an empty `TestClassify_Table` to fill in. Tests assert at least one case present; deleting the separate file is the closer.
  - `02-fake-clock` (`scheduler` package): `Scheduler.ShouldFire/Fire` with the `Now func() time.Time` seam already declared + initialized in `New()` but the two methods still call `time.Now()` directly. Trivial 2-char fix; the lesson is the _thinking_. 5 tests; 1 fails until the swap (others pass coincidentally because real-clock deltas happen to behave). Field had to be pre-declared so `go vet -tags=exercise` doesn't error before the user edits.
  - `03-coverage-gap` (`validate` package): `Validate(Config) error` with ~10 branches. 2 starter tests pass; success criterion is `go test -cover` showing 100% (no automated test failure — README explains the workflow).
- Default `go build ./...`, `go vet ./...`, `go vet -tags=exercise ./...`, `go test ./...` all clean. `go test -tags=exercise ./06-testing/...` fails exactly where expected: mini-project (~13 failures), exercise 01 (1), exercise 02 (1).
- `06-testing/PLAN.md` Status flipped; `README.md` status header updated.

**Files touched:** ~32 new files under `06-testing/` (examples + mini-project + 2 fixtures + exercises). No new go.mod deps — everything uses stdlib `testing`.

**Open / next:**

- User to work through examples 01→08 (each has a small "extend the test" TODO)
- Implement `logstats` until `go test -tags=exercise ./06-testing/mini-project/...` is green
- Implement the 3 exercises in any order
- Walkthrough doc in `06-testing/README.md` (currently stub + comparison table)
- Next scaffolding step (future session): `07-aws/`

**Notes:**

- **PLAN.md deviation, mini-project:** PLAN said "add tests retroactively to `dirsize` + `gh-repo-stats`." Those tests already exist (added during 01 + 03 scaffolding). Built a self-contained `logstats` mini-project instead — wired so every example pattern (01-08) maps to a specific test in `main_test.go`. Documented the deviation in `mini-project/README.md` so future-me doesn't re-add tests to the older projects.
- **PLAN.md deviation, examples:** Earlier sections had `main.go` with TODOs and `_ = foo` underscore tricks to keep them compiling. For 06-testing, examples ship as fully working packages with complete tests — TODOs live INSIDE the test file ("now add a case for X") rather than blocking compilation. Rationale: you can't learn testing by staring at a stub; you learn by reading the canonical form and extending it.
- **Exercise 02 had to pre-declare the `Now` field** in `scheduler.go` so `go vet -tags=exercise` doesn't fail on `s.Now undefined`. The exercise is now "use s.Now() in two places" not "add a field + use it." Same precedent as 03-walk/04-exec underscore tricks.
- **Exercise 02's tests are sneaky about coincidence:** `TestShouldFire_WithinCooldownReturnsFalse` and `_PerNameIsolation` pass even with the broken impl because real `time.Now()` deltas inside the test body happen to satisfy the assertions. Only `TestShouldFire_AfterCooldownReturnsTrue` actually requires the fake clock. That's deliberate — fewer failing tests, but the one that fails is precise. The user fix is still the right one even if only 1 of 5 tests was failing.
- **`TestMain` in mini-project** logs suite duration to stderr — visible under `-v`. The shape is intentional: shows the pattern without doing anything load-bearing.
- **Fuzz test in mini-project** seeds with both valid and invalid inputs (`"]ok"`, `""`) — the invariant tolerates errors. Running `-fuzz=FuzzParse` for 10s with the stubbed `Parse` finds nothing (stub always returns error), so the user should run it again AFTER implementing Parse to make the invariant meaningful.

## 2026-05-20 — `05-concurrency/` scaffolded

**Goal:** Flesh out `05-concurrency/` following the `04-http-servers/` pattern: examples with TODOs, mini-project + tests, exercises with failing tests. Spec is in `05-concurrency/PLAN.md`.

**Done:**

- 8 example folders, each with TODO-style `main.go` + concept README: `01-goroutine-basic` (go f() + WaitGroup), `02-channels` (unbuffered vs buffered + close + comma-ok), `03-select` (multiplex + `time.After` timeout + default branch), `04-waitgroup` (Add-before-go pattern), `05-mutex` (Counter with Lock/Unlock + race detector intro), `06-context-cancel` (cancellable worker loop + `WithCancel`), `07-worker-pool` (jobs/results channels + who-closes-what rule), `08-race-detector` (deliberately racy + fix)
- Mini-project `fanout-ping`: scaffolded into `Check(ctx, client, url) Result` + `Run(ctx, client, urls, concurrency) <-chan Result` + cobra wiring using `signal.NotifyContext`. `main_test.go` has 7 tests: happy path, non-OK status is not an error, transport timeout sets Err, concurrency-peak (uses `atomic.Int32` + CAS to record max in-flight; asserts peak ≤ N and ≥ 2 to catch a serial impl), per-request timeout under parallel load, context-cancellation-propagates (asserts every URL still produces exactly one Result on cancel, and that unscheduled work short-circuits), plus a stub-sanity guard
- 3 exercises with failing tests:
  - `01-rate-limiter` (`bucket` package): token-bucket via buffered channel — `New(capacity, refillEvery)` pre-fills + spawns refiller, `Allow()` non-blocking try-receive, `Wait(ctx)` blocking with ctx/stop arms, `Stop()` closes done chan. 5 tests covering initial fill, refill-over-time, Wait-with-token, Wait-honors-ctx, Wait-returns-ErrStopped-on-Stop
  - `02-broadcast` (`broker` package): fan-out one msg to N subs via per-sub buffered channels + RWMutex + non-blocking sends. 5 tests covering fan-out, order-per-subscriber, Unsubscribe closes channel, Close closes all, concurrent publishers safe
  - `03-pipeline` (`pipeline` package): Source/Square/Sum stages over channels, each with `defer close(out)` + ctx-aware select. 5 tests: end-to-end (1+4+9+16=30), empty input, composability (n^4 via Square∘Square), Source closes output, Square closes output when input closes, Sum returns ctx.Canceled + partial sum on cancel
- All exercise/mini-project tests carry `//go:build exercise` so default `go test ./...` is green
- Verified: `go build ./...`, `go vet ./...`, `go vet -tags=exercise ./...`, `go test ./...` all clean; `go test -tags=exercise ./05-concurrency/...` shows expected failures (5 mini-project + 5+5+5 exercises) with no panics, no deadlocks, no hangs

**Files touched:** ~28 new files under `05-concurrency/` (examples + mini-project + exercises).

**Open / next:**

- User to work through examples 01→08 (fill in TODO blocks)
- Implement `fanout-ping` until `go test -tags=exercise ./05-concurrency/mini-project/...` is green
- Implement the 3 exercises in any order
- Walkthrough doc in `05-concurrency/README.md` (currently stub + comparison table)
- Run `go test -race ./...` on `05-mutex` and `08-race-detector` once filled in — they're the section's payoff
- Next scaffolding step (future session): `06-testing/`

**Notes:**

- `New(capacity, refillEvery)` in 01-rate-limiter pre-fills the bucket synchronously _and_ spawns the refiller goroutine. The pre-fill is intentional — gives an initial burst allowance and means `TestAllow_BucketInitiallyFull` does not depend on goroutine scheduling. Documented in the stub comments.
- `TestCheck_TimeoutSetsErr` and `TestCheck_StubIsErroring` (mini-project) happen to pass against the always-erroring stub. Same precedent as `TestLoadConfig_RejectsEmpty` in 04 — tests stay correct once `Check` is real. `TestEmptyInput` in 03-pipeline similarly passes against the stub (`return 0, nil` matches the assertion) — acceptable.
- `TestRun_RespectsConcurrencyLimit` uses `atomic.Int32` + CAS to track peak in-flight without locks — useful pattern for the user to see in passing.
- `TestRun_ContextCancellationPropagates` asserts every URL still produces _exactly one_ `Result` after cancel (with `Err: ctx.Err()` for the unscheduled ones). This is the design choice the stub TODO documents — alternative is "early channel close on cancel, scrap remaining work" which is also valid; the tests pin the contract.
- 04-waitgroup and 01-goroutine-basic needed `_ = wg.Wait` (not `_ = wg`) underscore assignments — `go vet` flags `_ = wg` as "assignment copies lock value" since `sync.WaitGroup` embeds a `noCopy`. Worth remembering when scaffolding stubs that declare WaitGroups before their TODOs are uncommented.

## 2026-05-20 — `04-http-servers/` scaffolded

**Goal:** Flesh out `04-http-servers/` following the `03-http-clients/` pattern: examples with TODOs, mini-project + tests, exercises with failing tests. Spec is in `04-http-servers/PLAN.md`.

**Done:**

- 7 example folders, each with TODO-style `main.go` + concept README: `01-hello-server` (stdlib mux), `02-chi-router` (chi v5), `03-middleware` (logging/recovery/request-id), `04-graceful-shutdown` (SIGTERM + `srv.Shutdown`), `05-health-endpoints` (livez vs readyz), `06-json-api` (POST + DisallowUnknownFields + MaxBytesReader), `07-webhook-receiver` (HMAC-SHA256 with `hmac.Equal`)
- Mini-project `webhook-runner`: scaffolded into `LoadConfig` / `VerifyHMAC` / `runJob` / `newHandler` / cobra wiring with a `WEBHOOK_SECRET` env var and `--config` YAML flag. `main_test.go` has 8 tests: YAML round-trip, empty-jobs rejected, HMAC happy/bad-prefix/tampered/empty/bad-hex, bad-signature 401, unknown-job 404, exit-code capture (success + failure subtests), output truncation, and graceful-shutdown-drains-in-flight (real `sleep 0.3` subprocess + `srv.Shutdown` returns nil before request finishes)
- 3 exercises with failing tests:
  - `01-rate-limit-middleware`: `Limiter` struct with `Allow(key) bool` (fixed-window, clock-injectable via `l.Now`) + `Middleware(*Limiter)`; 5 tests covering under-limit, over-limit reject, reset-after-window (fake clock), per-key isolation, middleware 429
  - `02-basic-auth`: `BasicAuth(user, pass)` middleware using `crypto/subtle.ConstantTimeCompare`; 5 tests covering valid creds, no header, wrong password, wrong username, inner handler not called on failure
  - `03-request-tracing`: `WithRequestID(idGen, logger)` + `RequestIDFromContext`; 5 tests covering generated ID echoed on response, incoming X-Request-ID honored, ID readable from `r.Context()`, start/end log lines with status, missing ID returns empty string
- All exercise/mini-project tests carry `//go:build exercise` so default `go test ./...` is green; mini-project test file is `//go:build exercise && !windows` because it shells out to `sh -c`
- New deps via `go mod tidy`: `github.com/go-chi/chi/v5 v5.2.5`, `gopkg.in/yaml.v3 v3.0.1`
- Verified: `go build ./...`, `go vet ./...`, `go vet -tags=exercise ./...`, `go test ./...` all clean; `go test -tags=exercise ./04-http-servers/...` shows expected failures (16 in mini-project + 15 across exercises) with no panics
- `04-http-servers/PLAN.md` Status flipped + `README.md` status header updated

**Files touched:** ~25 new files under `04-http-servers/` (examples + mini-project + exercises). `go.mod` / `go.sum` updated.

**Open / next:**

- User to work through examples 01→07 (fill in TODO blocks)
- Implement `webhook-runner` until `go test -tags=exercise ./04-http-servers/mini-project/...` is green
- Implement the 3 exercises in any order
- Walkthrough doc in `04-http-servers/README.md` (currently stub + comparison table)
- Next scaffolding step (future session): `05-concurrency/`

**Notes:**

- Graceful-shutdown test depends on a real `sleep 0.3` subprocess + an 80ms warmup before calling `srv.Shutdown` — should be reliable on a normal laptop, may need bumping on a starved CI. Documented in the mini-project README.
- During scaffolding, the initial generation of `main_test.go` came out as ~200 lines of recursive `os.Create` wrappers — total nonsense. Caught it before commit and rewrote with plain `os.WriteFile`. Worth a sanity-read on auto-generated test scaffolds.
- 02-chi-router and 04-graceful-shutdown needed `_ = http.StatusOK` / `_ = os.Interrupt` underscore assignments so the stub `main.go` files compile when none of the TODO blocks are active — same trick as 02-files-and-os's 03-walk + 04-exec.
- `TestLoadConfig_RejectsEmpty` happens to pass against the always-erroring stub. Acceptable (same precedent as `TestCheckAll_EmptyInput` in 03-http-clients) — the test stays correct once `LoadConfig` is real.

**Goal:** Flesh out `03-http-clients/` following the `02-files-and-os/` pattern: examples with TODOs, mini-project + tests, exercises with failing tests.

**Done:**

- 7 example folders, each with TODO-style `main.go` + concept README: `01-basic-get`, `02-json-decode`, `03-timeouts`, `04-headers-auth`, `05-retry-backoff`, `06-context-cancel`, `07-stream-response`
- Mini-project `gh-repo-stats`: cobra scaffold split into `fetchStats` / `doWithRetry` / `loadCache` / `saveCache` / `writeCSV` / `newRootCmd`. `baseURL` is a parameter so tests inject `httptest.NewServer`. `main_test.go` has 6 tests: happy-path JSON decode, retry on 503, retry on 429, honors `If-None-Match` (304), CSV schema, cache JSON round-trip (including missing-file = empty)
- 3 exercises with failing tests:
  - `01-url-health-check`: `CheckAll(client, urls) []Result`; 4 tests covering mixed statuses (incl. redirect chain), transport error doesn't abort, input-order preservation, empty input
  - `02-pagination`: `ParseNextLink(linkHeader) string` + `FetchAll(client, startURL) ([][]byte, error)`; 5 tests covering link parse (found / no-next / empty), Link-header pagination across 3 pages, error stops pagination, single-page no-link
  - `03-mock-server-tests`: `DoWithRetry(client, req, maxAttempts)`; 5 tests covering happy path, 5xx retry, 429 retry, 4xx no-retry, give up after maxAttempts. Exercise also acts as the "use httptest.NewServer over mocking" lesson
- All exercise/mini-project tests carry `//go:build exercise` so default `go test ./...` is green; `-tags=exercise ./03-http-clients/...` shows expected failures with no panics
- Verified: `go build ./...`, `go vet ./...`, `go vet -tags=exercise ./...`, `go test ./...` all clean
- `03-http-clients/PLAN.md` and `README.md` status updated

**Files touched:** ~25 new files under `03-http-clients/` (examples + mini-project + exercises).

**Open / next:**

- User to work through examples 01→07 (fill in TODO blocks)
- Implement `gh-repo-stats` until `go test -tags=exercise ./03-http-clients/mini-project/...` is green
- Implement the 3 exercises in any order
- Walkthrough doc in `03-http-clients/README.md` (currently stub + comparison table)
- Next scaffolding step (future session): `04-http-servers/`

**Notes:**

- Examples 01/02/03 hit real public endpoints (`httpbin.org`, `api.github.com`, `norvig.com/big.txt`) so they're network-dependent. The mini-project + exercises are hermetic — they spin up `httptest.NewServer` and use no real network.
- `TestCheckAll_EmptyInput` happens to pass against the stub (`len(nil) == 0` matches the assertion). Acceptable — the test is still correct after implementation, and the other 3 tests in that file fail until `CheckAll` is real.
- Initially `TestCheckAll_PreservesInputOrder` panicked with index-out-of-range against the empty stub return; added a `len(got) != 3` guard + `t.Fatalf` before the per-index assertions so all exercise failures are clean.

## 2026-05-20 — `02-files-and-os/` scaffolded

**Goal:** Flesh out `02-files-and-os/` following the `01-cli-tools/` pattern: examples with TODOs, mini-project + tests, exercises with failing tests.

**Done:**

- 7 example folders, each with TODO-style `main.go` + concept README: `01-read-write`, `02-line-scanner`, `03-walk`, `04-exec`, `05-signals`, `06-tar-gz`, `07-atomic-write`
- Mini-project `logrotate`: scaffolded into testable pieces (`rotateOnce` / `gzipFile` / `pruneOld` / `newRootCmd`) + `main_test.go` (5 tests covering first-pass rotation, gzip-on-second-rotation, gzip round-trip, age-based pruning, keep-days=0 no-op). Time is injected into `pruneOld` so tests can pin the clock at 2026-05-20.
- 3 exercises with failing tests:
  - `01-dirdiff`: `Diff(left, right) ([]Entry, error)` with sha256-based comparison; 5 tests covering identical trees, OnlyLeft/OnlyRight, Modified, nested relative paths, missing-root error
  - `02-tail-f`: testable kernel `ReadAppend(*os.File, int64) ([]byte, int64, error)` instead of a polling loop — keeps tests fast; 4 tests covering first read, no-growth, append delta, truncation error
  - `03-pipe-cmd`: `Pipe(io.Reader, ...[]string) ([]byte, error)` running real `cat | tr | sort | wc` subprocesses; gated `//go:build exercise && !windows`; 5 tests
- All exercise/mini-project tests carry `//go:build exercise` so default `go test ./...` is green; `-tags=exercise ./02-files-and-os/...` shows expected failures with no panics
- Verified: `go build ./...`, `go vet ./...`, `go test ./...` all clean
- `02-files-and-os/PLAN.md` and `README.md` status updated

**Files touched:** ~25 new files under `02-files-and-os/` (examples + mini-project + exercises).

**Open / next:**

- User to work through examples 01→07 (fill in TODO blocks)
- Implement `logrotate` until `go test -tags=exercise ./02-files-and-os/mini-project/...` is green
- Implement the 3 exercises in any order (`03-pipe-cmd` is the most subtle — see its README on why `Start` must come before `Wait`)
- Walkthrough doc in `02-files-and-os/README.md` (currently stub + comparison table)
- Next scaffolding step (future session): `03-http-clients/`

**Notes:**

- Two scaffolded `main.go` files (`03-walk`, `04-exec`) needed `_ = root` / `_ = target` underscore assignments to satisfy "declared and not used" — pattern to remember when scaffolding TODO files that declare locals before any `_ =` blanket lines.
- `pipe-cmd` test build tag uses `exercise && !windows` because the test relies on `cat`/`tr`/`sort`/`wc`. macOS/Linux only.
- `pruneOld` has `keepDays <= 0` short-circuit so flag default of 0 means "don't prune" — matches the cobra flag help text.

## 2026-05-20 — `01-cli-tools/` scaffolded

**Goal:** Flesh out `01-cli-tools/` following the `00-setup/` pattern: examples with TODOs, mini-project + tests, exercises with failing tests.

**Done:**

- 6 example folders, each with TODO-style `main.go` + concept README: `01-os-args`, `02-flag-basics`, `03-cobra-hello`, `04-cobra-nested`, `05-env-and-config`, `06-exit-codes`
- Mini-project `dirsize`: cobra scaffold split into `scan` / `sortAndTrim` / `renderText` / `renderJSON` / `newRootCmd` + `main_test.go` (6 tests covering recursive sum, missing path, sort+top, JSON validity, text rendering)
- 3 exercises with failing tests:
  - `01-greplite`: library-shaped `Grep(io.Reader, pattern, Options) ([]Match, error)` with 5 tests (substring, ignore-case, line numbers, empty pattern, no match)
  - `02-envdump`: `Match` + `UnsetMatching(... Unsetter)` with injected unsetter interface for testability; 5 tests
  - `03-multi-subcommand`: `Store` (pure logic in `store.go`) + `cmd.go` cobra wiring; 6 tests against `Store`. Demonstrates the "thin CLI over pure logic" pattern
- Deps added via `go mod tidy`: `cobra v1.10.2`, `viper v1.21.0`
- All exercise/mini-project tests carry `//go:build exercise` so default `go test ./...` is green; `-tags=exercise` exposes failures
- Verified: `go build ./...`, `go vet ./...`, `go test ./...` all clean; `go test -tags=exercise ./01-cli-tools/...` shows expected failures with no panics
- PLAN.md status updated; section README status flipped to scaffolded

**Files touched:** ~28 new files under `01-cli-tools/` (examples + mini-project + exercises). `go.mod` updated for cobra/viper.

**Open / next:**

- User to work through examples 01→06 (fill in TODO blocks in each `main.go`)
- Implement `dirsize/main.go` until `go test -tags=exercise ./01-cli-tools/mini-project/...` is green
- Implement the 3 exercises in any order
- Plan walkthrough doc in `01-cli-tools/README.md` (currently just a stub + comparison table; PLAN.md still has "Concepts documented in README.md walkthrough" unchecked)
- Next scaffolding step (future session): `02-files-and-os/`

**Notes:**

- Caught mid-session: initial exercise tests lacked the `//go:build exercise` tag — would have broken the "default green" convention from 2026-05-20. Fixed before logging.
- Exercise 03 deliberately splits `store.go` (tested) from `cmd.go` (cobra wiring, not unit-tested) to model the "keep business logic out of CLI plumbing" pattern that pays off when wiring HTTP servers in section 04.

**Goal:** Stand up a DevOps-focused Go learning repo that doubles as an Obsidian vault and a self-paced bootcamp.

**Done:**

- Root foundation: `go.mod`, `go.sum` (clean), `.gitignore`, `.editorconfig`, `Makefile`, `LICENSE` (MIT), `README.md`, `PLAN.md`, `CONTRIBUTING.md`, `BOOTCAMP.md`, `SESSIONS.md`, `_assets/README.md`
- Bootcamp scaffolding: `BOOTCAMP.md` + 3 GitHub issue templates (`stuck-on-exercise`, `concept-question`, `improvement`) + Discussions-friendly `config.yml`
- 12 section folders with the standard quartet each: `README.md` (notes), `PLAN.md` (roadmap), `exercises/README.md`, `mini-project/README.md`
- `00-setup/` fully fleshed out as the template validator:
  - Full README with toolchain, CLI commands, file anatomy, project layout
  - 4 working examples: `01-hello-world`, `02-go-run-vs-build`, `03-modules-and-deps` (CLI-walkthrough only), `04-go-env-tour`
  - Mini-project `gostat`: starter + tests behind `//go:build exercise`
  - 3 exercises: `01-tidy-experiment` and `02-static-binary` (CLI walkthroughs), `03-env-explorer` (code exercise)
- CI: `.github/workflows/ci.yml` runs `go mod tidy` verification + `go vet` + `go build` + `go test` on push/PR
- Verified: `go mod tidy`, `go vet ./...`, `go build ./...`, `go test ./...` all pass cleanly

**Architectural decisions made:**

- Single root Go module (not per-section)
- Test-driven exercises with no solutions in repo
- Failing exercise tests are excluded from default `go test ./...` via `//go:build exercise`; run with `-tags=exercise`
- Bootcamp positioning chosen over "personal journal"
- Per-section `PLAN.md` files exist so loading a single section's plan into a Claude session stays cheap on tokens

**Files touched:** ~80 files. See `git status` after init.

**Open / next:**

- **User to run:** `git init && git add . && git commit -m "Initial bootcamp scaffolding" && git remote add origin git@github.com:alialjaffer/golang-lab.git && git push -u origin main`
- After pushing: enable GitHub Discussions on the repo (Settings → Features → Discussions)
- Next learning step: work through `00-setup/` exercises 01 + 02 (CLI walkthroughs) and implement `00-setup/exercises/03-env-explorer/starter.go`, then implement `00-setup/mini-project/main.go` (gostat)
- Next scaffolding step: flesh out `01-cli-tools/` following the `00-setup/` pattern

**Notes:**

- GitHub username: `alialjaffer`
- Go version on user's machine: 1.26.2 (darwin/arm64). `go.mod` declares `go 1.22` as the minimum.
- User's background: Python/Bash/TypeScript/Java; learning Go for DevOps; ~40% through theory; learns by doing
- CI uses `go-version: '1.22'` to match the minimum declared in `go.mod`
