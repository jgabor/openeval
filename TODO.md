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

## ⇶ Critical

## ⇉ Degraded

## → Normal

- [ ] [id:bgmqlctjbo] [task:0.0.13] Store versioned external producer metadata and producer-defined measurements on the corresponding attempt result
- [ ] [id:fjcwycnppn] [task:0.0.13] Require every external measurement to declare a state and unit; measured values must be finite numbers, while unavailable and malformed values omit the value
- [ ] [id:kuaykibjzv] [task:0.0.13] Distinguish a missing external artifact, unavailable measurement, malformed measurement, and measured zero
- [ ] [id:jmfuavefdq] [task:0.0.13] Keep runtime usage, runtime cost, and external producer measurements as separate authorities
- [ ] [id:whlwclqsqg] [task:0.0.13] Do not add an aggregate external-measurement map to the variation-level telemetry summary
- [ ] [id:pflsygwbgp] [task:0.0.14] Allocate `external-metrics.json` inside each attempt artifact directory and pass its path to the child as `OPENEVAL_EXTERNAL_METRICS_PATH`
- [ ] [id:oherwvaici] [task:0.0.14] Remove any existing entry at the metrics path before starting the child process
- [ ] [id:ejfkfczcbt] [task:0.0.14] Read the artifact after the child exits; accept absence without synthesizing measurements or success
- [ ] [id:xmlwovczys] [task:0.0.14] Validate a bounded regular non-symlink file, UTF-8, one complete JSON document without trailing data, and the `openeval.external-metrics.v1` schema
- [ ] [id:qlvovfdclj] [task:0.0.14] Validate required producer metadata, unique measurement names, states, units, required or forbidden values by state, and finite numeric values
- [ ] [id:ykstfajfun] [task:0.0.14] Retain malformed artifacts and validation errors as attempt evidence rather than zero-valued measurements
- [ ] [id:evyzurfgav] [task:0.0.14] Retain exact artifact bytes, relative path, size, and SHA-256 and attach parsed valid measurements to the exact attempt
- [ ] [id:pifccfzezx] [task:0.0.14] Keep the raw per-attempt artifact authoritative; summaries and reports are derived views
- [ ] [id:hqbhbeigvz] [task:0.0.14] Add `Variation.Command` as a YAML sequence whose elements are literal argv entries
- [ ] [id:ssudlpqfem] [task:0.0.14] Preserve the existing direct-agent invocation when `command` is absent or empty
- [ ] [id:gpfhmoktpt] [task:0.0.14] Reject an empty wrapper executable or empty argv entry before agent execution
- [ ] [id:snssylenxo] [task:0.0.14] Prepend the wrapper without a shell, whitespace splitting, interpolation, or environment expansion
- [ ] [id:ytmuvxwgld] [task:0.0.14] Keep the selected runtime driver responsible for the resolved agent binary and its normal arguments
- [ ] [id:cwmjzaenug] [task:0.0.14] Retain the resolved agent identity and effective argv for direct and wrapped attempts
- [ ] [id:onlagrkdlq] [task:0.0.14] Test executable paths and arguments containing spaces
- [ ] [id:oojmegqtyh] [task:0.0.14] Keep existing scenarios without `command` byte-compatible when loaded
- [ ] [id:jjwkabxrwl] [task:0.0.18] Align external measurements only when paired attempts have matching names and units
- [ ] [id:nnsrwssdgc] [task:0.0.18] Report unmatched, unavailable, malformed, and absent external evidence without inventing counterparts
- [ ] [id:qzwovgoweh] [task:0.0.18] Do not sum ratios, average counts across unequal task sets, combine external values with runtime usage, or infer that a producer caused runtime movement
- [ ] [id:lcnujtldrb] [task:0.0.19] Have both arms publish versioned per-attempt external measurement artifacts
- [ ] [id:muahxyfcry] [task:0.0.19] Verify that OpenEval validates, retains, and compares ktx evidence without interpreting ktx-specific measurement names
- [ ] [id:ocekyawtvh] [task:0.0.19] Verify that missing or unavailable ktx evidence does not become passthrough, success, or zero

## First-run and README

Completed with an OpenCode-first manual flow and `openeval demo`; mock remains labeled for CI only.

```bash
openeval run --scenario example-fixtures --variation baseline --agent opencode --rounds 3
openeval run --scenario example-fixtures --variation with-demo-skill --agent opencode --rounds 3
openeval compare ./scenarios/example-fixtures/runs/baseline ./scenarios/example-fixtures/runs/with-demo-skill
```

Completed with four standard-library Python maintenance tasks and process-isolated scenario-owned graders outside the agent workspace. Parent assertions reject workspace-test replacement and import-side-effect mutation while a child imports workspace code. Final verification passed `mage build`, an uncached full suite without live agents, provider credentials, or a collector, and `mage check`. One audited one-round, two-arm full-corpus sample on OpenCode 1.18.11 with `opencode/big-pickle` passed 4/4 tasks in each arm in 248.852 seconds at $2.822868 estimated combined cost, with no harness failures; all eight retained workspaces also passed the isolated graders without new inference. This local sample does not establish variance or a general provider/model ranking, and grader isolation is not host sandboxing.

Train `0.0.4` shipped a pinned, optional Jaeger Compose collector at `examples/otel/compose.yaml`. Docker remains tracing-only: harness, compare, report, doctor, and normal instrumentation stay collector-optional. The retained manual smoke proved that Jaeger received the normalized OpenEval summary trace and that `openeval traces` printed the matching UI URL; native OpenCode OTEL remains optional separate-trace enrichment.

The next README train can now promote the shipped compare and trace workflows without implying that Docker is required for harness proof.

- [ ] [id:chrlpdaedf] [compare-hero-readme:0.0.5] Reorder README: compare and traces before mock quick start

- [ ] [id:qldvdzoqzo] Opening sections show compare table example and trace lookup flow

---

## Install, doctor, and CI

- [ ] [id:vwenelnkgw] [frictionless-install:0.0.7] Single documented install path without cloning

  The canonical `github.com/jgabor/openeval/cmd/openeval@latest` probe is not publicly resolvable, and example assets are not embedded. The current verified path requires a repository checkout.

- [ ] [id:ivntyvxwqc] Verify and document `go install github.com/jgabor/openeval/cmd/openeval@latest` (or correct module path)
- [ ] [id:uemgcurert] Optional: curl script or Homebrew tap

- [ ] [id:inclaigcgq] [github-action-recipe:0.0.8] Example GitHub Action workflow

- [ ] [id:kwyfmyzdjh] Add `examples/ci/openeval.yml` (or similar) for consumers
- [ ] [id:cqonolslmc] Fast job: `openeval run --agent mock --rounds 1` on `example-fixtures`, no secrets
- [ ] [id:rvydheplua] Document optional job: OpenCode compare on release branches with provider credentials
- [ ] [id:adnvzovmae] Comment expected cost/runtime for real-agent jobs
- [ ] [id:wazakrjrvw] README links the recipe under CI / release gates
- [ ] [id:ivbsiciloi] Mock job runs on GitHub Actions without cursor credentials

---

## Authoring and scenarios

- [ ] [id:ujdmjsiobr] [init-scenario:0.0.9] `openeval init scenario`

  Custom scenarios today require hand-written YAML, fixtures, and verifier scripts.

- [ ] [id:ygpsvsddjy] Scaffold `scenario.yaml`, `fixtures/`, stub verifier under `verifiers/`
- [ ] [id:xokuvfdxwe] Optional: `--from <repo-path>` with language detection and verifier template
- [ ] [id:nbarkeyirq] Generated layout matches `examples/scenarios/example-fixtures/`
- [ ] [id:jkvedkfuwd] Scaffold runs with `--agent mock` after minimal edits
- [ ] [id:zorqcpiqgh] README custom-scenario section links to init

- [ ] [id:ydbgmcezbt] [external-benchmark:0.0.10] One integrated external scenario end-to-end

  README lists DeepSWE and margin-eval. Pick one; do not polish both in this train.

- [ ] [id:viyntmpjja] `openeval run --scenario <chosen>` works with documented prerequisites
- [ ] [id:ioaiwazxui] README subsection: dependencies, cost/runtime, default task subset
- [ ] [id:xirsajvynt] Smoke path completes at least one task/round on a real agent
- [ ] [id:yakozwidtm] Missing deps or auth produce actionable errors
- [ ] [id:gnqfxkmklm] Non-chosen scenario remains marked deferred in README

---

## Documentation

Completed through Option B: README separates shipped telemetry from planned or deferred inventory.

Option A (not selected):

- [ ] [id:pbmhjkpptz] Ship hook export for session cost, skills active, and tool calls
- [ ] [id:hnarslhsdo] Document with Jaeger screenshots in README

  Option B (chosen):

- [ ] [id:mplkdxhdhj] [skill-compare-playbook:0.0.12] Document skill-variation compare workflow

  The compare workflow and runtime-specific discovery behavior are documented. The remaining cross-link is now unblocked because `0.0.3` is complete.

  ```bash
  openeval run --scenario example-fixtures --variation baseline --agent opencode --rounds 3
  openeval run --scenario example-fixtures --variation with-demo-skill --agent opencode --rounds 3
  openeval compare ./scenarios/example-fixtures/runs/baseline ./scenarios/example-fixtures/runs/with-demo-skill
  ```

- [ ] [id:irsuinecko] Cross-link `0.0.1` and `0.0.3` when complete

---

## Paired campaigns and retained evidence

These trains extend the verifier-backed scenario harness from independent
variation runs to controlled, retained campaigns. They are generic OpenEval
capabilities. Integrations such as ktx consume the contracts without adding
tool-specific behavior to the runner. The contracts include argv-safe
per-variation wrappers and versioned external measurements retained with each
attempt; neither contract gives OpenEval producer-specific semantics.

- [ ] [id:fvalnoosno] [attempt-evidence-contract:0.0.13] Retain one structured result per attempt

  Agent drivers currently reduce runtime observations to cost and trace ID, and
  an agent-process error aborts the run. Preserve the observations needed to
  compare complete agent trajectories and distinguish task failure from invalid
  evidence.

- [ ] [id:uknptdxbsy] Replace the driver tuple return with a structured attempt result
- [ ] [id:cddhuesuyh] Retain status, process exit, trace/session ID, start/end time, and duration
- [ ] [id:owoitzsfkl] Retain input, output, reasoning, cache-read, and cache-write tokens independently
- [ ] [id:fgyozjkxul] Preserve runtime-reported cost separately from any price-derived estimate
- [ ] [id:mrorwyiive] Retain runtime, version, model, variant, and measurement provenance when available
- [ ] [id:ghniublbux] Preserve explicit numeric zero separately from unavailable or malformed fields
- [ ] [id:nvpicjpstq] Retain agent stdout and stderr as bounded artifacts with SHA-256 digests
- [ ] [id:hetrvwsjdd] Record agent failures as attempts and continue the run
- [ ] [id:eqvtduxsiy] Distinguish infrastructure-invalid, completed-quality-fail, and completed-quality-pass
- [ ] [id:herscmcokv] Add a versioned score schema and validate it when loading
- [ ] [id:vpjeddglzo] Keep existing score artifacts readable or document and test the intentional break

- [ ] [id:alknfnxqpd] [immutable-attempt-artifacts:0.0.14] Make attempt execution immutable and reproducible

  Named variation runs currently replace prior directories. A retained campaign
  must preserve the exact task, source, configuration, invocation, runtime, and
  result artifacts used for every attempt. It must also give generic wrappers a
  safe invocation contract and a bounded path for publishing attempt evidence.

  **Immutable campaign state**

- [ ] [id:fkdegzmmjc] Give each run and attempt a unique immutable ID
- [ ] [id:hcjukgnnlb] Never remove an existing run directory implicitly
- [ ] [id:qivouxgcwv] Retain a campaign manifest before agent execution
- [ ] [id:rimpwejcmi] Hash the scenario, task, verifier, source fixture, variation, skills, environment contract, and effective agent argv
- [ ] [id:ujaqmqbogk] Retain an artifact manifest with path, size, and SHA-256
- [ ] [id:llkmbsukkm] Record execution order, host/runtime identity, and OpenEval version
- [ ] [id:usnfolrcog] Validate that score rows and artifacts belong to the declared campaign
- [ ] [id:xaqpyeefwf] Add `openeval validate <run-dir>` with stable JSON output and documented exit codes

  **Per-attempt external measurements**

  **Argv-safe variation wrappers**

- [ ] [id:aicjvrtqze] [historical-task-bundles:0.0.15] Support frozen source and structured grading per task

  Scenario-level `fixtures/` gives every task the same tree. Historical coding
  tasks need independent source snapshots, hidden verifier inputs, and patch
  evidence.

- [ ] [id:xshjzqwsrq] Allow each task to select its own frozen source bundle
- [ ] [id:hkznkqljtb] Retain source provenance and digest without fetching during execution
- [ ] [id:xblrrguaqn] Separate agent-visible source from verifier-only and reference artifacts
- [ ] [id:fxaqpjsupc] Derive and retain a normalized patch against the frozen source
- [ ] [id:gacybdrtia] Retain changed-file, added-line, and deleted-line footprint metrics
- [ ] [id:grwshgeniw] Let verifier scripts emit a versioned structured result
- [ ] [id:ekwddtbgwu] Support test, semantic-equivalence, code-review, and footprint dimensions
- [ ] [id:rkgjdfinik] Retain verifier stdout, stderr, exit status, and artifact hashes
- [ ] [id:ctrgqerych] Distinguish verifier infrastructure errors from a valid failing grade
- [ ] [id:waidrjanvx] Keep optional LLM grading separate from deterministic test evidence and agent cost

- [ ] [id:hqneagzpng] [paired-variation-scheduler:0.0.16] Run controlled variation pairs in one campaign

  Separate variation commands execute whole arms sequentially and cannot
  counterbalance temporal or cache effects. Campaign scheduling must pair the
  same task and repetition before execution.

- [ ] [id:tmcpooshed] Accept two or more named variations in one run
- [ ] [id:eurpldgcxv] Define the comparison unit as task plus repetition
- [ ] [id:lbpjucpagj] Use a retained deterministic seed for task ordering
- [ ] [id:uepazzynlp] Counterbalance two-arm execution as AB/BA across pairs
- [ ] [id:ueefaoqluu] Run every scheduled attempt regardless of verifier reward
- [ ] [id:tcmgjwlobw] Retry only infrastructure-invalid attempts under an explicit bounded policy
- [ ] [id:omqddmkmlt] Retain invalid and superseding retry records without replacing evidence
- [ ] [id:isrszmjiut] Default to concurrency one; retain concurrency when explicitly changed
- [ ] [id:zgbvxjuqyp] Add a no-agent preflight that prints the complete schedule and estimated attempt count
- [ ] [id:gdzlryruut] Require explicit execution approval for paid campaigns
- [ ] [id:oubpxfdzbc] Preserve ordinary independent `openeval run` and pass@k behavior

- [ ] [id:naspbbtbil] [sandboxed-attempt-execution:0.0.17] Implement explicit pinned-container runs

  This is an opt-in campaign execution path, not Docker as the default for all
  OpenEval commands.

- [ ] [id:ydotvnaiex] Implement the existing `--image` run flag
- [ ] [id:uetsizgmqv] Pin and retain the image digest used by each attempt
- [ ] [id:ltfzgpupsf] Mount only the attempt workspace, declared artifacts, and read-only credentials
- [ ] [id:jnojozniki] Allocate disposable `HOME` and XDG directories per attempt
- [ ] [id:mvzoajaygn] Make runtime and verifier network policies explicit and separately configurable
- [ ] [id:cyowherhcl] Run deterministic grading without network access by default
- [ ] [id:asvkvkwuui] Bound CPU, memory, process count, and attempt duration
- [ ] [id:jvudhakqhi] Clean up containers after success, failure, timeout, or interruption
- [ ] [id:uwautlqtjn] Retain enough diagnostics to classify setup and execution failures
- [ ] [id:jxtgnyihes] Keep mock and host execution available for tests and local development

- [ ] [id:otwxklxihd] [paired-campaign-report:0.0.18] Add task-paired resource and quality reporting

  Pass@k and aggregate cost do not answer whether a treatment reduced the
  recorded workload or changed the typical task while preserving quality.

- [ ] [id:ycophvcawu] Pair attempts by task, repetition, baseline, and candidate
- [ ] [id:pclrnyejxi] Report exact recorded-workload token and cost totals per repetition
- [ ] [id:kdqsmibdgj] Report paired per-task percentage changes and their geometric mean
- [ ] [id:ktprfgvdbt] Report deterministic paired bootstrap intervals with a retained seed
- [ ] [id:odrqjazmno] Treat repeated runs of one task as one task cluster when pooling
- [ ] [id:kzuvisehza] Report test, equivalence, review, and footprint win/loss/tie counts
- [ ] [id:gmrswsvedj] Report tool-call inflation, duration, patch size, and failure classes
- [ ] [id:gfeypwiceg] Keep runtime-reported cost separate from price-derived estimates
- [ ] [id:spvohfzxoz] Never replace missing measurements with zero
- [ ] [id:fwxubivreo] Annotate unusual attempts without removing them automatically
- [ ] [id:dxmeaqoqic] Emit stable JSON from `compare` as well as human-readable text
- [ ] [id:ecrpncueah] Regenerate campaign summaries byte-identically from retained evidence
- [ ] [id:qdyygxwivr] Label descriptive intervals as uncertainty, not causal or significance claims

- [ ] [id:veyrwvzgug] [ktx-paired-campaign-proof:0.0.19] Prove the generic campaign contract with ktx

  Use ktx as an acceptance vehicle after the generic campaign capabilities
  above ship. OpenEval owns execution and grading; ktx owns compression and its
  native database evidence.

- [ ] [id:pmpknacrog] Configure passthrough and compression arms through the generic variation-command contract with the same OpenCode and ktx instrumentation
- [ ] [id:nfzxmzzlgy] Give each attempt a new disposable `KTX_DB`, `HOME`, and XDG root
- [ ] [id:tszefdfczi] Retain ktx binary, configuration, pipeline, database, and artifact digests
- [ ] [id:lnghhlrfjr] Consume ktx metrics through the generic per-attempt contract rather than one aggregate external map
- [ ] [id:enotbrxuws] Keep OpenCode runtime usage and ktx delivery evidence as separate authorities
- [ ] [id:xtptpxnmwx] Verify unchanged task quality before interpreting resource movement as useful
- [ ] [id:dxupoybfjg] Complete an unpaid or explicitly bounded three-task smoke campaign
- [ ] [id:jfhrrbelvr] Validate every attempt and regenerate the campaign summary before any paid expansion
- [ ] [id:iagegoiuef] Require separate approval before the ten-task, two-repetition, two-arm campaign
- [ ] [id:fsoukojhaw] Document claim limits: local task set, descriptive evidence, no provider-cache or causal claim

---

## Deferred (per decisions and README status)

Do not schedule these until the 0.0.x bar in AGENTS.md is met:

- [ ] [id:qxauqupzff] `0.1.0` / semver release tagging
- [ ] [id:gbavvilvty] Pi harness driver
- [ ] [id:plbfncdaas] OTLP query rollups / `openeval trends`
- [ ] [id:gmataebubq] DeepSWE and margin-eval both production-ready (one only in `0.0.10`)
- [ ] [id:kgrflotmho] Docker packaging as the default run path; explicit opt-in campaign sandboxing is tracked by `sandboxed-attempt-execution:0.0.17`
- [ ] [id:ctrftnacrq] Full cursor-otel-hook parity beyond harness trace correlation
- [ ] [id:fpmsiceutz] [milestone] 2026-06-14: opencode runtime driver plan (4 tasks) complete; see `.agentera/archive/plan-2026-06-14-opencode-runtime-driver.yaml`

## ⇢ Annoying

## ✓ Resolved

- [x] [id:atixjgfqoi] `openeval report` output reflects realistic task names and verifiers
- [x] [id:cyjamdzmqk] Recommend `--rounds 3` for pass@k stability
- [x] [id:dfzuqdppky] Document that a new user with authenticated OpenCode should see a compare table from README alone
- [x] [id:dkvkcglujb] No README path presents mock scores as production eval evidence
- [x] [id:dqnpbigikz] Check supported OpenCode version and `opencode auth list`
- [x] [id:dyohkidynw] Move mock to a secondary path (CI / no API key)
- [x] [id:dzuuyxkxec] README references doctor before first `run --agent opencode`
- [x] [id:ehocruylbs] Quick start order: install → compare flow → optional instrument/traces
- [x] [id:emxrzhhykp] Warn if OTLP endpoint unreachable (non-fatal)
- [x] [id:eulzgmlbvj] Check config at `$XDG_CONFIG_HOME/openeval/config.yaml`
- [x] [id:euzjoamigs] Check native OpenCode OTEL setup; check `~/.cursor/hooks.json` when Cursor is selected
- [x] [id:fmbtxqbvnu] [mini-production-scenario:0.0.3] Expand `example-fixtures` beyond hello-world tasks
- [x] [id:gohxckoruo] Keep shell/script verifiers
- [x] [id:hfxeubpdnz] Exit 0 prints a compare table when OpenCode setup is valid
- [x] [id:hmjzftjgjq] [doctor-command:0.0.6] `openeval doctor`
- [x] [id:iljhlbkurf] CI can run a subset with `--agent mock`
- [x] [id:lmdoedmntt] README section: `skills.aliases` → variation → two runs → compare → `cost_by_skill` in `score.json`
- [x] [id:mewfoiqmle] Retain `default` vs `with-demo-skill` variations for compare
- [x] [id:nsrfgfguxk] Add 3–5 tasks on a small fixture repo (e.g. fix a failing test, apply a lint rule, one-file feature with verifier)
- [x] [id:ofdciouory] [telemetry-shipped-slice:0.0.11] Align Supported telemetry with shipped behavior
- [x] [id:oomnrjbnrh] README: one recommended install path; source build under a separate heading
- [x] [id:ordfmqnxjr] Consistent with AGENTS.md and earn-your-complexity decision
- [x] [id:paiwwqkfdg] Does not require OTLP/Jaeger (see `0.0.4`)
- [x] [id:poezrkzicu] Mock labeled CI/hermetic only
- [x] [id:pqsfojcblt] README does not imply full inventory is available today
- [x] [id:pyfqsqnmhu] Each failure prints remediation steps
- [x] [id:qznnylthow] Check `skills.aliases` paths resolve for skill variations
- [x] [id:sjfpqmxacz] Harness still completes when collector is unreachable (non-blocking export)
- [x] [id:svkvvbwwxv] [demo-command:0.0.2] `openeval demo` for the compare-first flow
- [x] [id:thhdiumqjx] Replace README quick start with the three-command compare flow above
- [x] [id:tufgvozvvf] Trim or relocate Supported telemetry content per `0.0.11`
- [x] [id:uezrpburkb] Missing or invalid config reports copy/fix steps
- [x] [id:ufjbzpfhql] Install docs consistent with AGENTS.md versioning
- [x] [id:urzahlgtnq] [compare-first-quickstart:0.0.1] Compare-first quick start (OpenCode, not mock)
- [x] [id:vgahkmjeen] Document portable `.agents/skills` discovery first and Cursor-only `--plugin-dir` behavior separately
- [x] [id:xzppjsztey] Missing OpenCode or provider authentication reports `opencode auth login` / `opencode auth list` steps
- [x] [id:yecpkdfmqd] Command runs baseline + skill variation + `compare` (or prints commands with `--dry-run`)
- [x] [id:yzmnsefxvi] Compare output documents pass@k, `cost_usd_total`, cost per passed task, and deltas
- [x] [id:zcqhswvoqq] Check `opencode` on PATH or `agents.opencode.command`
- [x] [id:zmmjkcgyts] Split README into **Shipped** vs **Planned** sections
- [x] [id:svshibuuir] [task:0.0.3] Document the model default, overrides, catalog validation, shared-model comparison contract, and author selection factors
- [x] [id:wraukigrax] [task:0.0.3] Record measured one-round onboarding cost/runtime and limit the result to local descriptive evidence
- [x] [id:aksosbmnzi] [task:0.0.3] Complete final plan verification before closing `0.0.3`
- [x] [id:nxoornxntp] `docker compose` under `examples/otel/` (Jaeger on `:4318` / `:16686`), or
- [x] [id:fgeogsyebk] After instrument + collector up, `openeval traces --task … --round …` yields a resolvable Jaeger URL
- [x] [id:ubofpickiu] README trace lookup references the bundled setup
- [x] [id:qdoigmochg] OpenCode native OTEL plus a documented local collector, or
- [x] [id:ndapzuxapr] Documented one-liner in README
- [x] [id:rxpmzvrxjd] [otel-in-a-box:0.0.4] Bundled local OTLP for trace lookup
