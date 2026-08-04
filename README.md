# OpenEval

OpenEval is one CLI control plane for two related workflows:

- **Harness proof:** run pinned tasks, verify outcomes, compare pass@k and cost.
- **Continuous instrumentation:** export selected runtime events to an external OTLP backend.

OpenCode is the primary production runtime. Cursor is a supported secondary runtime. The mock driver exists for hermetic tests and CI; it does not produce production evaluation evidence.

> **Status: `0.0.0-dev`.** OpenEval is not on a stable release track. The shipped vertical slice is the local CLI, the `example-fixtures` scenario, OpenCode and Cursor harness drivers, comparison/reporting, Cursor hooks, and opt-in native OpenCode OTEL. A public no-clone install, Pi, external benchmark integrations, Docker execution, a bundled collector, and full cross-runtime telemetry normalization are not shipped.

## Support status

| Surface                    | Status                 | Notes                                                                                                            |
| -------------------------- | ---------------------- | ---------------------------------------------------------------------------------------------------------------- |
| OpenCode harness           | **Primary, shipped**   | Validated with OpenCode 1.18.11; autonomous JSONL runs, comparable cost estimates, portable skill checks         |
| Cursor harness             | **Secondary, shipped** | `cursor-agent` subprocess driver and Cursor-specific `.claude-plugin` loading                                    |
| Mock harness               | **CI only, shipped**   | Hermetic synthetic driver; never treat mock scores as agent-quality evidence                                     |
| OpenCode native OTEL       | **Opt-in, shipped**    | Separate runtime traces correlated to OpenEval summary traces by resource attributes                             |
| Cursor hook telemetry      | **Secondary, shipped** | Merged into `~/.cursor/hooks.json` without replacing existing hooks                                              |
| Pi                         | **Planned**            | No harness or telemetry driver                                                                                   |
| DeepSWE and margin-eval    | **Deferred**           | Names are reserved, but `run` reports that integration is not implemented                                        |
| Docker `--image` execution | **Deferred**           | The flag reports that packaged runs are not implemented                                                          |
| Install                    | **Checkout only**      | `go install ./cmd/openeval`; the repository is not currently available through a public `@latest` module install |
| OTLP storage/collector     | **External**           | OpenEval does not embed a trace database or currently bundle a collector                                         |

## Quick start: compare on OpenCode

Prerequisites:

- Go 1.26 or later.
- OpenCode 1.18.11 on `PATH`.
- Provider credentials configured in OpenCode.
- No OTLP collector is required for the comparison flow.

The current example scenario and skills are file-backed development assets. Repository access is currently required. Clone it so the quick-start commands can resolve those assets, then install the CLI:

```bash
git clone https://github.com/jgabor/openeval.git
cd openeval
go install ./cmd/openeval
```

Authenticate and diagnose setup. These commands do not make a paid model call:

```bash
opencode auth login
opencode auth list
openeval doctor --agent opencode
```

Preview the exact work and retained output paths:

```bash
openeval demo --dry-run
```

Run the quick-start comparison with one attempt per task in each arm:

```bash
openeval demo
```

`demo` runs `baseline` and `with-demo-skill`, passes the actual returned run directories to `compare`, prints the comparison table, and prints one unique evidence root. It never removes an earlier demo root. The one-round default keeps onboarding quick; use `--rounds 3` for stronger pass@k evidence.

For planning, allow about five minutes and about $3 for the default one-round, two-arm demo. One measured full-corpus run with OpenCode 1.18.11 and `opencode/big-pickle` passed 4/4 tasks in each arm, took 248.852 seconds, and had an estimated combined cost of $2.822868 with no harness failures. This is one local sample, not a runtime, cost, or result variance guarantee.

The table includes:

- `pass@1`
- `pass@<rounds>`
- `cost_usd_total`
- `cost_per_passed`
- candidate-minus-baseline deltas
- exact `score.json` paths

### Manual three-command comparison

Use a fresh root for each experiment. The explicit `--out` paths avoid inferred or variation-default directories.

```bash
ROOT=./scenarios/example-fixtures/manual-compare-001

openeval run --scenario example-fixtures --variation baseline \
  --agent opencode --rounds 3 --out "$ROOT/baseline"

openeval run --scenario example-fixtures --variation with-demo-skill \
  --agent opencode --rounds 3 --out "$ROOT/with-demo-skill"

openeval compare "$ROOT/baseline" "$ROOT/with-demo-skill"
```

Increment the root suffix for later experiments. `openeval demo` allocates this unique root automatically.

## Install path

The verified install path is from a repository checkout:

```bash
go install ./cmd/openeval
```

`go install github.com/jgabor/openeval/cmd/openeval@latest` is the intended future public path, but it is not currently a working unauthenticated install: the module is not publicly resolvable. The example scenario and skills are also file-backed rather than embedded. Do not present the future command as a shipped installation method.

For contributor work, the repository also provides Mage:

```bash
git clone https://github.com/jgabor/openeval.git
cd openeval
go install github.com/magefile/mage@v1.17.2
mage install
```

Useful checks:

```bash
mage build
mage test
mage lint
mage check
```

`mage check` runs the tidy check, tests, vet, lint, and `govulncheck`.

## Commands

| Command               | Purpose                                                                              |
| --------------------- | ------------------------------------------------------------------------------------ |
| `openeval doctor`     | Read-only setup checks; OpenCode is the default; `--json` emits `openeval.doctor.v1` |
| `openeval demo`       | Diagnose, run baseline and skill arms, compare, and retain unique output             |
| `openeval run`        | Execute one scenario variation against `opencode`, `cursor`, or `mock`               |
| `openeval compare`    | Compare two run directories with the same scenario ID                                |
| `openeval report`     | Print one run's task, pass@k, cost, and session summary                              |
| `openeval traces`     | Print the direct summary trace target and optional native-correlation help           |
| `openeval instrument` | Opt into OpenCode native OTEL or merge Cursor hooks                                  |
| `openeval hook`       | Receive Cursor hook callbacks and emit configured summary events                     |

Run `openeval <command> --help` for flags.

## Doctor contract

`openeval doctor` makes no paid model request. For OpenCode it checks:

1. The OpenEval config path and YAML.
2. `opencode` on `PATH` or `agents.opencode.command`.
3. The OpenCode version against the validated 1.18.11 contract.
4. `opencode auth list`.
5. The resolved model appears in the authenticated `opencode models` catalog.
6. Every `skills.aliases` path and OpenCode skill frontmatter.
7. OTLP endpoint reachability.
8. Native OpenCode OTEL opt-in state.

An unreachable collector is a warning and exits 0. Invalid config, runtime, authentication, or skill setup is fatal and exits 1. Every non-pass result includes remediation. Use stable structured output in automation:

```bash
openeval doctor --agent opencode --json
```

Cursor remains available as a secondary check:

```bash
openeval doctor --agent cursor
```

It checks `cursor-agent`, its version output, shared config/skills/telemetry, and OpenEval entries in `~/.cursor/hooks.json`.

## Configuration

OpenEval reads:

- Linux and macOS: `${XDG_CONFIG_HOME:-$HOME/.config}/openeval/config.yaml`
- Windows: `%LOCALAPPDATA%\openeval\config.yaml`

`openeval instrument --agent opencode` or `--agent cursor` creates the config if it is absent. The shipped example is `examples/config.minimal.yaml`:

```yaml
version: 1
telemetry:
  protocol: http/json
  endpoint: http://localhost:4318/v1/traces
  insecure: true
privacy:
  mask_prompts: true
  mask_secrets: true
scenarios:
  aliases: {}
skills:
  aliases:
    demo-skill: ./examples/skills/demo-skill
    plugin-skill: ./examples/skills/plugin-skill
agents:
  opencode:
    # command: /usr/local/bin/opencode
    # Select deliberately for your workload; this is also the built-in fallback.
    model: opencode/big-pickle
    native_otel: false
    cost:
      input_per_million: 3.0
      output_per_million: 15.0
  cursor:
    # command: /usr/local/bin/cursor-agent
    cost:
      input_per_million: 3.0
      output_per_million: 15.0
```

OpenCode and Cursor use the same configured input/output rates. OpenCode reasoning tokens use the output rate; cache-read and cache-write tokens use the input rate. OpenEval deliberately does not substitute OpenCode's runtime-reported cost, so cross-runtime estimates stay comparable.

### OpenCode model policy

OpenEval resolves the OpenCode model in this order:

```text
--model CLI flag > scenario model > agents.opencode.model > opencode/big-pickle
```

Use `openeval run --model provider/model` for one run or `openeval demo --model provider/model` for both demo arms. A scenario author can set top-level `model: provider/model`; a user-wide default belongs under `agents.opencode.model` in the config. `openeval doctor --agent opencode --model provider/model` validates the selection against the authenticated OpenCode catalog without making a paid model call.

`opencode/big-pickle` is a convenience default, not a claim that it fits every evaluation. Scenario and benchmark authors should select explicitly based on workload fit, required edit and shell tools, provider authentication, expected cost, latency, and their data-handling policy. Configure provider credentials through OpenCode; do not put credentials in scenario or config files committed to the repository. The selected model must support the tools needed to inspect and edit files and run shell commands for the shipped maintainer scenario.

Keep the model the same across configuration arms, such as baseline versus skill, so the intended configuration is the controlled difference. Use different models only when model comparison is the stated experiment. A fixed `provider/model` selector does not make outputs deterministic and does not freeze the provider's backend or deployment.

Runtime environment precedence is deterministic:

```text
inherited process environment < variation env < reserved OpenEval run context
```

Variations cannot replace `OPENEVAL_SCENARIO_ID`, `OPENEVAL_VARIATION`, `OPENEVAL_TASK_ID`, `OPENEVAL_ROUND`, or `OPENEVAL_TRACE_ID`.

## Scenarios, variations, and evidence

The shipped `example-fixtures` scenario contains four independent maintenance tasks in one small standard-library-only Python repository: duration parsing, URL credential redaction, account-name normalization, and log-level summaries. Each task has a focused, scenario-owned grader outside the copied agent workspace and needs no package installation, network access, external service, or repository history. Grader assertions and verdict control stay in the parent process; a child process imports workspace code and returns normalized call results. This process boundary is not a campaign sandbox and does not prevent arbitrary host filesystem mutation. The scenario has three relevant arms:

```yaml
variations:
  default: {}
  baseline: {}
  with-demo-skill:
    skills: [demo-skill]
```

Differences between the baseline and skill arms describe the outcomes observed in those local runs. They do not prove that the skill caused the difference; sampling and other uncontrolled runtime factors can also affect results. The four small tasks are descriptive onboarding evidence, not a general ranking of providers or models.

For a fast credential-free automation check, use the one-task subset with the mock agent:

```bash
openeval run --scenario ./examples/scenarios/example-fixtures/edit-file-only.yaml \
  --agent mock --rounds 1
```

A scenario file contains task prompts, script verifiers, and named variations:

```yaml
id: my-scenario
tasks:
  - id: edit-file
    prompt: Add a one-line comment without changing behavior
    verifier:
      type: script
      run: ./verifiers/edit-file.sh
variations:
  baseline: {}
  candidate:
    env:
      FEATURE_ENABLED: "true"
    skills:
      - my-skill
```

From a repository checkout, pass the shipped ID. You can also pass an alias from `scenarios.aliases` or any YAML path:

```bash
openeval run --scenario ./evals/my-scenario.yaml \
  --variation baseline --agent opencode --rounds 3 \
  --out ./scenarios/my-scenario/experiment-001/baseline
```

Each task round gets an isolated workspace. OpenEval copies the scenario's `fixtures/` directory into `workspace/fixtures`, runs the agent in that workspace, then runs the source-controlled verifier against that workspace only after a successful agent event stream. Workspace files cannot replace the scenario-owned grading assertions.

### Output paths

- Default or omitted variation without `--out`: `scenarios/<id>/runs/<n>`; numeric runs increment.
- Named variation without `--out`: `scenarios/<id>/runs/<variation>`; the legacy path is replaced on rerun.
- Explicit `--out`: exactly that directory; use a fresh path for retained manual evidence.
- `demo`: a unique timestamp-plus-random root with `baseline/` and `with-demo-skill/` children.

Generated `scenarios/**/runs/` and demo evidence are local artifacts. Do not commit them unless you deliberately want fixtures.

## Portable skills

OpenEval resolves each variation skill from `skills.aliases` or `examples/skills/<name>` and copies it to:

```text
<round-workspace>/.agents/skills/<name>/SKILL.md
```

Before an OpenCode model call, OpenEval runs `opencode debug skill` in an isolated directory to reject a same-named global skill, then runs it in the seeded workspace. The requested name must resolve to the exact local `SKILL.md`. Invalid frontmatter, hidden skills, global collisions, and unexpected locations stop before paid execution.

Cursor uses the same workspace copy. If a source skill also contains `.claude-plugin`, only the Cursor driver adds its runtime-specific `--plugin-dir` argument. OpenCode does not emulate Cursor plugins or hooks.

## Score, report, and compare

Each run writes `score.json` with schema `openeval.score.v1`. It includes:

- scenario, agent, resolved model when known, variation, rounds, and task count
- per-task and per-round verifier results
- pass@k summaries
- estimated USD cost per round and aggregate cost per passed task
- one OpenEval trace ID per round
- active skill names and cost attribution by active skill

The current score schema does not retain OpenCode's input/output/reasoning/cache token classes separately. That richer attempt evidence remains planned.

```bash
openeval report <run-dir>
openeval compare <baseline-run-dir> <candidate-run-dir>
```

`report` prints the retained model when known. `compare` requires matching `scenario_id` values and warns if agents or known models differ. A model warning does not block an intentional model comparison; a missing model in a legacy score remains unknown rather than being treated as a mismatch.

## Traces and telemetry

Harness execution does not require a collector. OpenEval writes trace IDs and scores even when the configured OTLP endpoint is unreachable. Span export is best effort and nonblocking.

### Shipped now

- One direct OpenEval summary span per successful harness round, sent as OTLP HTTP/JSON to `telemetry.endpoint`.
- Summary attributes for estimated cost, active skills, the prompt-masking flag, scenario, variation, task, and round.
- Cursor lifecycle/tool hook events through merged `~/.cursor/hooks.json` entries.
- Explicit opt-in native OpenCode OTEL using standard inherited exporter headers/resource attributes.
- Trace help that distinguishes direct summary trace IDs from separately generated native traces.

Look up a round:

```bash
openeval traces <run-dir> --task parse-duration-units --round 1
```

The direct `trace_id` and Jaeger URL target the OpenEval summary span. If native OpenCode OTEL is enabled, `traces` also prints:

- the native OTLP receiver path
- `openeval.trace_id=<summary-trace-id>` as the resource-attribute search key
- an explicit statement that native and summary traces are separate trace trees

### Opt into native OpenCode OTEL

```bash
openeval instrument --agent opencode
```

This sets `agents.opencode.native_otel: true`. For harness runs, OpenEval:

1. Converts the configured summary endpoint such as `http://localhost:4318/v1/traces` to the base endpoint `http://localhost:4318` because OpenCode appends `/v1/traces`.
2. Preserves inherited `OTEL_EXPORTER_OTLP_HEADERS` and non-OpenEval `OTEL_RESOURCE_ATTRIBUTES`.
3. Appends URL-encoded scenario, variation, task, round, and OpenEval trace ID attributes.

`instrument` prints the `OTEL_EXPORTER_OTLP_ENDPOINT` export needed for OpenCode sessions started outside `openeval run`.

### Native telemetry privacy boundary

OpenCode creates native span payloads inside its own runtime. Those payloads can include runtime-generated prompt, tool, or model data. **OpenEval cannot apply `privacy.mask_prompts` or `privacy.mask_secrets` to native OpenCode payloads.** Review OpenCode's telemetry and your collector policy before opting in.

OpenEval's direct summary spans remain the stable trace targets recorded in `score.json`; native traces are separate and correlated only by resource attributes. Do not assume one shared parent/child trace tree.

Native span flushing on non-interactive OpenCode process exit is not verified by OpenEval's default no-collector path. Verify receipt against your collector before depending on native traces. The summary trace ID and score path remain available regardless.

### Planned or deferred telemetry

- Separate retained input, output, reasoning, cache-read, and cache-write counts in attempt evidence.
- Full cross-runtime tool, prompt, context, and reasoning normalization.
- Longitudinal queries and dashboards in an external span store.
- A bundled local collector/Jaeger setup.
- Pi telemetry.

OpenEval will not embed a trace database in the CLI binary.

## Cursor secondary path

Install and authenticate the headless `cursor-agent` binary, then diagnose it:

```bash
cursor-agent login
openeval doctor --agent cursor
openeval run --scenario example-fixtures --variation baseline \
  --agent cursor --rounds 3 --out ./scenarios/example-fixtures/cursor-001/baseline
```

Optional continuous instrumentation:

```bash
openeval instrument --agent cursor
```

This merges OpenEval callbacks into existing Cursor hooks and writes Cursor's legacy OpenEval hook config. It does not replace unrelated hook entries.

## Mock for CI

Use mock only to test scenario plumbing, verifiers, score serialization, and comparison automation without credentials:

```bash
openeval demo --agent mock --rounds 1
```

The mock driver writes synthetic artifacts and cost. A green mock comparison proves the harness wiring, not coding-agent quality.

Repository tests use mock, shell stubs, helper processes, and local test servers. They never require a live agent, provider credential, or collector.

## Deferred scenarios and packaging

DeepSWE and margin-eval are not integrated. Current invocations fail with a direct deferred-integration error rather than producing plausible mock evidence.

Docker execution is also not implemented. `examples/docker/README.md` records the planned interface only; `openeval run --image ...` exits before execution.

## License

[Apache-2.0](./LICENSE) · Version `0.0.0-dev` · Jonathan Gabor ([jgabor.se](https://jgabor.se))
