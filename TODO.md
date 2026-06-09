# TODO

Product and engineering backlog. Items use **train** tags (`0.0.1`, `0.0.2`, …) to sequence work while the project stays on **0.0.x** — see [AGENTS.md](./AGENTS.md#versioning).

**Intended outcomes**

1. Compare two variation arms and read pass@k and cost deltas from the CLI.
2. Open a harness round trace in Jaeger via `openeval traces`.
3. Run scenarios that resemble maintainer work, not hello-world fixtures.
4. Verify setup with `openeval doctor` and install without cloning the repo.
5. Run at least one integrated external benchmark with documented prerequisites.

**Order:** trains `0.0.1`–`0.0.5` (first-run and README), then `0.0.6`–`0.0.8` (install and CI), then authoring and telemetry docs.

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

## Deferred (per decisions and README status)

Do not schedule these until the 0.0.x bar in AGENTS.md is met:

- [ ] `0.1.0` / semver release tagging
- [ ] Pi and OpenCode harness drivers (observability-only today)
- [ ] OTLP query rollups / `openeval trends`
- [ ] DeepSWE and margin-eval both production-ready (one only in `0.0.10`)
- [ ] Docker packaging as the default run path
- [ ] Full cursor-otel-hook parity beyond harness trace correlation
