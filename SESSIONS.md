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

## 2026-05-23 — `07-gcp/` scaffolded (parallel sibling to `07-aws/`)

**Goal:** Flesh out `07-gcp/` per `07-gcp/PLAN.md`. PLAN was already detailed (concepts list, 7 example folders, mini-project `gcssync`, 3 exercises). Section is intentionally self-contained — a student who skipped 07-aws should be able to do this one without losing context, and a student who did 07-aws should map the foundational concepts (creds chain, pagination, interface-based mocking) across both clouds.

**Done:**

- 7 example folders, each with TODO-style `main.go` + concept README:
  - `01-adc-and-project` — `google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")` walks the ADC chain (env var → gcloud user creds → metadata server); project ID is documented as the *separate, parallel* 4-way chain (env / JSON key / gcloud config / metadata) since GCP has no account-ID-baked-into-creds analogue
  - `02-gcs-list` — `storage.NewClient(ctx)` + `client.Buckets(ctx, project)` + `bucket.Objects(ctx, &storage.Query{Prefix: ...})`; the canonical `it.Next()` → `iterator.Done` shape; explains why GCP-Go uses a sentinel (composes with `errors.Is` + ctx-cancel) vs AWS's `HasMorePages()`
  - `03-gcs-upload-download` — `obj.NewWriter(ctx)` (Close() is what commits — the load-bearing footgun) + `obj.NewReader(ctx)` (lazy; HTTP fires on first Read, NOT on construction — second footgun); the type asymmetry vs AWS's `Body io.Reader` input field
  - `04-gcs-signed-url` — V4 (`Scheme: storage.SigningSchemeV4`) signed GET URL with 5-minute expiry; documents both signing-identity paths (JSON key auto-loaded vs IAM signBlob fallback via `GoogleAccessID`) since path 2 is the production-common one (no JSON key sprawl)
  - `05-gce-list` — `compute.NewInstancesRESTClient(ctx)` + `AggregatedList(ctx, req)` returning per-zone-map; the empty-zone-skip pattern (`if pair.Value == nil || len(pair.Value.Instances) == 0 { continue }`); the `labels.env=prod` filter DSL is GCP-wide (Compute, Logging, BigQuery, Asset Inventory)
  - `06-impersonate-sa` — `impersonate.CredentialsTokenSource(ctx, impersonate.CredentialsConfig{TargetPrincipal, Scopes, Lifetime})` + `option.WithTokenSource(ts)` on the storage client; documents the `roles/iam.serviceAccountTokenCreator` binding being on the target SA (not project-wide); positions impersonation as the JSON-key replacement
  - `07-mocking-gcs` — **ships fully working.** This is the load-bearing one. `*storage.Client` chains through concrete types (`*BucketHandle`, `*ObjectHandle`, `*Reader`, `*Writer`) with unexported fields, so you CAN'T write an interface that satisfies it directly the way `*s3.Client` works in 07-aws. The fix is a thin **adapter wrapper**: `GCSAPI` (3 methods — Read/Write/List) lives in your code; `RealGCS` drives `*storage.Client`; `fakeGCS` in tests is hand-rolled. Two consumers (`FetchKey`, `TotalSize`) demo the pattern; 8 tests cover them. `go test ./07-gcp/07-mocking-gcs/` passes.
- Mini-project `gcssync`: mirror local dir → GCS bucket. Three-method `GCSAPI` (List/Upload/Delete) reused from 07-mocking-gcs's doctrine. CRC32C with the **Castagnoli polynomial** is the load-bearing GCS quirk — `crc32.MakeTable(crc32.Castagnoli)`, NOT the default IEEE table. 10 tests behind `//go:build exercise`: `TestWalkLocal_FlattensTreeWithForwardSlashKeys` (also pins Castagnoli), `TestListRemote_BuildsKeyedMap`, `TestPlan_*` (3 — upload/skip/omit rules, --delete inclusion, deterministic-order), `TestSync_HappyPath`, `TestSync_DryRunMakesNoMutatingCalls`, `TestSync_RespectsConcurrencyLimit` (atomic.Int32 + CAS to record peak in-flight from inside `fakeGCS.Upload`, blocks uploads on a hold channel until peak reaches N), `TestSync_DeleteRemovesStaleKeys`, `TestSync_PropagatesError`, `TestSync_CRC32CMatchSkipsUpload` (the load-bearing GCS test — same CRC → skip, not re-upload), `TestCRCHelperUsesCastagnoli` (regression catcher — guards against the default IEEE polynomial accidentally creeping in), `TestComputeCRC32C_StreamMatchesChecksum`.
- 3 exercises with failing tests under `//go:build exercise`:
  - `01-bucket-inventory` (`inventory` package): `Inventory(ctx, GCSAPI, project) ([]Row, error)` + `WriteCSV(w, rows)`. `GCSAPI` here = `ListBuckets` + `ListObjects` (drained, no iterator). 5 tests: empty project, flatten buckets×objects, preserve ListBuckets order (no sorting), fail-fast on per-bucket error, CSV header+RFC3339 timestamps.
  - `02-find-unlabeled` (`unlabeled` package): `FindUnlabeled(ctx, ComputeAPI, project, requiredKey) ([]string, error)`. `ComputeAPI.AggregatedListInstances` returns a pre-flattened slice (wrapper hides the per-zone-map). 6 tests: flags-missing-key, empty-value-counts-as-present (the key existing is what matters), preserve-API-order across zones, all-labeled → empty, validation on empty requiredKey, error propagation.
  - `03-cleanup-old` (`cleanup` package): `Cleanup(ctx, GCSAPI, bucket, prefix, cutoff, dryRun) ([]string, error)`. `GCSAPI` here = `ListObjects` + `DeleteObject`. 6 tests: deletes-only-old, dry-run-skips-delete, prefix-narrows-server-side (fake honours prefix to simulate GCS), list-error-aborts, delete-error-returns-partial-list, validation on empty bucket.
- All exercise/mini-project tests carry `//go:build exercise`; default `go test ./...` stays green (all 13 `07-gcp` runtime packages show `[no test files]` except `07-mocking-gcs` which passes its 8 tests).
- Verified: `go build ./...`, `go vet ./...`, `go vet -tags=exercise ./...`, `go test ./...` all clean. `go test -tags=exercise ./07-gcp/...` shows expected failures: mini-project (11), exercise 01 (4 explicit + 1 coincidence pass — `TestInventory_EmptyProject` passes because the stub returns nil which is correctly an empty result), exercise 02 (5 explicit + 1 coincidence pass — `TestFindUnlabeled_AllInstancesLabeled` passes because nil result happens to match), exercise 03 (5). No panics, no hangs.
- `07-gcp/PLAN.md` Status flipped (all four boxes — including "Concepts in README walkthrough" since the walkthrough was written this session).
- `07-gcp/README.md` written: hook, what-you'll-learn bullets, mental-model table (Go-GCP vs Python vs Go-AWS for sibling reference), DevOps angle (Castagnoli polynomial + iterator pattern + AggregatedList empty zones + impersonation > JSON keys + Writer.Close commits), Walkthrough 1-7 with one paragraph per example, mini-project hook with tests pinned, exercises list with WHY each one exists, further reading (Cloud client libraries index, godoc, ADC guide, iterator package, impersonation docs, V4 signed URLs spec). Matches the shape of the other 11 section READMEs.
- **No CI bump.** Go 1.26.0 from prior session still holds. GCP deps added (`cloud.google.com/go/storage` v1.62.2, `cloud.google.com/go/compute` v1.64.0, `cloud.google.com/go/iam` v1.7.0, `google.golang.org/api` v0.280.0 with `iterator`/`option`/`impersonate` subpackages, `golang.org/x/oauth2/google`) + ~12 transitives (auth, gax-go, s2a-go, enterprise-certificate-proxy, GoogleCloudPlatform/opentelemetry-operations-go, etc.). go.mod/go.sum updated; some minor x/* version bumps as a side effect of `go get`.

**Files touched:** ~28 new files under `07-gcp/` (examples 01-07 each = main.go + README.md; mini-project = main.go + main_test.go + README.md; exercises/{01,02,03} each = package.go + package_test.go + README.md; exercises/README.md; section README.md; PLAN.md edited). `go.mod`, `go.sum` updated.

**Open / next:**

- User to work through examples 01→06 (each TODO-block in `main.go`)
- Implement `gcssync` until `go test -tags=exercise ./07-gcp/mini-project/...` is green
- Implement the 3 exercises in any order
- Root `README.md` roadmap table currently shows one row for "07 — AWS SDK" — deliberately *not* touched this session since splitting it into 07a/07b or adding a sibling row is a structural-presentation decision. Worth deciding: add a row "07-gcp — GCP SDK"? Rename to "07 — Cloud SDKs (AWS, GCP)"? Keep as-is and let directory discovery do the work? Whatever the user picks, the section README + PLAN already cross-reference 07-aws as the sibling.
- Examples 01-06 require a GCP project + ADC creds to actually run; they compile and lint clean without. Example 05 (`gce-list`) also needs Compute Engine API enabled on the project (`gcloud services enable compute.googleapis.com`). Example 06 (`impersonate-sa`) additionally needs a second SA + a `roles/iam.serviceAccountTokenCreator` binding — documented in its README.
- The mini-project's `cmd.RunE` wires the real GCSAPI as a TODO — once `Sync` etc. pass tests, the user adds `gcsutil.NewReal(ctx)` (from 07-mocking-gcs) to make the binary runnable end-to-end.

**Notes:**

- **The GCP-specific testing wrinkle is the headline of this section.** AWS SDK v2's methods live directly on `*s3.Client`, so a 1-method interface satisfies it for free. GCS chains through concrete types (`*BucketHandle` → `*ObjectHandle` → `*Reader`) with unexported fields, so you can't pick off methods piecemeal. The `RealGCS` adapter pattern lives entirely in user code; the SDK doesn't expose any seam. README 07 explicitly pins this contrast in a table.
- **Castagnoli vs IEEE CRC32 is the load-bearing production fact.** The default `hash/crc32.ChecksumIEEE` will compute a different hash than GCS's server-side CRC32C, so a naive `gcssync` re-uploads every file every run forever. `TestCRCHelperUsesCastagnoli` exists explicitly to regression-catch this — if a future student substitutes `crc32.ChecksumIEEE` for the Castagnoli table, the test fails. Documented in the mini-project README and in section README's DevOps angle.
- **`AggregatedList` empty-zone skip** is the equivalent of AWS's Reservations-wrapping-Instances papercut — every cloud SDK has one. `len(pair.Value.Instances) == 0` is the canonical guard. Documented in example 05 and the section README.
- **The 4-way project-ID chain (env / JSON-key / gcloud / metadata)** is mostly invisible — clients resolve it for you — but it's worth knowing because tooling that wants "the project I'm operating in" needs to walk the same chain. `compute.NewInstancesRESTClient(ctx)` takes project per-call; `storage.NewClient(ctx)` doesn't need one because bucket names are globally unique. Example 01 pins this.
- **Impersonation is positioned as the JSON-key replacement.** The 06-impersonate README explicitly recommends `iam.serviceAccountTokenCreator` over key files for new deployments — short-lived tokens beat long-lived keys for credential-sprawl reasons. Same recommendation Google's own docs make.
- **Self-contained framing in section README.** First paragraph reads "If you did `07-aws` first, you'll find the foundational concepts faster — they map cleanly across clouds, only the surface syntax differs. If you skipped `07-aws`, nothing here assumes it." Same framing in PLAN.md prologue. The mental-model table includes a Go-AWS column (3rd) for the user who did 07-aws first; the Python column (2nd) is enough for the cross-language user.
- **Tests don't hit GCP.** Every test uses a hand-rolled fake satisfying the local `GCSAPI`/`ComputeAPI` interface. The fakes drain iterators into slices so test setup is just `map[string][]ObjectAttrs{...}` instead of "construct a `*storage.ObjectIterator` from the outside" (which you can't). The 8 ships-working tests in `07-mocking-gcs` use the same pattern — that's the demonstration consumers will copy.
- **Exercise 02 reuses the wrapper pattern.** `ComputeAPI.AggregatedListInstances` returns a flat `[]InstanceSummary`, not the raw per-zone map. This is intentional — the GCP iterator-of-zones pattern is taught in example 05; exercise 02 lets the student work with the post-wrapper shape, which is what production code actually does. Documented in exercise 02's README.

---

## 2026-05-22 — README walkthroughs fleshed out for sections 01-11

**Goal:** Per the prior session's closing note, the bootcamp is structurally complete but the per-section README walkthroughs across sections 01-11 were stubs (23-31 lines each, just headers + "see PLAN.md"). User selected "Write README walkthroughs across all sections" as the next step. The reference was `00-setup/README.md` (already fleshed out at ~229 lines).

**Done:**

- Wrote full walkthroughs for all 11 sections (01-cli-tools through 11-iac-tooling) matching the `00-setup/README.md` shape: hook paragraph, "What you'll learn" bullets, "Mental model from other languages" table (expanded vs the original stubs — added rows for the concepts that are actually load-bearing in each section), "The DevOps angle" with non-obvious production details rather than just the pitch, numbered Walkthrough section with one paragraph per example (linking to each subfolder), Mini-project paragraph with the pinned-contract one-liners pulled from prior SESSIONS notes, Exercises section with brief WHY-this-exercise-exists rather than just the title, Further reading section with stable canonical links (stdlib godocs, language-team posts, official tutorials).
- The walkthroughs pull the load-bearing-detail content from the prior SESSIONS.md entries (07-aws, 08-kubernetes, 09-docker, 10-observability, 11-iac-tooling all had rich notes already). For sections 01-06 (which were scaffolded in earlier sessions without comparable detail in SESSIONS.md), drew on the PLAN.md tables + each section's existing stub + general DevOps-Go knowledge — sticking to widely-accepted patterns, no fabricated APIs.
- Each README links to: the section's own subfolders (examples + mini-project + exercises) by relative path, the relevant stdlib godoc, and 1-2 canonical external sources (Cloudflare timeouts post, AWS jitter post, k8s sample-controller, OTel semantic conventions, HashiCorp framework docs, etc.).
- Cross-references between sections where they earn it: section 10's mini-project is documented as building on section 04's `webhook-runner`; section 08's mini-project Deduper is documented as reusing the `Now func() time.Time` pattern from `06-testing/02-fake-clock`; section 07's "interface-at-consumption-site" doctrine is referenced from sections 08/09/10 mini-projects' fakes.
- 11-iac-tooling's README explicitly documents the Pulumi-skipped decision and notes exercise 02 was skipped for the same reason (so future-me reading the README doesn't try to find an exercise that doesn't exist).
- 00-setup intentionally left untouched (already complete).
- No PLAN.md status boxes were updated as part of this session — the user can tick `[ ] Concepts documented in README walkthrough` → `[x]` themselves if they want (deferred since the spec wording varies slightly across PLANs and I didn't want to silently rewrite them).

**Files touched:** 11 files — `{01-cli-tools,02-files-and-os,03-http-clients,04-http-servers,05-concurrency,06-testing,07-aws,08-kubernetes,09-docker,10-observability,11-iac-tooling}/README.md`. No code changes, no go.mod / go.sum / CI changes.

**Open / next:**

- User decides whether to tick the "Concepts documented in README walkthrough" status boxes in each PLAN.md (left untouched in this session)
- User to actually work through the TODOs / exercises / mini-projects across the suite (the multi-month phase the bootcamp was designed for)
- If/when sections grow (e.g., a new example is added), the relevant README's Walkthrough section needs the new item appended — the structure is now in place for that to be a one-line edit
- If the user disagrees with framing in any section's README (the "What you'll learn" angle, the cross-references, the further reading picks), point it out and I'll adjust; the shape is consistent across all 11 so a single style decision propagates

**Notes:**

- **Style decisions pinned across all 11 READMEs** (for future consistency):
  - Status header uses ☑ (boxed-check) for "scaffolded and walkthrough done" — same glyph the existing stubs used
  - "What you'll learn" is a bullet list, not numbered (numbering is reserved for the walkthrough order which is genuinely sequential)
  - "Mental model from other languages" tables consistently list Go first, then Python, then TS/Node, then (when relevant) Bash and Java. Order matches `00-setup`.
  - "The DevOps angle" sections lead with the production-grade non-obvious details, not the "Go has a single static binary!" pitch (that's beaten to death in `00-setup` already)
  - Walkthrough numbered list links to each subfolder with a one-paragraph teaser pulling out the load-bearing concept; the example's own README is where the full explanation lives
  - Mini-project section is short — a hook paragraph + a "tests pin: ..." sentence pulling the contract-bearing tests out, then "spec in mini-project/"
  - Exercises section is one bullet per exercise with a brief WHY (the pattern it teaches), not a restatement of the spec
  - Further reading is 4-5 stable canonical links per section; avoid blog posts that might rot
- **Section 10's "trace_id in logs requires the span to be sampled" callout** is the kind of detail that would be hard to learn outside this README — preserved verbatim from the prior session notes because it's load-bearing for anyone actually wiring this up
- **Section 04's `/healthz` vs `/readyz` framing** ("If your DB is down, the k8s liveness probe restarts your pod, then your replacement also can't reach the DB...") is the cascade-fail story; it's the WHY the K8s docs gloss over and the WHY that justifies the deliberate split in example 05
- **Section 11 README is the longest** (intentionally — it covers the most novel material; the framework's typed-value model has no analog in any other ecosystem)
- The TodoList shows 12/12 tasks complete (the 11 sections + the SESSIONS.md update). Total ~30k tokens of new content across the 11 READMEs.

---

## 2026-05-22 — `11-iac-tooling/` scaffolded (Terraform-only; Pulumi skipped per user)

**Goal:** Flesh out `11-iac-tooling/` following the `10-observability/` pattern: examples with TODOs, mini-project + tests, exercises with failing tests. Spec is in `11-iac-tooling/PLAN.md`. User instruction mid-session: skip all Pulumi content (originally PLAN had `03-pulumi-hello/` + exercise `02-pulumi-loop`).

**Done:**

- 2 example folders, each with TODO-style `main.go` + concept README:
  - `01-tf-provider-skeleton` (the minimum-viable provider against `terraform-plugin-framework` v1.19.0: `provider.Provider` interface = Metadata/Schema/Configure/Resources/DataSources; `resource.Resource` interface = Metadata/Schema/Create/Read/Update/Delete; `providerserver.Serve` is the entry point that speaks gRPC over stdin/stdout to the `terraform` CLI; the model uses `tfsdk:"..."` reflection tags + `types.String` ternary value type for null/unknown/value semantics; Read/Update/Delete are stubs because the `echo_value` resource has no remote state — Create just copies input → output)
  - `02-tf-provider-crud` (full CRUD on the same `echo_value` resource; introduces validators via `terraform-plugin-framework-validators` `stringvalidator.LengthAtLeast(1)` + plan modifiers via `stringplanmodifier.RequiresReplace()` and `UseStateForUnknown()`; computed `length` attribute the provider derives; stable sha1-prefix `id`; the canonical refresh pattern in Read = state→model→remote-API→state; the three-values doctrine (Config/Plan/State, when each is the source of truth); local-install + `terraform apply` walkthrough in README)
- Mini-project `tf-provider-fileops`: a `fileops_templated_file` resource that renders Go `text/template` + vars and writes the result to a local path. Scaffold split into `main.go` (provider + resource + model + CRUD TODO stubs) + `helpers.go` (4 pure I/O helpers — `RenderTemplate` / `WriteTemplatedFile` / `ReadTemplatedFile` / `DeleteTemplatedFile`) + `main_test.go`. 10 tests total: 3 for `RenderTemplate` (happy path, `missingkey=error` is the pinned production stance vs default `<no value>` substitution, invalid syntax errors), 2 for `WriteTemplatedFile` (creates + overwrites), 2 for `ReadTemplatedFile` (happy + `errors.Is(err, os.ErrNotExist)` survives — the regression-catcher for drift detection), 2 for `DeleteTemplatedFile` (removes + idempotent-on-missing — terraform destroy must be re-runnable), 1 provider-smoke (`TestProviderSchemaCompiles` reads `p.Resources(ctx)` len), 1 acceptance-test stub `TestAccTemplatedFile_SkipsWithoutTF_ACC` (skips without `TF_ACC=1`; full `terraform-plugin-testing` wiring documented in README as extension exercise).
- 2 exercises with failing tests under `//go:build exercise` (exercise `02-pulumi-loop` skipped per user):
  - `01-tf-data-source` (`fileinfods` package): implement a read-only data source `fileops_file_info` (`path` Required, `exists` Computed bool, `size` Computed int64). 5 tests: 3 for `ReadFileInfo` pure helper (existing file → Size+Exists, missing file → `Exists=false, NOT error` — pinned contract since data sources are lenses not guards, empty-file is a valid file), 1 for Metadata TypeName (`<provider>_file_info`), 1 for Schema completeness (all 3 attributes present).
  - `03-provider-error-handling` (`provdiag` package): `WriteFileDiagnostics(path, err) diag.Diagnostics` that translates raw `os` errors into actionable user-facing diagnostics. 5 tests: nil-err → empty diags, `os.ErrPermission` → mentions chmod/chown/privileges + path, `os.ErrNotExist` → mentions create-dir/mkdir, generic err → includes underlying error string + path, all-severities-are-`SeverityError` (no warnings — anything that prevented a write IS an error).
- All exercise/mini-project tests carry `//go:build exercise`; default `go test ./...` stays green (all 5 `11-iac-tooling` runtime packages show `[no test files]`).
- Verified: `go build ./...`, `go vet ./...`, `go vet -tags=exercise ./...`, `go test ./...` all clean. `go test -tags=exercise ./11-iac-tooling/...` shows expected failures: mini-project (7 explicit + 4 coincidence-passes — `TestRenderTemplate_MissingVarIsError` and `TestRenderTemplate_InvalidSyntax` pass because the stub returns an error which IS the expected behaviour for those failing cases; `TestProviderSchemaCompiles` passes because the resource factory IS correctly registered in the stub; `TestAccTemplatedFile_SkipsWithoutTF_ACC` skips cleanly), exercise 01 (5 — `Schema_HasThreeAttributes` reports 3 missing attrs in one failure), exercise 03 (3 explicit + 2 coincidence-passes — nil-err vacuously satisfies, all-severities-are-error vacuously satisfies because empty diag slice has no severities to check). No panics, no hangs.
- `11-iac-tooling/PLAN.md` Status flipped (Examples/Mini-project/Exercises ticked; README walkthrough still ☐); `11-iac-tooling/README.md` status header updated. PLAN's Pulumi section replaced with an explicit "skipped per user" note.
- **No Go version bump.** terraform-plugin-framework v1.19.0 + v0.19.0 validators are compatible with go 1.26.0.

**Files touched:** ~15 new files under `11-iac-tooling/` (examples + mini-project + 2 exercises). `go.mod`, `go.sum` updated. New deps: `github.com/hashicorp/terraform-plugin-framework` v1.19.0, `github.com/hashicorp/terraform-plugin-framework-validators` v0.19.0, plus ~15 transitive (`terraform-plugin-go` for tfprotov6/tfprotov5/tftypes, `terraform-plugin-log` for tflog/tfsdklog, `terraform-registry-address`, `terraform-svchost`, `go-uuid`, `yamux`, `oklog/run`, `protocompile`, `jhump/protoreflect`, `fatih/color`, `mattn/go-isatty`, `go-colorable`, `vmihailenco/tagparser/v2`).

**Open / next:**

- User to work through examples 01-02 (each TODO-block in `main.go`); example 02 is the bigger one (CRUD + validators + plan modifiers)
- Implement `tf-provider-fileops` until `go test -tags=exercise ./11-iac-tooling/mini-project/...` is green
- Implement the 2 exercises in any order
- Walkthrough doc in `11-iac-tooling/README.md` (currently stub + comparison table)
- For acceptance tests against a real `terraform` binary, install Terraform locally and run `TF_ACC=1 go test -tags=exercise -run TestAcc ./11-iac-tooling/mini-project/...` after wiring `terraform-plugin-testing` per the mini-project README extension section
- All examples + mini-project compile without `terraform` installed. `terraform apply` against a built provider binary requires the binary be moved to `~/.terraform.d/plugins/registry.terraform.io/examples/<name>/<version>/<os>_<arch>/` — documented in each example's README
- **The bootcamp is now structurally complete.** All 11 sections (00-setup through 11-iac-tooling) have examples + mini-project + exercises scaffolded. The remaining work is per-section README walkthroughs (currently stubs + comparison tables in most sections) and the user actually implementing the TODOs/exercises across the suite.

**Notes:**

- **`terraform-plugin-framework` v1 vs `terraform-plugin-sdk/v2`:** the framework is the modern API HashiCorp directs new providers to use — it has cleaner ergonomics (typed `types.String`/`types.Int64`/etc with proper null/unknown semantics, separate Schema vs Plan vs State accessors, structured diagnostics with Summary+Detail). The legacy `sdk/v2` ("Schema SDK") is still around for existing providers but new code should default to framework. Documented in 01's README.
- **The `tfsdk:"..."` reflection tag** is the framework's bridge between Go structs and Terraform values. The framework uses runtime reflection to marshal between `types.String`/`types.Bool`/etc and your struct fields. This is why `Path types.String` rather than `Path string` — Terraform values have a ternary state (value/null/unknown), and a plain `string` can't represent "unknown" (known-after-apply).
- **`providerserver.Serve` does the gRPC dance.** When `terraform apply` runs, the CLI exec()s the provider binary, then connects to its stdin/stdout over the HashiCorp go-plugin protocol. The `Address` field declares where the provider would live in the registry — locally you install at `~/.terraform.d/plugins/<address>/<version>/<os>_<arch>/`. Both examples + the mini-project's README spell out the install path.
- **Validators run during `terraform plan`,** before Create/Update. The user sees a useful error message at plan time instead of the Create call panicking against a bad value. Example 02 demonstrates `stringvalidator.LengthAtLeast(1)` on `input`. Validators are pure-validation; for cross-attribute or stateful checks, use the provider's `Validators` schema slot or `Resource.ValidateConfig` hook.
- **Plan modifiers run AFTER validators but BEFORE the CRUD handler.** They shape what the diff *looks like* to the user. `RequiresReplace()` turns an in-place Update into a Destroy+Create (what AWS calls "RECREATE_AUTO" in CloudFormation). `UseStateForUnknown()` prevents the "(known after apply)" placeholder in plan output when an attribute is computed but stable across applies — without it, every plan shows `id` as unknown, which clutters output.
- **Mini-project pinned production stance: `Option("missingkey=error")`.** Default Go `text/template` silently substitutes `<no value>` for a missing variable. That's fine for log messages, but an IaC tool that writes the result to a file Terraform tracks forever should fail loudly on typo'd vars — `listen: ":<no value>"` shipping to `/etc/myapp/config.yaml` would be a real bug. Documented in `helpers.go` and the mini-project README.
- **`errors.Is(err, os.ErrNotExist)` survives `os.ReadFile`'s wrapping** — the test `TestReadTemplatedFile_NotFoundReturnsErrNotExist` exists specifically to regression-catch a future refactor that swaps `os.ReadFile` for a custom helper that loses the sentinel. The resource's `Read` handler depends on the sentinel to call `resp.State.RemoveResource(ctx)` for drift detection; if the sentinel disappears, drift detection silently breaks.
- **`DeleteTemplatedFile` is idempotent** — terraform destroy is re-runnable. If a previous destroy was interrupted, or the file was manually removed, the second destroy must succeed. This is the universal pattern in Terraform providers; AWS providers do the same (DELETE on a 404 resource returns success).
- **Acceptance tests gated on `TF_ACC=1`** is the HashiCorp convention. The pattern: `if os.Getenv("TF_ACC") == "" { t.Skip(...) }` at the top of every `TestAcc*` function. The mini-project ships a stub of this shape (no real terraform-plugin-testing wiring) because (a) it adds a heavy dep, (b) the user might not have `terraform` installed yet, (c) the pure helpers exercise the interesting logic anyway. Extension section in the README shows the full wiring.
- **Data sources are read-only siblings of resources** — same Schema, only `Read` (no Create/Update/Delete). Exercise 01's `fileops_file_info` data source demonstrates the convention. Crucially: a missing file should set `Exists=false`, NOT return an error. Data sources are lenses on real-world state, not guards — the user composes them into `count = data.x.exists ? 1 : 0` patterns. Pinned in `TestReadFileInfo_MissingFileIsNotAnError`.
- **Actionable diagnostics matter more than detailed ones.** Exercise 03 pins the contract: `Summary` should be a short noun phrase ("permission denied"), `Detail` should include the path AND a remediation hint (chmod / mkdir / "run with sufficient privileges"). All three error types in the exercise must produce `SeverityError`, never `Warning` — warnings would let a user merrily apply and then wonder why nothing happened. Documented in the exercise README.
- **The bootcamp is structurally complete** — 11 sections × (examples + mini-project + exercises) all scaffolded. The next phase is README walkthroughs across sections (most are currently stub + comparison table) and the user actually completing the implementations. SESSIONS.md should keep getting entries as the user works through them, but the "scaffold the next section" loop has reached its natural end.

---

## 2026-05-22 — `10-observability/` scaffolded

**Goal:** Flesh out `10-observability/` following the `09-docker/` pattern: examples with TODOs, mini-project + tests, exercises with failing tests. Spec is in `10-observability/PLAN.md`.

**Done:**

- 6 example folders, each with TODO-style `main.go` + concept README:
  - `01-slog-basics` (`slog.NewTextHandler` / `slog.NewJSONHandler`, levels-live-on-the-handler, `slog.Group`, `AddSource: true`, "why slog killed zap/logrus for new code")
  - `02-slog-context` (the `WithLogger`/`FromContext` pattern via an unexported typed `ctxKey struct{}`; `FromContext` falls back to `slog.Default()` so callers never nil-check; `logger.With(...)` is a builder, not mutation)
  - `03-prom-counter` (`promauto.NewCounterVec` + `promhttp.Handler()` + the canonical RED `http_requests_total{method,path,status}`; the cardinality-trap README section is the headline — never label by user_id/raw URL)
  - `04-prom-histogram` (`HistogramVec` + `ExponentialBuckets(0.001, 2, 14)` + the `time.Since(start).Seconds()` + `defer Observe(...)` pattern + Prometheus convention "seconds, not millis"; README contrasts Histogram vs Summary — "use Histogram, period")
  - `05-otel-tracing` (the three things — exporter / TracerProvider / Tracer; `tracer.Start(ctx, name)` returns BOTH a new ctx AND the span — and you MUST use the returned ctx for children; `stdouttrace.WithPrettyPrint()` for terminal-visible spans; `tp.Shutdown` is not optional — without it the batcher drops the final batch)
  - `06-trace-http` (`otelhttp.NewHandler` server-side + `otelhttp.NewTransport` client-side + `propagation.NewCompositeTextMapPropagator(TraceContext{}, Baggage{})`; the W3C `traceparent` header anatomy; the "global propagator default is no-op" gotcha that breaks 100% of first-time setups)
- Mini-project `webhook-runner-instrumented`: rebuilds 04-http-servers' webhook-runner with full observability surface. `Metrics` struct holds `Requests` (CounterVec method/path/status) + `Duration` (HistogramVec method/path) + `InFlight` (Gauge) + `Jobs` (CounterVec job/result with ok|fail|unknown enum), plus the registry itself so tests can spin up independent servers without `DefaultRegisterer` collisions. `observability(opts)` middleware does the cross-cutting work — generates `request_id` (atomic counter, deterministic for tests), starts the server span, builds a `*slog.Logger` pre-bound with `request_id` AND `trace_id` (the latter pulled from `span.SpanContext().TraceID()`), stashes it on ctx via unexported `ctxKey{}`, increments `InFlight` with defer-Dec, captures status via a `statusRecorder` wrapper, observes duration on End. `VerifyHMAC` and `RunJob` get their own spans (so a trace shows verify→run as a chain). The outer mux serves `/metrics` outside the observability wrap (no point counting scrapes). `main_test.go` has 12 tests + helpers: `TestVerifyHMAC_*` (3), `TestRunJob_*` (3 — success/non-zero-exit/output-truncation, all driven by real `sh -c` subprocesses), `TestServer_HappyPath` (asserts both `http_requests_total{status=200}` AND `webhook_jobs_total{result=ok}`), `TestServer_BadSignatureReturns401AndCounts`, `TestServer_UnknownJobReturns404AndCountsUnknown`, `TestServer_FailingJobCountsFail` (exit≠0 is still HTTP 200 — the test pins this contract: "process failed" is not "transport failed"), `TestServer_MetricsEndpointExposesSeries` (4 series names), `TestServer_RequestLogIncludesRequestID` (asserts `"request_id"` AND `"trace_id"` AND both `request.start`/`request.end` events appear in the bytes.Buffer-backed JSONHandler), `TestServer_SpansCreated` (3 expected spans via `tracetest.NewInMemoryExporter` + `sdktrace.WithSyncer(exp)`), `TestServer_InFlightReturnsToZero` (deferred Dec invariant).
- 3 exercises with failing tests under `//go:build exercise`:
  - `01-log-context-key` (`reqlog` package): `WithRequestID`/`RequestIDFromContext` + `WithLogger`/`LoggerFromContext` + `Middleware(base, idGen)`. 5 tests: round-trip, missing-returns-empty, honours-incoming-`X-Request-ID`-header (does-not-regenerate-over-it AND echoes it back on the response), generates-via-injected-idGen-when-absent, ctx-bound-logger-has-`request_id`-pre-bound. `idGen` is injected just like the clock in `06-testing/02-fake-clock` — tests pass a constant `"GENERATED"` generator; production uses `uuid.New().String`.
  - `02-rate-limited-logging` (`ratelog` package): `Limiter` with injectable `Now func() time.Time` + per-key `lastSeen map[string]time.Time` (mutex-guarded). 5 tests using a fake clock: first-call-allowed, second-within-window-blocked, after-window-allowed-again, per-key-isolation, rate-limited-key-stays-blocked-while-other-keys-free. README mentions the natural follow-up (wrap as a `slog.Handler`) but the exercise itself just pins `Allow()` semantics.
  - `03-trace-sql-call` (`dbspan` package): `Query(ctx, tracer, db, sql)` that wraps a DB call with a span named `db.query`, sets `db.statement` attribute, `db.rows_affected` on success, AND `span.RecordError(err) + span.SetStatus(codes.Error, ...)` on failure. 5 tests: span-named, statement-attribute-set, rows_affected-on-success, error-recorded-AND-status-set (asserts both — production needs both: RecordError adds the event for dashboards, SetStatus(Error) paints the span red and bubbles up to trace-error-rate alerts), returns-underlying-result. Tests use `tracetest.NewInMemoryExporter` + `sdktrace.WithSyncer(exp)` for deterministic span inspection.
- All exercise/mini-project tests carry `//go:build exercise`; default `go test ./...` stays green (all 10-observability runtime packages show `[no test files]`).
- Verified: `go build ./...`, `go vet ./...`, `go vet -tags=exercise ./...`, `go test ./...` all clean. `go test -tags=exercise ./10-observability/...` shows expected failures: mini-project (9 explicit + 3 coincidence-passes — `TestVerifyHMAC_BadSignature` / `TestVerifyHMAC_MissingPrefix` pass because the stub returns false which IS correct for bad sigs; `TestServer_BadSignatureReturns401AndCounts` passes for the same reason; `TestServer_InFlightReturnsToZero` passes because the defer-Dec runs regardless), exercise 01 (4 — `TestRequestIDFromContext_MissingReturnsEmpty` coincidence-passes because the stub returns ""), exercise 02 (4 — `TestAllow_SecondCallWithinWindowIsBlocked` coincidence-passes because the stub always returns false), exercise 03 (5). No panics, no hangs.
- `10-observability/PLAN.md` Status flipped (Examples/Mini-project/Exercises ticked; README walkthrough still ☐); `10-observability/README.md` status header updated.
- **No CI bump.** Go 1.26.0 from last session still holds. All new deps are compatible with that toolchain.

**Files touched:** ~25 new files under `10-observability/` (examples + mini-project + exercises). `go.mod`, `go.sum` updated. New deps: `github.com/prometheus/client_golang` v1.23.2 (+ `client_model`, `common`, `procfs`, `klauspost/compress` transitively), `go.opentelemetry.io/otel/exporters/stdout/stdouttrace` v1.43.0. The other OTel packages (`otel`, `otel/sdk`, `otel/trace/noop`, `otelhttp`, `propagation`) were already pulled in transitively by `k8s.io/client-go` last session — they just got promoted from indirect to direct.

**Open / next:**

- User to work through examples 01→06 (each TODO-block in `main.go`)
- Implement `webhook-runner-instrumented` until `go test -tags=exercise ./10-observability/mini-project/...` is green
- Implement the 3 exercises in any order
- Walkthrough doc in `10-observability/README.md` (currently stub + comparison table)
- Examples 03/04 (`prom-counter`, `prom-histogram`) and 06 (`trace-http`) bind to `:8080` and serve real HTTP — `curl localhost:8080/metrics` and `curl localhost:8080/parent` are the canonical interactions; nothing needs to be installed (no Prometheus server, no Jaeger). Example 05 (`otel-tracing`) writes span JSON straight to stdout — same "no infra" property.
- Next scaffolding step (future session): `11-iac-tooling/` (the last section)

**Notes:**

- **Why a custom registry on the `Metrics` struct, not `promauto`+default registry:** the test suite builds multiple `newServer` instances. With `promauto`, the second `MustRegister` panics ("duplicate metrics collector registration"). Holding `reg *prometheus.Registry` inside `Metrics` and serving it via `promhttp.HandlerFor(m.registry(), ...)` is the cleanest fix — production also benefits when reloads spin up a fresh registry. Same pattern as k8s informer factories: every test gets its own world.
- **`trace_id` in log lines only works if the span is sampled.** The test TracerProvider uses `sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))` which samples everything by default — so every test request's logs have `trace_id`. In production with `ParentBased(TraceIDRatioBased(0.01))`, only ~1% of requests' log lines carry `trace_id`. The middleware code path is the same either way: `if sc := span.SpanContext(); sc.HasTraceID()`.
- **`request_id` as an atomic counter** is deliberately the simplest thing — keeps tests deterministic ("first request gets id=1"). Production should use `uuid.New().String()` or `ulid.Make().String()` (smaller, time-sortable). Documented in mini-project README.
- **The "exit code ≠ 0 returns HTTP 200" contract** in `TestServer_FailingJobCountsFail` is a real design call: a process failure is application-level (the runner did its job — it ran the command and reported the result), not transport-level. HTTP 5xx should only fire when the server itself misbehaves (spawn failed, ctx died, etc.). This split matches `kubectl get pods` exit codes and GitHub Actions exit semantics. The `result=fail` counter is how alerts catch the "all jobs are failing" condition.
- **`/metrics` is served OUTSIDE the observability middleware** — wrapping it would mean every Prometheus scrape inflates `http_requests_total` and skews the duration histogram. The outer mux pattern (instrument `/`, plain-serve `/metrics`) is the canonical fix; alternatives include adding a `path != "/metrics"` early-return inside the middleware (cheaper but less explicit).
- **OTel span attribute name choices matter** — exercise 03's README explicitly pins `db.statement`, `db.rows_affected` etc. as the OpenTelemetry semantic conventions. Backends (Jaeger, Tempo, Honeycomb, Datadog) ship dashboards that auto-recognize these. If you invent your own names, you re-implement the dashboard.
- **`span.RecordError(err) + span.SetStatus(codes.Error, ...)` together** is the pinned contract in exercise 03. RecordError adds an event; SetStatus paints the span red. Doing only one means dashboards see either "a successful span with a stray exception event" (no SetStatus) or "an errored span with no error detail" (no RecordError). Production code must do both.
- **`tracetest.NewInMemoryExporter` + `sdktrace.WithSyncer(exp)`** is the gateway drug for OTel testing. WithSyncer (not WithBatcher) means spans are exported synchronously — no race between `span.End()` and `exp.GetSpans()`. Use this in tests; never in production (synchronous export blocks the request hot path).
- **The `otelhttp` middleware is the only piece of OTel auto-instrumentation we touch** in this section. Other auto-instrumenters (`otelsql`, `otelgrpc`, etc.) follow the same shape — wrap your transport / handler / dialer; they handle span creation + propagation invisibly. Documented in 06-trace-http README.
- **Stdout tracer for examples (`stdouttrace`)** lets the user see span output with zero infra. Production swaps to `otlptracegrpc.New` or `otlptracehttp.New` pointing at a collector. The TracerProvider construction is otherwise identical — only the exporter changes. README 05 calls this out.
- **No prom-summary example.** Summaries are legacy; the consensus in the Prometheus community is "always Histogram, never Summary" for new code (summaries don't aggregate across replicas). PLAN.md mentions Summary in the "concepts" list but the examples deliberately skip it. README 04 spells out why.

---

## 2026-05-22 — `09-docker/` scaffolded

**Goal:** Flesh out `09-docker/` following the `08-kubernetes/` pattern: examples with TODOs, mini-project + tests, exercises with failing tests. Spec is in `09-docker/PLAN.md`.

**Done:**

- 6 example folders, each with TODO-style `main.go` + concept README:
  - `01-connect` (`client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())` + `cli.ServerVersion(ctx)` smoke probe — explains why API negotiation is non-optional)
  - `02-list-containers` (`cli.ContainerList(ctx, container.ListOptions{All: ...})` + the `Names` `/`-prefix gotcha + `filters.NewArgs(filters.Arg("status", "running"))` DSL)
  - `03-pull-and-run` (the 6-step run sequence: `ImagePull` → drain reader → `ContainerCreate` → `ContainerStart` → `ContainerWait` (two-channel pattern) → `stdcopy.StdCopy` → `ContainerRemove`; the "you MUST drain the pull reader" footgun is the headline)
  - `04-logs-stream` (`ContainerLogs(..., {Follow: true})` + the multiplex frame format `[stream code 1B][len 4B][payload]` + the TTY exception that breaks `StdCopy`)
  - `05-exec` (`ContainerExecCreate` → `ContainerExecAttach` (this is what starts it) → `StdCopy` → `ContainerExecInspect` for exit code; the two-phase API + "exit code lives in a separate Inspect call" are the load-bearing concepts)
  - `06-events` (`cli.Events(ctx, events.ListOptions{Filters: f})` two-channel return + the per-event `Type`/`Action`/`Actor` shape + filter-at-source rationale)
- Mini-project `image-pruner`: policy-driven image cleanup. Three policies OR'd (`--untagged`, `--max-age`, `--no-containers`) + `--dry-run` + `--force`. Scaffold split into `Policy` struct, `DockerAPI` interface (3 methods: `ImageList`/`ImageRemove`/`ContainerList`), `Plan(images, containers, policy, now)` (pure), `isUntagged(image.Summary)` helper, `Sync(ctx, api, policy, now, out)` (does the work). `main_test.go` has 14 tests + a `captureRemoveFake` wrapper covering: `isUntagged` (nil tags / `<none>:<none>` placeholder / real tag), `Plan` rules (untagged, max-age, container-ref protection, OR-not-AND, empty policy = no-op, deterministic sort), `Sync` (happy path, dry-run-makes-no-remove-calls, list-error-propagates, remove-error-propagates, force-flag-threads-through-to-RemoveOptions). Uses `atomic.Int32` for remove-call counting + a `captureRemoveFake` to assert the force flag actually reaches `RemoveOptions.Force` (regression catcher).
- 3 exercises with failing tests under `//go:build exercise`:
  - `01-container-stats` (`stats` package): `CPUPercent(curr, prev Snapshot) float64` + `MemoryPercent(s Snapshot) float64`. 6 tests covering the typical-case formula (200ms / 4s / 4 cores → 20%), identical-snapshots (no-time-passed → 0, not NaN), counter-reset (container restart → 0, never negative), first-sample (zero prev → must not return NaN), memory-with-limit (256MiB/1GiB → 25%), memory-without-limit (limit=0 → 0, no division-by-zero). Deliberately split from the streaming-decode part since the math is where people get it wrong.
  - `02-buildkit-tar` (`buildtar` package): `BuildContext([]File) ([]byte, error)` — assemble an in-memory tar suitable for `cli.ImageBuild`. 5 tests covering single-file round-trip, multi-file order + mode + body preservation, empty-input-is-valid-empty-tar, binary-data-survives (256-byte payload of every byte value), tar-terminator-block-present (catches missing `tw.Close()`). Test helper `readBack` does the inverse with `tar.NewReader`.
  - `03-restart-on-exit` (`restart` package): `ShouldRestart(events.Message) bool` + `Run(ctx, DockerAPI) error`. `DockerAPI` = `Events` + `ContainerStart`. 10 tests: 5 for `ShouldRestart` (non-zero-exit triggers, exit=0 ignored — `docker stop` clean shutdown, non-container Type ignored, non-die Action ignored, missing exitCode attribute is conservatively false), 5 for `Run` (restarts on non-zero exit, doesn't restart clean exit, continues after `ContainerStart` error, propagates transport errors from `errCh`, returns nil on ctx-cancel). `fakeAPI` is channel-backed so the test drives the runner by `<-` into its `msgCh`.
- All exercise/mini-project tests carry `//go:build exercise`; default `go test ./...` stays green (all 10 `09-docker` runtime packages show `[no test files]`).
- Verified: `go build ./...`, `go vet ./...`, `go vet -tags=exercise ./...`, `go test ./...` all clean. `go test -tags=exercise ./09-docker/...` shows expected failures: mini-project (12), exercise 01 (2 — the other 4 stats tests pass coincidentally because zero matches their edge-case expectations), exercise 02 (5), exercise 03 (5 — others pass coincidentally). No panics, no hangs.
- `09-docker/PLAN.md` Status flipped (Examples/Mini-project/Exercises ticked; README walkthrough still ☐); `09-docker/README.md` status header updated.
- **Go version held at 1.26.0** — docker/docker v28.5.2 didn't require a bump beyond what client-go already pulled in last session. CI `go-version: '1.26'` unchanged.

**Files touched:** ~22 new files under `09-docker/` (examples + mini-project + exercises). `go.mod`, `go.sum` updated. Docker deps added: `github.com/docker/docker` v28.5.2+incompatible + ~25 transitive deps (containerd/errdefs, distribution/reference, docker/go-connections, docker/go-units, moby/docker-image-spec, opencontainers/image-spec, opencontainers/go-digest, several go.opentelemetry.io/* packages, moby/sys/atomicwriter, moby/term, morikuni/aec, etc.).

**Open / next:**

- User to work through examples 01→06 (each TODO-block in `main.go`)
- Implement `image-pruner` until `go test -tags=exercise ./09-docker/mini-project/...` is green
- Implement the 3 exercises in any order
- Walkthrough doc in `09-docker/README.md` (currently stub + comparison table)
- Examples 01-06 require a running Docker daemon (`docker info` must succeed) to actually run; they compile and lint clean without one
- Next scaffolding step (future session): `10-observability/`

**Notes:**

- **`+incompatible` in `github.com/docker/docker v28.5.2+incompatible`** is expected — the moby/docker repo doesn't declare a `go.mod` major version (still on v0.x.y in its own go.mod for backward-compat reasons). The "+incompatible" suffix is Go modules' way of saying "I trust you, package author, that v28 means major version 28 despite the missing /v28 path segment." Not a problem; this is the canonical import path everyone uses.
- **API version negotiation is non-optional.** Without `client.WithAPIVersionNegotiation()`, your tool pins to the SDK's compile-time max API version — and any older daemon (any docker host older than the SDK release) rejects your calls with "client too new." The pattern is so universal it should be muscle memory. Documented in 01-connect's README.
- **The pull-reader drain footgun** in 03-pull-and-run is the #1 source of "my pull silently doesn't finish" bugs. The daemon considers the pull "in progress" until the response body is read to EOF. `io.Copy(io.Discard, rc)` is mandatory even if you don't care about the bytes. Documented inline in the example AND in its README.
- **`stdcopy.StdCopy` vs `io.Copy`** — without a TTY, docker multiplexes stdout+stderr over one stream with an 8-byte-per-chunk header. `io.Copy(stdout, logs)` writes the header bytes into your output. With `Tty: true` on container create, the stream ISN'T multiplexed and `StdCopy` would misread it. 04-logs-stream's README has the full frame diagram + the TTY-detection pattern (Inspect first).
- **`ContainerWait`'s two-channel return** is the SDK's "long-poll-until-condition" pattern. Same shape as `Events` (06). Always select on the errCh too — daemon-going-away is silent on the statusCh.
- **Mini-project `--force` regression catcher:** `TestSync_ForceFlagThreadsThrough` exists because it's easy to thread `Force` from CLI → `Policy` but forget to copy it into `RemoveOptions.Force` at the actual call site. The test uses a `captureRemoveFake` wrapper to assert the value lands at the right point.
- **Mini-project policy OR-ing** matches `docker image prune --filter "until=168h"` semantics: an image hits the prune list if ANY enabled policy says yes. The `RemoveUnused` axis is a separate filter — even an old, untagged image is kept if a stopped container references it. Tests pin this contract.
- **Exercise 01 (container-stats) deliberately omits the streaming-decode part.** The interesting bug is the math: cumulative-counter → delta → percentage with the `onlineCPUs` multiplier. People get the `onlineCPUs` factor wrong all the time (cap at 100% even when the container is using 4 cores). The "wire it to a real daemon" is documented in the exercise's README as a follow-up.
- **Exercise 01 first-sample test (`TestCPUPercent_FirstSample`)** explicitly only asserts "must not return NaN" — both conventions (return 0 for first sample, OR return the value computed against zero-prev) are accepted. The NaN case is the actual bug; a sane implementation never produces it.
- **Exercise 03 `ShouldRestart` test pins "missing exitCode → false"** as the conservative choice. A real-world daemon should always include the attribute on a die event, but defensive code shouldn't assume it.
- **Exercise 03 `Run` continues on `ContainerStart` errors** — the daemon being flaky on one container shouldn't stop the supervisor. Same shape as the mini-project's "log and keep going" approach.

---

## 2026-05-22 — `08-kubernetes/` scaffolded

**Goal:** Flesh out `08-kubernetes/` following the `07-aws/` pattern: examples with TODOs, mini-project + tests, exercises with failing tests. Spec is in `08-kubernetes/PLAN.md`. (SESSIONS note from the prior entry said "08-iac-tf-with-go" — that was a typo for the section title; root README/PLAN sequence has 08 as Kubernetes. IaC is section 11.)

**Done:**

- 6 example folders, each with TODO-style `main.go` + concept README:
  - `01-load-config` (`rest.InClusterConfig` → fall back to `clientcmd.BuildConfigFromFlags` + `Discovery().ServerVersion()` smoke probe)
  - `02-list-pods` (`CoreV1().Pods(ns).List` + label selectors + the `Pods("")` all-namespaces convention + the most-touched fields on `corev1.Pod`)
  - `03-get-deployment` (`AppsV1().Deployments(ns).Get` + Spec/Status split + `apierrors.IsNotFound`)
  - `04-watch-basic` (`Watch(ctx, opts) watch.Interface` + `ResultChan` consumption + the "raw watch dies" footgun explanation that motivates 05)
  - `05-informer` (`NewSharedInformerFactoryWithOptions` + ResourceEventHandlerFuncs + resync semantics + the DeletedFinalStateUnknown tombstone case + `WaitForCacheSync`)
  - `06-create-configmap` (`Create(ctx, obj, CreateOptions{})` + ObjectMeta vs server-filled fields + `IsAlreadyExists` + server-side-apply mention as the production-grade alternative)
- Mini-project `crashloop-alert`: informer-based watcher + dedup + pluggable Sink (stdout or webhook). Scaffold split into `IsCrashLooping(*Pod) bool` / `CrashLoopingContainer(*Pod) string` / `Deduper` (clock-injectable via `Now func() time.Time`) / `Sink` interface with `StdoutSink` + `WebhookSink` impls / `newPodHandler` (closes over deduper + sink) / `Run(ctx, kubernetes.Interface, ns, *Deduper, Sink, errOut) error`. `main_test.go` has 11 tests covering: detection (Waiting/CrashLoopBackOff vs Running vs ImagePullBackOff vs multi-container any-counts), Deduper (first-pass, blocks-within-cooldown, alerts-after, per-key isolation), StdoutSink writes parseable JSON line, WebhookSink POSTs JSON and errors on non-2xx, end-to-end `Run` test using `fake.NewSimpleClientset` (verifies a crashlooping pod fires exactly one alert and a healthy pod is silent), and a dedup test that drives an Update through the fake clientset and verifies the handler suppresses.
- 3 exercises with failing tests under `//go:build exercise`:
  - `01-namespace-audit` (`audit` package): `Audit(ctx, NamespaceAPI, requiredLabel) ([]string, error)`. 5 tests: flags-missing-label, all-present, empty-value-counts-as-present (kubectl-equivalent contract), API-order preservation, list-error propagates. `NamespaceAPI` interface = the one `List` method the package needs.
  - `02-resource-counter` (`counter` package): `Count(ctx, ClusterAPI) ([]Row, error)` — pods/deployments/services per namespace. 5 tests: tally-per-ns, empty cluster, namespace-with-no-resources, ns-order preservation, error propagation. `ClusterAPI` = 4 methods (one List per resource type, judgment call vs four one-method interfaces).
  - `03-rolling-restart` (`rollrestart` package): `RollingRestart(ctx, DeploymentAPI, ns, name, now time.Time) error` — patches the pod template's `kubectl.kubernetes.io/restartedAt` annotation, same key kubectl uses, so it plays nicely with `kubectl rollout status`. 5 tests: patch type is StrategicMerge, name plumbed through, body parses as JSON with the right RFC3339 annotation, error propagates, two calls produce distinct bodies (timestamp refresh check). `DeploymentAPI` interface = just the `Patch` method.
- All exercise/mini-project tests carry `//go:build exercise`; default `go test ./...` stays green. Default suite shows all 8 `08-kubernetes` runtime packages as `[no test files]` (mini-project test gated; example packages are stub-only main.go).
- Verified: `go build ./...`, `go vet ./...`, `go vet -tags=exercise ./...`, `go test ./...` all clean. `go test -tags=exercise ./08-kubernetes/...` shows expected failures: mini-project (9 explicit + 4 stub-coincidence passes), exercise 01 (5), 02 (5), 03 (5). No panics, no hangs.
- `08-kubernetes/PLAN.md` Status flipped (Examples/Mini-project/Exercises ticked; README walkthrough still ☐); `08-kubernetes/README.md` status header updated.
- **CI bump:** `go get k8s.io/client-go@latest` bumped `go.mod` from `go 1.24` → `go 1.26.0` (toolchain requirement of client-go v0.36.1). Updated `.github/workflows/ci.yml` `go-version: '1.24'` → `'1.26'` to match.

**Files touched:** ~28 new files under `08-kubernetes/` (examples + mini-project + exercises). `go.mod`, `go.sum`, `.github/workflows/ci.yml` updated. k8s deps added: `k8s.io/api`, `k8s.io/apimachinery`, `k8s.io/client-go` v0.36.1 + ~30 transitive deps.

**Open / next:**

- User to work through examples 01→06 (each TODO-block in `main.go`)
- Implement `crashloop-alert` until `go test -tags=exercise ./08-kubernetes/mini-project/...` is green
- Implement the 3 exercises in any order
- Walkthrough doc in `08-kubernetes/README.md` (currently stub + comparison table)
- Examples 01-06 require a reachable k8s cluster (minikube/kind/colima) to actually run; they compile and lint clean without
- Next scaffolding step (future session): `09-docker/`

**Notes:**

- **`Pods("")` all-namespaces convention** is documented in 02-list-pods README. It's a client-go thing — Python's k8s client takes an explicit `all_namespaces=True` kwarg.
- **Spec vs Status doctrine** in 03-get-deployment is the most-load-bearing concept in all of client-go. The example pins it with `Spec.Replicas` (your intent) vs `Status.ReadyReplicas` (what the controller reconciled).
- **Raw Watch vs Informers** — 04's README explicitly motivates 05 by listing the failure modes of raw watches (connection rotation, etcd compaction, no initial state). The reader should leave 04 frustrated and 05 relieved.
- **Informer resync vs real updates:** UpdateFunc has to be idempotent because the factory re-fires it every resync period (30s in the examples). 05's README pins this — "spamming work on every resync is the most common informer bug."
- **The `DeletedFinalStateUnknown` tombstone case** is acknowledged in 05's README but the example code skips it. Reason: the example's other concepts are already heavy; tombstones are a "thing to know exists" not a thing to write the first time.
- **Mini-project: `Run` blocks on `<-ctx.Done()`** and returns nil on clean shutdown — same shape as 05-informer. The end-to-end test cancels the ctx to stop the goroutine.
- **Mini-project: `fake.NewSimpleClientset(initialObjs...)`** is the gateway drug for k8s testing — you can pre-seed pods, then the informer factory built on top delivers them as Add events to your handler. The `TestRun_InformerFiresOnCrashLoopingPod` test relies on this entirely; no real cluster needed.
- **Mini-project: dedup uses a real-clock wallclock**, not an event-version counter. That's the right choice for a per-pod alert rate limit; an event-version counter would be wrong (different pods could share keys, resync would reset, etc.). Test injects `Now func() time.Time` to keep it deterministic — same fake-clock pattern as 06-testing/02-fake-clock.
- **Exercise 02 uses a single 4-method interface** rather than four one-method interfaces. The exercise's README acknowledges this is a judgment call — both are idiomatic.
- **Exercise 03 patches `kubectl.kubernetes.io/restartedAt`** — same key kubectl uses, so the tool's restarts are visible to `kubectl rollout status` / `kubectl rollout history`. The README pins this WHY so future-me doesn't change the key thinking it's arbitrary.
- **Exercise 03 injects `now time.Time`** instead of having tests inject a Now function. Reason: this function does one thing once per call — passing a value is simpler than passing a clock. Same trade-off discussion as in 06-testing/02-fake-clock README.
- **CI Go version was already mismatched** (1.24 in CI, but go.mod was bumped by the AWS SDK to 1.24 in last session — and now client-go v0.36.1 bumps it again to 1.26.0). Bumped CI to 1.26 to match. Watch for this every time we add a major SDK.

---

## 2026-05-22 — `07-aws/` scaffolded

**Goal:** Flesh out `07-aws/` following the `06-testing/` pattern: examples with TODOs, mini-project + tests, exercises with failing tests. Spec is in `07-aws/PLAN.md`.

**Done:**

- 7 example folders, each with TODO-style `main.go` + concept README:
  - `01-config-and-creds` (`config.LoadDefaultConfig` + the credentials chain + `cfg.Credentials.Retrieve`)
  - `02-s3-list` (ListBuckets + `s3.NewListObjectsV2Paginator`)
  - `03-s3-upload-download` (PutObject Body=io.Reader + GetObject Body=io.ReadCloser round-trip)
  - `04-s3-presigned` (`s3.NewPresignClient` + `s3.WithPresignExpires(5*time.Minute)`)
  - `05-ec2-list` (DescribeInstancesPaginator + Filter DSL + Reservations→Instances flatten note)
  - `06-assume-role` (`sts.NewFromConfig` + `stscreds.NewAssumeRoleProvider` + `aws.NewCredentialsCache` wrap)
  - `07-mocking-sdk` (interface-at-consumption-site pattern, ships with WORKING `s3util_test.go` — same deviation as `06-testing` examples, since the lesson IS the testing pattern)
- Mini-project `s3sync`: mirror local dir → S3 bucket. Scaffold split into `WalkLocal` / `computeMD5` / `ListRemote` / `Plan` / `Sync` + an `S3API` interface (3 methods: ListObjectsV2, PutObject, DeleteObject). `main_test.go` has 9 tests covering: forward-slash-key walking, ETag-quote unwrap, plan rules (upload-new/changed, skip-identical, omit-extras-without-delete, include-extras-with-delete), happy-path Sync, dry-run-doesn't-call-AWS, concurrency-peak (uses atomic.Int32 + CAS to record max in-flight + `holdPut` channel pattern lifted from 05-concurrency mini-project), --delete behavior, error propagation. Uses `fakeS3` — no real AWS.
- 3 exercises with failing tests under `//go:build exercise`:
  - `01-bucket-inventory` (`inventory` package): `Inventory(ctx, api) ([]Row, error)` + `WriteCSV(w, rows) error`. 5 tests covering empty account, flatten-buckets-and-objects, bucket order preservation, single-bucket error aborts, CSV header + RFC3339 timestamps. `S3API` interface = ListBuckets + ListObjectsV2.
  - `02-find-untagged` (`untagged` package): `FindUntagged(ctx, api, requiredKey) ([]string, error)`. 5 tests covering missing-tag flag, empty-VALUE counts as present, multi-reservation flatten, paginator-must-consume-both-pages, error propagation. `EC2API` interface = DescribeInstances.
  - `03-cleanup-old` (`cleanup` package): `Cleanup(ctx, api, bucket, prefix, cutoff, dryRun) ([]string, error)`. 5 tests covering only-old-deleted, dry-run-no-delete-calls, prefix-passed-through, list-error-aborts, delete-error-propagates. `S3API` interface = ListObjectsV2 + DeleteObject.
- All exercise/mini-project tests carry `//go:build exercise` so default `go test ./...` stays green. Default suite shows `07-aws/07-mocking-sdk` as ok (its tests target the working canonical impl — same deviation from earlier sections as 06-testing examples).
- Verified: `go build ./...`, `go vet ./...`, `go vet -tags=exercise ./...`, `go test ./...` all clean. `go test -tags=exercise ./07-aws/...` shows expected failures: mini-project (9), exercise 01 (5), 02 (5), 03 (5).
- `07-aws/PLAN.md` Status flipped (Examples/Mini-project/Exercises ticked; README walkthrough still ☐); `07-aws/README.md` status header updated.
- **CI bump:** `go get` of AWS SDK v2 bumped `go.mod` from `go 1.23.0` → `go 1.24` (toolchain requirement of `aws-sdk-go-v2`). Updated `.github/workflows/ci.yml` `go-version: '1.22'` → `'1.24'` to match.

**Files touched:** ~26 new files under `07-aws/` (examples + mini-project + exercises). `go.mod`, `go.sum`, `.github/workflows/ci.yml` updated. AWS deps added: `aws-sdk-go-v2`, `config`, `service/s3`, `service/ec2`, `service/sts`, `credentials/stscreds`.

**Open / next:**

- User to work through examples 01→07 (each TODO-block in `main.go`; example 07 ships fully working — extend its TODO instead)
- Implement `s3sync` until `go test -tags=exercise ./07-aws/mini-project/...` is green
- Implement the 3 exercises in any order
- Walkthrough doc in `07-aws/README.md` (currently stub + comparison table)
- Examples 01-06 require AWS creds + (often) a real bucket/instance/role to actually run; they compile and lint clean without
- Next scaffolding step (future session): `08-iac-tf-with-go/` (per the root PLAN — IaC integration / wrapping terraform from Go)

**Notes:**

- **Examples ship as TODO `main.go`** for 01-06 (matching 05-concurrency / 04-http-servers convention), but 07-mocking-sdk ships with **working code + tests** because its entire lesson IS the testing pattern. Same deviation as 06-testing examples — documented in 07's README.
- **The S3API interface in `s3sync` is intentionally not a single god-interface** but only the 3 methods the tool actually uses. Same pattern enforced in each exercise — every exercise defines its OWN narrow interface. That's the lesson worth repeating.
- **`fakeS3` in mini-project tracks peak in-flight with `atomic.Int32` + CAS** — copied from `05-concurrency/mini-project/fanout-ping`. The `holdPut chan struct{}` channel that blocks all PutObject calls until the test closes it is new for this section, and is the cleanest way I found to assert concurrency parallelism with a fake.
- **`waitForInflight` busy-waits without `time.Sleep`** — it yields via a tiny goroutine + channel-receive. Reason: time.Sleep introduces unrelated flakiness if a worker is just slow on CI. The wait loop bounds at 2 seconds.
- **In `FindUntagged`, empty-VALUE is treated as "tag is present"** — this is a real AWS behavior call (Owner="" is common, often meaning "intentionally unowned-but-tracked"). The test pins this contract.
- **In `Cleanup`, the fake honours the Prefix filter server-side** so the prefix-passthrough test isn't pure ceremony — it verifies the user actually passes Prefix into the Input.
- **`PresignClient` doesn't call AWS** — the signing is local. Noted in `04-s3-presigned/README.md` for cost/CloudTrail awareness.
- **`aws.NewCredentialsCache(provider)`** in 06 is non-obvious — without it every API call re-runs AssumeRole. Documented in the example's README.
- **CI Go version was already mismatched** (1.22 in CI vs 1.23.0 in go.mod from previous sessions) — likely worked due to toolchain auto-download. Bumped to 1.24 now that the AWS SDK requires it explicitly.

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
