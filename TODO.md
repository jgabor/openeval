# TODO

Product and engineering backlog. Items use **train** tags (`0.0.1`, `0.0.2`, …) to sequence work while the project stays on **0.0.x** — see [AGENTS.md](./AGENTS.md#versioning).

**Intended outcomes**

1. Compare two variation arms and read pass@k and cost deltas from the CLI.
2. Open a harness round trace in Jaeger via `openeval traces`.
3. Run scenarios that resemble maintainer work, not hello-world fixtures.
4. Verify setup with `openeval doctor` and install without cloning the repo.
5. Run at least one integrated external benchmark with documented prerequisites.

**Order:** trains `0.0.1`–`0.0.5` establish compare-first onboarding;
`0.0.6`–`0.0.8` cover install and CI; `0.0.9`–`0.0.12` cover authoring
and honest documentation. Paired campaign work begins at `0.0.13` only after
the release-worthy vertical slice is intact.

---

## First-run and README

- [ ] [compare-first-quickstart:0.0.1] Compare-first quick start (`cursor-agent`, not mock)

  README quick start currently runs `--agent mock`. The compare example in README uses DeepSWE, which is not runnable on the default path. Mock produces valid-looking `score.json` without invoking an agent.

  ```bash
  openeval run --scenario example-fixtures --variation default --agent cursor --rounds 3
  openeval run --scenario example-fixtures --variation with-demo-skill --agent cursor --rounds 3
  openeval compare ./scenarios/example-fixtures/runs/default ./scenarios/example-fixtures/runs/with-demo-skill
  ```

  - [ ] Replace README quick start with the three-command compare flow above
  - [ ] Document that a new user with authenticated `cursor-agent` should see a compare table from README alone
  - [ ] Move mock to a secondary path (CI / no API key)
  - [ ] Compare output documents pass@k, `cost_usd_total`, cost per passed task, and deltas

- [ ] [demo-command:0.0.2] `openeval demo` for the compare-first flow

  - [ ] Command runs baseline + skill variation + `compare` (or prints commands with `--dry-run`)
  - [ ] Missing `cursor-agent` reports install/login steps
  - [ ] Missing or invalid config reports copy/fix steps
  - [ ] Exit 0 prints a compare table when cursor setup is valid
  - [ ] Does not require OTLP/Jaeger (see `0.0.4`)

- [ ] [mini-production-scenario:0.0.3] Expand `example-fixtures` beyond hello-world tasks

  `hello-verify` and `edit-file` exercise the pipeline only. README references DeepSWE and margin-eval; the shipped quick-start scenario does not reflect that scope.

  - [ ] Add 3–5 tasks on a small fixture repo (e.g. fix a failing test, apply a lint rule, one-file feature with verifier)
  - [ ] Keep shell/script verifiers
  - [ ] Retain `default` vs `with-demo-skill` variations for compare
  - [ ] `openeval report` output reflects realistic task names and verifiers
  - [ ] CI can run a subset with `--agent mock`

- [ ] [otel-in-a-box:0.0.4] Bundled local OTLP for trace lookup

  README lists an OTLP backend as a prerequisite but does not ship a default collector. Harness writes `trace_id` to `score.json` without a running collector; live span export is not verified on the default path.

  Pick one primary path:

  - [ ] `docker compose` under `examples/otel/` (Jaeger on `:4318` / `:16686`), or
  - [ ] `openeval instrument --agent cursor --otel-up docker`, or
  - [ ] Documented one-liner in README

  - [ ] After instrument + collector up, `openeval traces --task … --round …` yields a resolvable Jaeger URL
  - [ ] README trace lookup references the bundled setup
  - [ ] Harness still completes when collector is unreachable (non-blocking export)

- [ ] [compare-hero-readme:0.0.5] Reorder README: compare and traces before mock quick start

  - [ ] Opening sections show compare table example and trace lookup flow
  - [ ] Mock labeled CI/hermetic only
  - [ ] Quick start order: install → compare flow → optional instrument/traces
  - [ ] Trim or relocate Supported telemetry content per `0.0.11`
  - [ ] No README path presents mock scores as production eval evidence

---

## Install, doctor, and CI

- [ ] [doctor-command:0.0.6] `openeval doctor`

  - [ ] Check `cursor-agent` on PATH or `agents.cursor.command`
  - [ ] Check `cursor-agent status` (authenticated)
  - [ ] Check config at `$XDG_CONFIG_HOME/openeval/config.yaml`
  - [ ] Check `skills.aliases` paths resolve for skill variations
  - [ ] Warn if OTLP endpoint unreachable (non-fatal)
  - [ ] Check `~/.cursor/hooks.json` after `instrument`
  - [ ] Each failure prints remediation steps
  - [ ] README references doctor before first `run --agent cursor`

- [ ] [frictionless-install:0.0.7] Single documented install path without cloning

  README lists `mage install`, `go build`, and `go install openeval@latest`. Module publish status for the last path is unclear.

  - [ ] Verify and document `go install github.com/jgabor/openeval/cmd/openeval@latest` (or correct module path)
  - [ ] Optional: curl script or Homebrew tap
  - [ ] README: one recommended install path; source build under a separate heading
  - [ ] Install docs consistent with AGENTS.md versioning

- [ ] [github-action-recipe:0.0.8] Example GitHub Action workflow

  - [ ] Add `examples/ci/openeval.yml` (or similar) for consumers
  - [ ] Fast job: `openeval run --agent mock --rounds 1` on `example-fixtures`, no secrets
  - [ ] Document optional job: `cursor-agent` compare on release branches (`CURSOR_API_KEY`)
  - [ ] Comment expected cost/runtime for real-agent jobs
  - [ ] README links the recipe under CI / release gates
  - [ ] Mock job runs on GitHub Actions without cursor credentials

---

## Authoring and scenarios

- [ ] [init-scenario:0.0.9] `openeval init scenario`

  Custom scenarios today require hand-written YAML, fixtures, and verifier scripts.

  - [ ] Scaffold `scenario.yaml`, `fixtures/`, stub verifier under `verifiers/`
  - [ ] Optional: `--from <repo-path>` with language detection and verifier template
  - [ ] Generated layout matches `examples/scenarios/example-fixtures/`
  - [ ] Scaffold runs with `--agent mock` after minimal edits
  - [ ] README custom-scenario section links to init

- [ ] [external-benchmark:0.0.10] One integrated external scenario end-to-end

  README lists DeepSWE and margin-eval. Pick one; do not polish both in this train.

  - [ ] `openeval run --scenario <chosen>` works with documented prerequisites
  - [ ] README subsection: dependencies, cost/runtime, default task subset
  - [ ] Smoke path completes at least one task/round on a real agent
  - [ ] Missing deps or auth produce actionable errors
  - [ ] Non-chosen scenario remains marked deferred in README

---

## Documentation

- [ ] [telemetry-shipped-slice:0.0.11] Align Supported telemetry with shipped behavior

  README Supported telemetry lists resources, events, and aggregates; most are planned or partial.

  Option A:

  - [ ] Ship hook export for session cost, skills active, and tool calls
  - [ ] Document with Jaeger screenshots in README

  Option B:

  - [ ] Split README into **Shipped** vs **Planned** sections

  - [ ] README does not imply full inventory is available today
  - [ ] Consistent with AGENTS.md and earn-your-complexity decision

- [ ] [skill-compare-playbook:0.0.12] Document skill-variation compare workflow

  Variations `with-demo-skill` and `with-plugin-skill` work; the compare workflow is not documented as a single procedure.

  ```bash
  openeval run --scenario example-fixtures --variation default --agent cursor --rounds 3
  openeval run --scenario example-fixtures --variation with-demo-skill --agent cursor --rounds 3
  openeval compare ./scenarios/example-fixtures/runs/default ./scenarios/example-fixtures/runs/with-demo-skill
  ```

  - [ ] README section: `skills.aliases` → variation → two runs → compare → `cost_by_skill` in `score.json`
  - [ ] Document `--plugin-dir` for `.claude-plugin` skills
  - [ ] Recommend `--rounds 3` for pass@k stability
  - [ ] Cross-link `0.0.1` and `0.0.3` when complete

---

## Paired campaigns and retained evidence

These trains extend the verifier-backed scenario harness from independent
variation runs to controlled, retained campaigns. They are generic OpenEval
capabilities. Integrations such as ktx consume the contracts without adding
tool-specific behavior to the runner.

- [ ] [attempt-evidence-contract:0.0.13] Retain one structured result per attempt

  Agent drivers currently reduce runtime observations to cost and trace ID, and
  an agent-process error aborts the run. Preserve the observations needed to
  compare complete agent trajectories and distinguish task failure from invalid
  evidence.

  - [ ] Replace the driver tuple return with a structured attempt result
  - [ ] Retain status, process exit, trace/session ID, start/end time, and duration
  - [ ] Retain input, output, reasoning, cache-read, and cache-write tokens independently
  - [ ] Preserve runtime-reported cost separately from any price-derived estimate
  - [ ] Retain runtime, version, model, variant, and measurement provenance when available
  - [ ] Preserve explicit numeric zero separately from unavailable or malformed fields
  - [ ] Retain agent stdout and stderr as bounded artifacts with SHA-256 digests
  - [ ] Record agent failures as attempts and continue the run
  - [ ] Distinguish infrastructure-invalid, completed-quality-fail, and completed-quality-pass
  - [ ] Add a versioned score schema and validate it when loading
  - [ ] Keep existing score artifacts readable or document and test the intentional break

- [ ] [immutable-attempt-artifacts:0.0.14] Make run evidence immutable and reproducible

  Named variation runs currently replace prior directories. A retained campaign
  must preserve the exact task, source, configuration, runtime, and result
  artifacts used for every attempt.

  - [ ] Give each run and attempt a unique immutable ID
  - [ ] Never remove an existing run directory implicitly
  - [ ] Retain a campaign manifest before agent execution
  - [ ] Hash the scenario, task, verifier, source fixture, variation, skills, environment contract, and agent command
  - [ ] Retain an artifact manifest with path, size, and SHA-256
  - [ ] Record execution order, host/runtime identity, and OpenEval version
  - [ ] Validate that score rows and artifacts belong to the declared campaign
  - [ ] Add `openeval validate <run-dir>` with stable JSON output and documented exit codes

- [ ] [historical-task-bundles:0.0.15] Support frozen source and structured grading per task

  Scenario-level `fixtures/` gives every task the same tree. Historical coding
  tasks need independent source snapshots, hidden verifier inputs, and patch
  evidence.

  - [ ] Allow each task to select its own frozen source bundle
  - [ ] Retain source provenance and digest without fetching during execution
  - [ ] Separate agent-visible source from verifier-only and reference artifacts
  - [ ] Derive and retain a normalized patch against the frozen source
  - [ ] Retain changed-file, added-line, and deleted-line footprint metrics
  - [ ] Let verifier scripts emit a versioned structured result
  - [ ] Support test, semantic-equivalence, code-review, and footprint dimensions
  - [ ] Retain verifier stdout, stderr, exit status, and artifact hashes
  - [ ] Distinguish verifier infrastructure errors from a valid failing grade
  - [ ] Keep optional LLM grading separate from deterministic test evidence and agent cost

- [ ] [paired-variation-scheduler:0.0.16] Run controlled variation pairs in one campaign

  Separate variation commands execute whole arms sequentially and cannot
  counterbalance temporal or cache effects. Campaign scheduling must pair the
  same task and repetition before execution.

  - [ ] Accept two or more named variations in one run
  - [ ] Define the comparison unit as task plus repetition
  - [ ] Use a retained deterministic seed for task ordering
  - [ ] Counterbalance two-arm execution as AB/BA across pairs
  - [ ] Run every scheduled attempt regardless of verifier reward
  - [ ] Retry only infrastructure-invalid attempts under an explicit bounded policy
  - [ ] Retain invalid and superseding retry records without replacing evidence
  - [ ] Default to concurrency one; retain concurrency when explicitly changed
  - [ ] Add a no-agent preflight that prints the complete schedule and estimated attempt count
  - [ ] Require explicit execution approval for paid campaigns
  - [ ] Preserve ordinary independent `openeval run` and pass@k behavior

- [ ] [sandboxed-attempt-execution:0.0.17] Implement explicit pinned-container runs

  This is an opt-in campaign execution path, not Docker as the default for all
  OpenEval commands.

  - [ ] Implement the existing `--image` run flag
  - [ ] Pin and retain the image digest used by each attempt
  - [ ] Mount only the attempt workspace, declared artifacts, and read-only credentials
  - [ ] Allocate disposable `HOME` and XDG directories per attempt
  - [ ] Make runtime and verifier network policies explicit and separately configurable
  - [ ] Run deterministic grading without network access by default
  - [ ] Bound CPU, memory, process count, and attempt duration
  - [ ] Clean up containers after success, failure, timeout, or interruption
  - [ ] Retain enough diagnostics to classify setup and execution failures
  - [ ] Keep mock and host execution available for tests and local development

- [ ] [paired-campaign-report:0.0.18] Add task-paired resource and quality reporting

  Pass@k and aggregate cost do not answer whether a treatment reduced the
  recorded workload or changed the typical task while preserving quality.

  - [ ] Pair attempts by task, repetition, baseline, and candidate
  - [ ] Report exact recorded-workload token and cost totals per repetition
  - [ ] Report paired per-task percentage changes and their geometric mean
  - [ ] Report deterministic paired bootstrap intervals with a retained seed
  - [ ] Treat repeated runs of one task as one task cluster when pooling
  - [ ] Report test, equivalence, review, and footprint win/loss/tie counts
  - [ ] Report tool-call inflation, duration, patch size, and failure classes
  - [ ] Keep runtime-reported cost separate from price-derived estimates
  - [ ] Never replace missing measurements with zero
  - [ ] Annotate unusual attempts without removing them automatically
  - [ ] Emit stable JSON from `compare` as well as human-readable text
  - [ ] Regenerate campaign summaries byte-identically from retained evidence
  - [ ] Label descriptive intervals as uncertainty, not causal or significance claims

- [ ] [ktx-paired-campaign-proof:0.0.19] Prove the generic campaign contract with ktx

  Use ktx as an acceptance vehicle after the generic campaign capabilities
  above ship. OpenEval owns execution and grading; ktx owns compression and its
  native database evidence.

  - [ ] Run passthrough and compression arms with the same OpenCode and ktx instrumentation
  - [ ] Give each attempt a new disposable `KTX_DB`, `HOME`, and XDG root
  - [ ] Retain ktx binary, configuration, pipeline, database, and artifact digests
  - [ ] Consume ktx metrics per attempt rather than through one aggregate external map
  - [ ] Keep OpenCode runtime usage and ktx delivery evidence as separate authorities
  - [ ] Verify unchanged task quality before interpreting resource movement as useful
  - [ ] Complete an unpaid or explicitly bounded three-task smoke campaign
  - [ ] Validate every attempt and regenerate the campaign summary before any paid expansion
  - [ ] Require separate approval before the ten-task, two-repetition, two-arm campaign
  - [ ] Document claim limits: local task set, descriptive evidence, no provider-cache or causal claim

---

## Deferred (per decisions and README status)

Do not schedule these until the 0.0.x bar in AGENTS.md is met:

- [ ] `0.1.0` / semver release tagging
- [ ] Pi harness driver
- [ ] OTLP query rollups / `openeval trends`
- [ ] DeepSWE and margin-eval both production-ready (one only in `0.0.10`)
- [ ] Docker packaging as the default run path; explicit opt-in campaign sandboxing is tracked by `sandboxed-attempt-execution:0.0.17`
- [ ] Full cursor-otel-hook parity beyond harness trace correlation
- [milestone] 2026-06-14: opencode runtime driver plan (4 tasks) complete; see `.agentera/archive/plan-2026-06-14-opencode-runtime-driver.yaml`
