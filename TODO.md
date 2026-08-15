# TODO

Product and engineering backlog. Items use **train** tags (`0.0.1`, `0.0.2`, …) to sequence work while the project stays on **0.0.x** — see [AGENTS.md](./AGENTS.md#versioning).

**Intended outcomes**

1. Compare two variation arms and read pass@k and cost deltas from the CLI.
2. Open a harness round trace in Jaeger via `openeval traces`.
3. Run scenarios that resemble maintainer work, not hello-world fixtures.
4. Verify setup with `openeval doctor` and install without cloning the repo.
5. Run at least one integrated external benchmark with documented prerequisites.

**Order:** trains `0.0.1`–`0.0.6` are complete. `0.0.7` consolidates the
compare-and-trace onboarding story; `0.0.8`–`0.0.11` cover install, CI,
scenario authoring, and one external benchmark. Retained attempt evidence grows
through narrow contracts in `0.0.12`–`0.0.19`; sandboxing and paired campaigns
follow in `0.0.20`–`0.0.28`.

## ⇶ Critical

## ⇉ Degraded

## → Normal

## Open trains

- [ ] [id:inclaigcgq] [github-action-recipe:0.0.9] Example GitHub Action workflow
- [ ] [id:kwyfmyzdjh] [task:0.0.9] Add `examples/ci/openeval.yml` (or similar) for consumers
- [ ] [id:cqonolslmc] [task:0.0.9] Fast job: `openeval run --agent mock --rounds 1` on `example-fixtures`, no secrets
- [ ] [id:rvydheplua] [task:0.0.9] Document optional job: OpenCode compare on release branches with provider credentials
- [ ] [id:adnvzovmae] [task:0.0.9] Comment expected cost/runtime for real-agent jobs
- [ ] [id:wazakrjrvw] [task:0.0.9] README links the recipe under CI / release gates
- [ ] [id:ivbsiciloi] [task:0.0.9] Mock job runs on GitHub Actions without cursor credentials

- [ ] [id:ujdmjsiobr] [init-scenario:0.0.10] `openeval init scenario`
- [ ] [id:ygpsvsddjy] [task:0.0.10] Scaffold `scenario.yaml`, `fixtures/`, stub verifier under `verifiers/`
- [ ] [id:xokuvfdxwe] [task:0.0.10] Optional: `--from <repo-path>` with language detection and verifier template
- [ ] [id:nbarkeyirq] [task:0.0.10] Generated layout matches `examples/scenarios/example-fixtures/`
- [ ] [id:jkvedkfuwd] [task:0.0.10] Scaffold runs with `--agent mock` after minimal edits
- [ ] [id:zorqcpiqgh] [task:0.0.10] README custom-scenario section links to init

- [ ] [id:ydbgmcezbt] [external-benchmark:0.0.11] One integrated external scenario end-to-end
- [ ] [id:viyntmpjja] [task:0.0.11] `openeval run --scenario <chosen>` works with documented prerequisites
- [ ] [id:ioaiwazxui] [task:0.0.11] README subsection: dependencies, cost/runtime, default task subset
- [ ] [id:xirsajvynt] [task:0.0.11] Smoke path completes at least one task/round on a real agent
- [ ] [id:yakozwidtm] [task:0.0.11] Missing deps or auth produce actionable errors
- [ ] [id:gnqfxkmklm] [task:0.0.11] Non-chosen scenario remains marked deferred in README

- [ ] [id:fvalnoosno] [attempt-evidence-contract:0.0.12] Retain one structured result per attempt
- [ ] [id:uknptdxbsy] [task:0.0.12] Replace the driver tuple return with a structured attempt result
- [ ] [id:cddhuesuyh] [task:0.0.12] Retain status, process exit, trace/session ID, start/end time, and duration
- [ ] [id:owoitzsfkl] [task:0.0.12] Retain input, output, reasoning, cache-read, and cache-write tokens independently
- [ ] [id:fgyozjkxul] [task:0.0.12] Preserve runtime-reported cost separately from any price-derived estimate
- [ ] [id:mrorwyiive] [task:0.0.12] Retain runtime, version, model, variant, and measurement provenance when available
- [ ] [id:ghniublbux] [task:0.0.12] Preserve explicit numeric zero separately from unavailable or malformed fields

- [ ] [id:hetrvwsjdd] [attempt-failure-evidence:0.0.13] Retain failed attempts as classifiable evidence
- [ ] [id:nvpicjpstq] [task:0.0.13] Retain agent stdout and stderr as bounded artifacts with SHA-256 digests
- [ ] [id:eqvtduxsiy] [task:0.0.13] Distinguish infrastructure-invalid, completed-quality-fail, and completed-quality-pass
- [ ] [id:herscmcokv] [task:0.0.13] Add a versioned score schema and validate it when loading
- [ ] [id:vpjeddglzo] [task:0.0.13] Keep existing score artifacts readable or document and test the intentional break

- [ ] [id:alknfnxqpd] [immutable-attempt-artifacts:0.0.14] Make attempt execution immutable and reproducible
- [ ] [id:fkdegzmmjc] [task:0.0.14] Give each run and attempt a unique immutable ID
- [ ] [id:hcjukgnnlb] [task:0.0.14] Never remove an existing run directory implicitly
- [ ] [id:qivouxgcwv] [task:0.0.14] Retain a campaign manifest before agent execution
- [ ] [id:rimpwejcmi] [task:0.0.14] Hash the scenario, task, verifier, source fixture, variation, skills, environment contract, and effective agent argv
- [ ] [id:ujaqmqbogk] [task:0.0.14] Retain an artifact manifest with path, size, and SHA-256
- [ ] [id:llkmbsukkm] [task:0.0.14] Record execution order, host/runtime identity, and OpenEval version
- [ ] [id:usnfolrcog] [task:0.0.14] Validate that score rows and artifacts belong to the declared campaign
- [ ] [id:xaqpyeefwf] [task:0.0.14] Add `openeval validate <run-dir>` with stable JSON output and documented exit codes

- [ ] [id:bgmqlctjbo] [external-measurement-contract:0.0.15] Model producer-defined measurements per attempt
- [ ] [id:fjcwycnppn] [task:0.0.15] Require every external measurement to declare a state and unit; measured values must be finite numbers, while unavailable and malformed values omit the value
- [ ] [id:kuaykibjzv] [task:0.0.15] Distinguish a missing external artifact, unavailable measurement, malformed measurement, and measured zero
- [ ] [id:jmfuavefdq] [task:0.0.15] Keep runtime usage, runtime cost, and producer measurements as separate authorities
- [ ] [id:whlwclqsqg] [task:0.0.15] Do not add an aggregate external-measurement map to the variation-level telemetry summary

- [ ] [id:pflsygwbgp] [external-measurement-ingestion:0.0.16] Ingest one bounded external measurement artifact per attempt
- [ ] [id:oherwvaici] [task:0.0.16] Remove any existing entry at the metrics path before starting the child process
- [ ] [id:ejfkfczcbt] [task:0.0.16] Read the artifact after the child exits; accept absence without synthesizing measurements or success
- [ ] [id:xmlwovczys] [task:0.0.16] Validate a bounded regular non-symlink file, UTF-8, one complete JSON document without trailing data, and the `openeval.external-metrics.v1` schema
- [ ] [id:qlvovfdclj] [task:0.0.16] Validate required producer metadata, unique measurement names, states, units, required or forbidden values by state, and finite numeric values
- [ ] [id:ykstfajfun] [task:0.0.16] Retain malformed artifacts and validation errors as attempt evidence rather than zero-valued measurements
- [ ] [id:evyzurfgav] [task:0.0.16] Retain exact artifact bytes, relative path, size, and SHA-256 and attach parsed valid measurements to the exact attempt
- [ ] [id:pifccfzezx] [task:0.0.16] Keep the raw per-attempt artifact authoritative; summaries and reports are derived views

- [ ] [id:hqbhbeigvz] [variation-command:0.0.17] Support argv-safe per-variation wrappers
- [ ] [id:ssudlpqfem] [task:0.0.17] Preserve the existing direct-agent invocation when `command` is absent or empty
- [ ] [id:gpfhmoktpt] [task:0.0.17] Reject an empty wrapper executable or empty argv entry before agent execution
- [ ] [id:snssylenxo] [task:0.0.17] Prepend the wrapper without a shell, whitespace splitting, interpolation, or environment expansion
- [ ] [id:ytmuvxwgld] [task:0.0.17] Keep the selected runtime driver responsible for the resolved agent binary and its normal arguments
- [ ] [id:cwmjzaenug] [task:0.0.17] Retain the resolved agent identity and effective argv for direct and wrapped attempts
- [ ] [id:onlagrkdlq] [task:0.0.17] Test executable paths and arguments containing spaces
- [ ] [id:oojmegqtyh] [task:0.0.17] Keep existing scenarios without `command` byte-compatible when loaded

- [ ] [id:aicjvrtqze] [historical-task-bundles:0.0.18] Support frozen source bundles per task
- [ ] [id:xshjzqwsrq] [task:0.0.18] Allow each task to select its own frozen source bundle
- [ ] [id:hkznkqljtb] [task:0.0.18] Retain source provenance and digest without fetching during execution
- [ ] [id:xblrrguaqn] [task:0.0.18] Separate agent-visible source from verifier-only and reference artifacts
- [ ] [id:fxaqpjsupc] [task:0.0.18] Derive and retain a normalized patch against the frozen source
- [ ] [id:gacybdrtia] [task:0.0.18] Retain changed-file, added-line, and deleted-line footprint metrics

- [ ] [id:grwshgeniw] [structured-grading:0.0.19] Retain structured verifier evidence per task
- [ ] [id:ekwddtbgwu] [task:0.0.19] Support test, semantic-equivalence, code-review, and footprint dimensions
- [ ] [id:rkgjdfinik] [task:0.0.19] Retain verifier stdout, stderr, exit status, and artifact hashes
- [ ] [id:ctrgqerych] [task:0.0.19] Distinguish verifier infrastructure errors from a valid failing grade
- [ ] [id:waidrjanvx] [task:0.0.19] Keep optional LLM grading separate from deterministic test evidence and agent cost

- [ ] [id:hqneagzpng] [paired-variation-scheduler:0.0.20] Run controlled variation pairs in one campaign
- [ ] [id:tmcpooshed] [task:0.0.20] Accept two or more named variations in one run
- [ ] [id:eurpldgcxv] [task:0.0.20] Define the comparison unit as task plus repetition
- [ ] [id:lbpjucpagj] [task:0.0.20] Use a retained deterministic seed for task ordering
- [ ] [id:uepazzynlp] [task:0.0.20] Counterbalance two-arm execution as AB/BA across pairs
- [ ] [id:ueefaoqluu] [task:0.0.20] Run every scheduled attempt regardless of verifier reward
- [ ] [id:isrszmjiut] [task:0.0.20] Default to concurrency one; retain concurrency when explicitly changed
- [ ] [id:oubpxfdzbc] [task:0.0.20] Preserve ordinary independent `openeval run` and pass@k behavior

- [ ] [id:tcmgjwlobw] [campaign-safety-controls:0.0.21] Add bounded campaign retry and approval controls
- [ ] [id:omqddmkmlt] [task:0.0.21] Retain invalid and superseding retry records without replacing evidence
- [ ] [id:zgbvxjuqyp] [task:0.0.21] Add a no-agent preflight that prints the complete schedule and estimated attempt count
- [ ] [id:gdzlryruut] [task:0.0.21] Require explicit execution approval for paid campaigns

- [ ] [id:naspbbtbil] [sandboxed-attempt-execution:0.0.22] Implement explicit pinned-container runs
- [ ] [id:ydotvnaiex] [task:0.0.22] Implement the existing `--image` run flag
- [ ] [id:uetsizgmqv] [task:0.0.22] Pin and retain the image digest used by each attempt
- [ ] [id:ltfzgpupsf] [task:0.0.22] Mount only the attempt workspace, declared artifacts, and read-only credentials
- [ ] [id:jnojozniki] [task:0.0.22] Allocate disposable `HOME` and XDG directories per attempt
- [ ] [id:mvzoajaygn] [task:0.0.22] Make runtime and verifier network policies explicit and separately configurable

- [ ] [id:cyowherhcl] [sandbox-limits:0.0.23] Enforce deterministic sandbox limits and cleanup
- [ ] [id:asvkvkwuui] [task:0.0.23] Bound CPU, memory, process count, and attempt duration
- [ ] [id:jvudhakqhi] [task:0.0.23] Clean up containers after success, failure, timeout, or interruption
- [ ] [id:uwautlqtjn] [task:0.0.23] Retain enough diagnostics to classify setup and execution failures
- [ ] [id:jxtgnyihes] [task:0.0.23] Keep mock and host execution available for tests and local development

- [ ] [id:otwxklxihd] [paired-campaign-report:0.0.24] Report paired task metrics and uncertainty
- [ ] [id:ycophvcawu] [task:0.0.24] Pair attempts by task, repetition, baseline, and candidate
- [ ] [id:pclrnyejxi] [task:0.0.24] Report exact recorded-workload token and cost totals per repetition
- [ ] [id:kdqsmibdgj] [task:0.0.24] Report paired per-task percentage changes and their geometric mean
- [ ] [id:ktprfgvdbt] [task:0.0.24] Report deterministic paired bootstrap intervals with a retained seed
- [ ] [id:odrqjazmno] [task:0.0.24] Treat repeated runs of one task as one task cluster when pooling

- [ ] [id:kzuvisehza] [campaign-outcome-report:0.0.25] Report paired quality, footprint, and failure outcomes
- [ ] [id:gmrswsvedj] [task:0.0.25] Report tool-call inflation, duration, patch size, and failure classes
- [ ] [id:gfeypwiceg] [task:0.0.25] Keep runtime-reported cost separate from price-derived estimates
- [ ] [id:spvohfzxoz] [task:0.0.25] Never replace missing measurements with zero
- [ ] [id:fwxubivreo] [task:0.0.25] Annotate unusual attempts without removing them automatically
- [ ] [id:dxmeaqoqic] [task:0.0.25] Emit stable JSON from `compare` as well as human-readable text
- [ ] [id:ecrpncueah] [task:0.0.25] Regenerate campaign summaries byte-identically from retained evidence
- [ ] [id:qdyygxwivr] [task:0.0.25] Label descriptive intervals as uncertainty, not causal or significance claims

- [ ] [id:jjwkabxrwl] [measurement-alignment:0.0.26] Align external measurements without inventing counterparts
- [ ] [id:nnsrwssdgc] [task:0.0.26] Report unmatched, unavailable, malformed, and absent external evidence without inventing counterparts
- [ ] [id:qzwovgoweh] [task:0.0.26] Do not sum ratios, average counts across unequal task sets, combine external values with runtime usage, or infer that a producer caused runtime movement

- [ ] [id:pmpknacrog] [ktx-campaign-integration:0.0.27] Integrate ktx through generic campaign contracts
- [ ] [id:nfzxmzzlgy] [task:0.0.27] Give each attempt a new disposable `KTX_DB`, `HOME`, and XDG root
- [ ] [id:tszefdfczi] [task:0.0.27] Retain ktx binary, configuration, pipeline, database, and artifact digests
- [ ] [id:lnghhlrfjr] [task:0.0.27] Consume ktx metrics through the generic per-attempt contract rather than one aggregate external map
- [ ] [id:enotbrxuws] [task:0.0.27] Keep OpenCode runtime usage and ktx delivery evidence as separate authorities
- [ ] [id:lcnujtldrb] [task:0.0.27] Have both arms publish versioned per-attempt external measurement artifacts

- [ ] [id:veyrwvzgug] [ktx-paired-campaign-proof:0.0.28] Prove the generic campaign contract with ktx
- [ ] [id:xtptpxnmwx] [task:0.0.28] Verify unchanged task quality before interpreting resource movement as useful
- [ ] [id:dxupoybfjg] [task:0.0.28] Complete an unpaid or explicitly bounded three-task smoke campaign
- [ ] [id:jfhrrbelvr] [task:0.0.28] Validate every attempt and regenerate the campaign summary before any paid expansion
- [ ] [id:iagegoiuef] [task:0.0.28] Require separate approval before the ten-task, two-repetition, two-arm campaign
- [ ] [id:fsoukojhaw] [task:0.0.28] Document claim limits: local task set, descriptive evidence, no provider-cache or causal claim
- [ ] [id:muahxyfcry] [task:0.0.28] Verify that OpenEval validates, retains, and compares ktx evidence without interpreting ktx-specific measurement names
- [ ] [id:ocekyawtvh] [task:0.0.28] Verify that missing or unavailable ktx evidence does not become passthrough, success, or zero

---

## Deferred (per decisions and README status)

Do not schedule these until the 0.0.x bar in AGENTS.md is met:

- [ ] [id:qxauqupzff] `0.1.0` / semver release tagging
- [ ] [id:gbavvilvty] Pi harness driver
- [ ] [id:plbfncdaas] OTLP query rollups / `openeval trends`
- [ ] [id:gmataebubq] DeepSWE and margin-eval both production-ready (one only in `0.0.11`)
- [ ] [id:kgrflotmho] Docker packaging as the default run path; explicit opt-in campaign sandboxing is tracked by `sandboxed-attempt-execution:0.0.22`
- [ ] [id:ctrftnacrq] Full cursor-otel-hook parity beyond harness trace correlation

## ⇢ Annoying

## ✓ Resolved

- [x] [id:aksosbmnzi] [task:0.0.3] Complete final plan verification before closing `0.0.3`
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
- [x] [id:fgeogsyebk] After instrument + collector up, `openeval traces --task … --round …` yields a resolvable Jaeger URL
- [x] [id:fmbtxqbvnu] [mini-production-scenario:0.0.3] Expand `example-fixtures` beyond hello-world tasks
- [x] [id:fpmsiceutz] [milestone] 2026-06-14: opencode runtime driver plan (4 tasks) complete; see `.agentera/archive/plan-2026-06-14-opencode-runtime-driver.yaml`
- [x] [id:gohxckoruo] Keep shell/script verifiers
- [x] [id:hfxeubpdnz] Exit 0 prints a compare table when OpenCode setup is valid
- [x] [id:hmjzftjgjq] [doctor-command:0.0.5] `openeval doctor`
- [x] [id:hnarslhsdo] Document with Jaeger screenshots in README
- [x] [id:iljhlbkurf] CI can run a subset with `--agent mock`
- [x] [id:irsuinecko] Cross-link `0.0.1` and `0.0.3` when complete
- [x] [id:lmdoedmntt] README section: `skills.aliases` → variation → two runs → compare → `cost_by_skill` in `score.json`
- [x] [id:mewfoiqmle] Retain `default` vs `with-demo-skill` variations for compare
- [x] [id:mplkdxhdhj] [skill-compare-playbook:0.0.7] Document skill-variation compare workflow
- [x] [id:ndapzuxapr] Documented one-liner in README
- [x] [id:nsrfgfguxk] Add 3–5 tasks on a small fixture repo (e.g. fix a failing test, apply a lint rule, one-file feature with verifier)
- [x] [id:nxoornxntp] `docker compose` under `examples/otel/` (Jaeger on `:4318` / `:16686`), or
- [x] [id:ofdciouory] [telemetry-shipped-slice:0.0.6] Align Supported telemetry with shipped behavior
- [x] [id:oomnrjbnrh] README: one recommended install path; source build under a separate heading
- [x] [id:ordfmqnxjr] Consistent with AGENTS.md and earn-your-complexity decision
- [x] [id:paiwwqkfdg] Does not require OTLP/Jaeger (see `0.0.4`)
- [x] [id:pbmhjkpptz] Ship hook export for session cost, skills active, and tool calls
- [x] [id:poezrkzicu] Mock labeled CI/hermetic only
- [x] [id:pqsfojcblt] README does not imply full inventory is available today
- [x] [id:pyfqsqnmhu] Each failure prints remediation steps
- [x] [id:qdoigmochg] OpenCode native OTEL plus a documented local collector, or
- [x] [id:qldvdzoqzo] Opening sections show compare table example and trace lookup flow
- [x] [id:qznnylthow] Check `skills.aliases` paths resolve for skill variations
- [x] [id:rxpmzvrxjd] [otel-in-a-box:0.0.4] Bundled local OTLP for trace lookup
- [x] [id:sjfpqmxacz] Harness still completes when collector is unreachable (non-blocking export)
- [x] [id:svkvvbwwxv] [demo-command:0.0.2] `openeval demo` for the compare-first flow
- [x] [id:svshibuuir] [task:0.0.3] Document the model default, overrides, catalog validation, shared-model comparison contract, and author selection factors
- [x] [id:thhdiumqjx] Replace README quick start with the three-command compare flow above
- [x] [id:tufgvozvvf] Trim or relocate Supported telemetry content per `0.0.6`
- [x] [id:ubofpickiu] README trace lookup references the bundled setup
- [x] [id:uezrpburkb] Missing or invalid config reports copy/fix steps
- [x] [id:ufjbzpfhql] Install docs consistent with AGENTS.md versioning
- [x] [id:urzahlgtnq] [compare-first-quickstart:0.0.1] Compare-first quick start (OpenCode, not mock)
- [x] [id:vgahkmjeen] Document portable `.agents/skills` discovery first and Cursor-only `--plugin-dir` behavior separately
- [x] [id:wraukigrax] [task:0.0.3] Record measured one-round onboarding cost/runtime and limit the result to local descriptive evidence
- [x] [id:xzppjsztey] Missing OpenCode or provider authentication reports `opencode auth login` / `opencode auth list` steps
- [x] [id:yecpkdfmqd] Command runs baseline + skill variation + `compare` (or prints commands with `--dry-run`)
- [x] [id:yzmnsefxvi] Compare output documents pass@k, `cost_usd_total`, cost per passed task, and deltas
- [x] [id:zcqhswvoqq] Check `opencode` on PATH or `agents.opencode.command`
- [x] [id:zmmjkcgyts] Split README into **Shipped** vs **Planned** sections
- [x] [id:chrlpdaedf] [compare-hero-readme:0.0.7] Make compare and trace lookup the README hero
- [x] [id:ivntyvxwqc] [task:0.0.8] Verify and document `go install github.com/jgabor/openeval/cmd/openeval@latest` (or correct module path)
- [x] [id:uemgcurert] [task:0.0.8] Optional: curl script or Homebrew tap
- [x] [id:vwenelnkgw] [frictionless-install:0.0.8] Single documented install path without cloning
