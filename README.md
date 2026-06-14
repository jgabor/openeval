# OpenEval

One CLI control plane for measuring CLI coding agent quality: **harness proof** on pinned scenarios and **continuous observability** on real sessions.

| Rhythm | Commands | Question it answers |
| ------ | -------- | ------------------- |
| **Harness proof** (primary) | `run`, `compare`, `report`, `traces` | Did pass@k or cost per passed task move on release or on schedule? |
| **Continuous observability** | `instrument`, `hook` (agent callback) | Did real-session cost, tokens, or tool usage drift after a skill change? |

Both rhythms share one config file, one `score.json` contract, and correlated OTLP traces — not two separate products.

- **Run:** execute pinned scenario tasks with verifiers; aggregate pass@k and cost
- **Compare:** diff two variation arms on the same scenario
- **Instrument:** install hooks that export masked spans to your OTLP backend
- **Traces:** link a harness round to the configured collector and Jaeger UI

Run the same scenarios on Cursor, OpenCode, Pi, and other supported agents. Better numbers upstream mean cheaper, more reliable agents for the teams and end users downstream.

> **Status:** `0.0.0-dev` — Core CLI ships in this repo (`mage install` or `go build -o bin/openeval ./cmd/openeval`). Cursor harness + hook instrumentation are implemented; DeepSWE, margin-eval, Docker images, and full cross-runtime telemetry inventory are deferred.

### Terminology

| Term          | Meaning                                                                                            |
| ------------- | -------------------------------------------------------------------------------------------------- |
| **Eval**      | The overall activity: instrument an agent, run a scenario, score and compare results               |
| **Scenario**  | A pinned task set and verifiers (what problems run). Selected with `--scenario`                    |
| **Variation** | A labeled config arm under test (baseline, a version, a feature flag). Selected with `--variation` |
| **Run**       | One execution of a scenario, stored under `./scenarios/<scenario>/runs/…`                          |

Shipped examples live under `examples/`. Your scenarios live under `evals/` (or any path you pass to `--scenario`).

## Quick start

**Primary path:** install the CLI, run the shipped `example-fixtures` scenario, then optionally instrument Cursor for continuous OTLP export.

```bash
mage install   # or: go build -o bin/openeval ./cmd/openeval

# 1. Harness proof — pinned tasks, verifiers, score.json
openeval run --scenario example-fixtures --agent mock --rounds 1
openeval report ./scenarios/example-fixtures/runs/1

# 2. Continuous observability — optional, same config + OTLP endpoint
cp examples/config.minimal.yaml "$XDG_CONFIG_HOME/openeval/config.yaml"
openeval instrument --agent cursor
```

Use `--agent cursor` instead of `mock` when `cursor-agent` is installed and authenticated. See [Cursor](#cursor---agent-cursor) below.

### Prerequisites

- Linux, macOS, or Windows
- Docker (optional, for packaged eval runs)
- An OTLP backend (Jaeger, Grafana Cloud, or any OTLP collector) — for instrumentation and trace lookup
- A supported CLI agent installed locally (Cursor, OpenCode, or Pi) — for production harness runs

### Install

```bash
go install openeval@latest
```

From source:

```bash
git clone https://github.com/jgabor/openeval.git
cd openeval
go install github.com/magefile/mage@v1.17.2
mage install
```

### Instrument (continuous observability)

Install hooks that read the shared OpenEval config and export masked spans to OTLP:

```bash
openeval instrument --agent cursor
```

`openeval instrument` merges hook entries into `~/.cursor/hooks.json` and creates the config file if missing (see [Configuration](#configuration)). Harness runs work without instrumentation; add it when you want real-session cost and behavior traces alongside periodic proof runs.

### Configuration

Config paths:

- Linux and macOS: `$XDG_CONFIG_HOME/openeval/config.yaml`
- Windows: `%LOCALAPPDATA%\openeval\config.yaml`

Copy the example or edit the file `openeval instrument` creates:

```bash
cp examples/config.minimal.yaml "$XDG_CONFIG_HOME/openeval/config.yaml"
```

Minimal example (`examples/config.minimal.yaml`):

```yaml
version: 1
telemetry:
  protocol: http/json
  endpoint: http://localhost:4318/v1/traces
  insecure: true
privacy:
  mask_prompts: true
  mask_secrets: true
```

Telemetry settings follow standard OTLP environment variable names. Hook-instrumented agents read this file, so you do not need export variables in the agent process.

Export options (set in config or via OTLP env vars):

- Protocols: gRPC, HTTP/Protobuf, HTTP/JSON
- Generation-based batching for HTTP/JSON (spans grouped by `generation_id`, sent on stop event)
- Automated install scripts for Windows and macOS/Linux

### Scenarios and variations

- **`--scenario`** selects which tasks to run: a built-in id (`example-fixtures`), a path to your YAML (`./evals/my-scenario.yaml`), a registered alias, or an integrated external scenario (`deepswe`, `margin-eval`).
- **`--variation`** selects a named config arm defined in the scenario file (or `default` when omitted). Compare two runs that share the same scenario.
- **`--rounds`** sets attempts per task (default: `3`). Use multiple rounds for pass@k; treat a single round as exploratory only.
- **`--out`** is optional. Default layout:
  - no `--variation` (or `default`) → `./scenarios/<scenario>/runs/<n>` with auto-incrementing `n` per scenario; prior runs are kept
  - named `--variation` → `./scenarios/<scenario>/runs/<variation>`; re-running overwrites that directory unless you pass `--out`

`<scenario>` is the resolved scenario id (`example-fixtures`, `deepswe`, …), not the agent name. Agent, rounds, and timestamps live in `score.json` and traces.

**How variations apply:** each scenario YAML lists `variations` with the agent config for that arm — skills to enable, env vars, prompt pins, and similar. `openeval run --variation <name>` loads that block before starting tasks. The variation label is stored in `score.json`; skill names and telemetry reflect what actually ran.

**Compare rules:** `openeval compare` requires the same `scenario_id` in both `score.json` files. It warns when `agent` differs. Pass two run directories; labels come from each run's `variation` field.

Drop instrumentation that never changes scenario scores or cost (see [Supported telemetry](#supported-telemetry) — capture only what informs comparisons).

### Scenario format

Minimal shape (see `examples/scenarios/example-fixtures/scenario.yaml`):

```yaml
id: example-fixtures
tasks:
  - id: hello-verify
    prompt: Verify the hello-world script prints the expected greeting
    verifier: { type: script, run: ./verifiers/hello-verify.sh }
  - id: edit-file
    prompt: Add a one-line comment to fixtures/src/main.py without changing behavior
    verifier: { type: script, run: ./verifiers/edit-file.sh }
variations:
  default: {}
```

### Run

```bash
# Example scenario shipped with the repo → ./scenarios/example-fixtures/runs/1
openeval run \
  --scenario example-fixtures \
  --agent cursor \
  --rounds 3

# Custom scenario you author → ./scenarios/my-scenario/runs/1
openeval run \
  --scenario ./evals/my-scenario.yaml \
  --agent cursor

# Integrated external scenarios
openeval run --scenario deepswe --agent cursor --rounds 5
# → ./scenarios/deepswe/runs/1

openeval run --scenario margin-eval --agent cursor --rounds 3
# → ./scenarios/margin-eval/runs/1

# Fixed package in Docker → ./scenarios/example-fixtures/runs/2
openeval run \
  --scenario example-fixtures \
  --image ./examples/docker/agent-eval \
  --rounds 5
```

### Compare variations

Same scenario, two config arms — then diff pass@k and cost:

```bash
# External scenario: current release vs release candidate
openeval run --scenario deepswe --agent cursor \
  --variation v1.2.3 --rounds 5
# → ./scenarios/deepswe/runs/v1.2.3

openeval run --scenario deepswe --agent cursor \
  --variation v1.3.0-rc1 --rounds 5
# → ./scenarios/deepswe/runs/v1.3.0-rc1

openeval compare \
  ./scenarios/deepswe/runs/v1.2.3 \
  ./scenarios/deepswe/runs/v1.3.0-rc1
```

External scenarios delegate to the upstream runner and normalize scores into the same `score.json` contract.

### Commands

| Command                                             | Purpose                                                                    |
| --------------------------------------------------- | -------------------------------------------------------------------------- |
| `openeval run`                                      | Execute a scenario; writes `score.json` and traces under the run directory |
| `openeval compare <dir-a> <dir-b>`                  | Diff two runs (same `scenario_id` required)                                |
| `openeval report <run-dir>`                         | Human-readable summary of one run                                          |
| `openeval traces <run-dir> --task <id> --round <n>` | Open or print trace links for a task attempt                               |

### Results (example)

From `./scenarios/example-fixtures/runs/1/score.json`:

```json
{
  "schema": "openeval.score.v1",
  "scenario_id": "example-fixtures",
  "agent": "cursor",
  "variation": "",
  "rounds": 1,
  "tasks": 2,
  "summary": {
    "pass_at_1": 1.0,
    "pass_at_3": 1.0,
    "tasks_passed": 2,
    "tasks_total": 2,
    "cost_usd_total": 0.53,
    "cost_usd_per_passed_task": 0.27,
    "tokens_input_total": 0,
    "tokens_output_total": 0
  },
  "by_task": [
    {
      "task_id": "hello-verify",
      "verifier": "pass",
      "rounds": [
        {
          "round": 1,
          "verifier": "pass",
          "cost_usd": 0.21,
          "trace_id": "a1b2c3..."
        }
      ],
      "pass_at_k": { "1": 1.0 }
    },
    {
      "task_id": "edit-file",
      "verifier": "pass",
      "rounds": [
        {
          "round": 1,
          "verifier": "pass",
          "cost_usd": 0.32,
          "trace_id": "d4e5f6..."
        }
      ],
      "pass_at_k": { "1": 1.0 }
    }
  ],
  "telemetry": {
    "skills_active": [],
    "cost_by_skill_usd": {},
    "sessions": 2
  }
}
```

### `openeval report` output (example)

```text
scenario: example-fixtures
agent: cursor
variation: default
rounds: 1

pass@1: 1.00  (2/2 tasks passed on first attempt)
pass@3: 1.00  (2/2 tasks passed within 1 attempt)
cost:   $0.53 total  ($0.27 per passed task)

tasks:
  hello-verify  PASS  pass@1=1.00  cost=$0.21
  edit-file     PASS  pass@1=1.00  cost=$0.32

traces: 2 sessions, OTLP service openeval-agent
        open Jaeger with trace_id from score.json or run:
        openeval traces ./scenarios/example-fixtures/runs/1 --task edit-file --round 1
```

### Trace lookup

`openeval traces` prints the harness round `trace_id`, your configured OTLP endpoint, and a Jaeger UI link (default all-in-one install on port 16686):

```bash
openeval traces ./scenarios/example-fixtures/runs/1 --task edit-file --round 1
```

```text
trace_id: a1b2c3...
otlp_service: openeval-agent
otlp_endpoint: http://localhost:4318/v1/traces
jaeger_ui: http://localhost:16686/trace/<normalized-trace-id>
```

### Skill-backed variations

`example-fixtures` ships skill variations for harness proof. Shipped skills live under `examples/skills/` and resolve from `skills.aliases` in config or from the built-in `examples/skills/<name>` path:

```yaml
variations:
  with-demo-skill:
    skills: [demo-skill]
  with-plugin-skill:
    skills: [plugin-skill]  # passes --plugin-dir when .claude-plugin is present
```

```bash
openeval run --scenario example-fixtures --variation with-demo-skill --agent mock --rounds 1
```

### `openeval compare` output (example)

```text
comparing: v1.2.3 to v1.3.0-rc1
scenario:  deepswe
agent:     cursor

                    v1.2.3    v1.3.0-rc1    delta
pass@1                 0.40          0.55     +0.15
pass@3                 0.60          0.70     +0.10
cost_usd_total         4.20          3.85     -0.35
cost_per_passed        7.00          5.50     -1.50

results:
  summary: pass@3 improved, cost per passed task down
  logs:
    ./scenarios/deepswe/runs/v1.2.3/score.json
    ./scenarios/deepswe/runs/v1.3.0-rc1/score.json
```

## Supported telemetry

Instrumentation explains **why** harness scores moved and flags when to re-run a scenario — it is not a parallel analytics product. Capture only what informs pass@k, cost per passed task, or justified rollups (see Decision: earn your complexity).

When hooks or native OTEL export are enabled, OpenEval normalizes session metadata into OTLP spans that share the same config and trace correlation as harness runs. Planned and partial capture across runtimes:

- Inventory: enabled resources in the session (listed per domain below)
- Metadata: descriptors (names, descriptions, hooks)
- Content: full captured text or payload (privacy-masked by default)
- Size: bytes and tokens

### Resources

- System prompt
  - Content (full prompt text)
  - Size (bytes and tokens)
- Plugins
  - Inventory (enabled plugins)
  - Metadata (names, hooks, events)
- Skills
  - Inventory (enabled skills)
  - Metadata (names, snippets, descriptions)
  - Content (full skill text)
  - Size (bytes and tokens, total and per skill)
- Tools
  - Inventory (enabled tools)
  - Metadata (names, snippets, descriptions)
  - Content (full tool definitions)
  - Size (bytes and tokens, total and per tool)
- Context
  - Messages (including token counts by type: input, output, reasoning, cache)
  - Reasoning traces
  - Referenced files

### Events

- Session lifecycle (start and end)
- Tool calls (before, after, failure)
- Shell commands
- MCP calls
- File read and edit
- Prompt submission
- Context compaction
- Subagent activity

### Aggregates

Totals are exported per dimension. Averages are computed when filtering or comparing runs.

- Cache
  - Hits and misses (input and output)
  - Cost (hits compared to misses)
- Cost
  - Messages (totals and averages)
  - Skills (totals and averages)
  - Tools (totals and averages)
  - Sessions (totals and averages)

Spans nest as session, generation, and hook events. Fields follow [OpenTelemetry GenAI semantic conventions](https://opentelemetry.io/docs/specs/semconv/gen-ai/) where applicable (`gen_ai.tool.definitions`, `gen_ai.usage.*`).

Use this inventory to interpret harness deltas (for example, cost per skill on a variation arm) and to spot drift in real sessions that warrants another `openeval run`. Full cross-runtime capture and longitudinal CLI rollups are deferred.

## Privacy

- Data masking before export (prompts and secrets)
- Rules configurable per deployment

## Scenarios

A scenario pins tasks, verifiers, and runtime versions so runs stay comparable. **`--variation`** is separate: it tags which agent config produced a given run. Compare only runs that share the same `--scenario`.

| Scenario                                           | `--scenario` value                               | Source                                                              |
| -------------------------------------------------- | ------------------------------------------------ | ------------------------------------------------------------------- |
| Example fixtures                                   | `example-fixtures`                               | `examples/scenarios/example-fixtures/`                              |
| Custom                                             | `./evals/my-scenario.yaml` or a registered alias | Your `evals/` directory; not shipped by OpenEval                    |
| [DeepSWE](https://deepswe.datacurve.ai/)           | `deepswe`                                        | Integrated; long-horizon SWE tasks, behavior-based verification     |
| [Margin Eval](https://github.com/Margin-Lab/evals) | `margin-eval`                                    | Integrated; run bundles, resume on failure, side-by-side comparison |

Host instrumentation is the default: hooks and plugins capture telemetry from the agent setup engineers actually use. For reproducible and repeatable measurement, OpenEval also supports optional Docker packaging that bundles the agent runtime, skill configuration, and scenario into one image. The same image runs across many rounds to sample pass@k, cost, and outcome variance under fixed pins. Docker is optional infrastructure: use it when isolation and repeatability matter, use host instrumentation when fidelity to a real desktop or local setup matters.

## Agents

| Agent    | Status                                   |
| -------- | ---------------------------------------- |
| Cursor   | `cursor-agent` subprocess driver + hooks |
| OpenCode | Native OTEL plus plugin adapters         |
| Pi       | OTEL export via adapters                 |

### Cursor (`--agent cursor`)

OpenEval runs tasks with the headless **`cursor-agent`** CLI (not the `cursor` GUI binary). Each task round gets an isolated workspace under the run directory, seeded from the scenario `fixtures/` tree. Verifiers run in that same workspace after the agent finishes.

**Prerequisites:**

1. Install the Cursor CLI so `cursor-agent` is on your `PATH` (or set `agents.cursor.command` in config).
2. Authenticate once: `cursor-agent login` (or set `CURSOR_API_KEY`).
3. Optional: `openeval instrument --agent cursor` for hook-based OTLP export alongside runs.

```bash
cursor-agent login
cursor-agent status    # should show an authenticated account

openeval run --scenario example-fixtures --agent cursor --rounds 1
```

Override the binary path or token pricing estimates in config:

```yaml
agents:
  cursor:
    command: /usr/local/bin/cursor-agent
    cost:
      input_per_million: 3.0
      output_per_million: 15.0
```

When `cursor-agent` is missing or not authenticated, `openeval run --agent cursor` fails with an actionable error. Tests use `--agent mock` (or the default mock driver) so CI does not require a live agent.

### OpenCode (`--agent opencode`)

OpenEval runs tasks with the headless **`opencode run`** subcommand. The driver resolves the `opencode` binary from `PATH` (or `agents.opencode.command`), invokes it with `--format json --dir <workspace> <prompt>`, and parses the line-delimited event stream for the session ID and per-step token usage. Each task round gets an isolated workspace under the run directory, seeded from the scenario `fixtures/` tree. Verifiers run in that same workspace after the agent finishes.

**Prerequisites:**

1. Install the [opencode CLI](https://github.com/sst/opencode) so `opencode` is on your `PATH` (or set `agents.opencode.command` in config).
2. Authenticate your provider: `opencode auth` (or set the provider's standard env vars).
3. Optional: `opencode stats` to confirm the authenticated account and recent usage.

```bash
opencode auth
opencode run "say hi"    # should succeed and print a session id

openeval run --scenario example-fixtures --agent opencode --rounds 1
```

Override the binary path or token pricing estimates in config:

```yaml
agents:
  opencode:
    command: /usr/local/bin/opencode
    cost:
      input_per_million: 3.0
      output_per_million: 15.0
```

When `opencode` is missing or not authenticated, `openeval run --agent opencode` fails with an actionable error that names the `agents.opencode.command` config key. Tests use shell stubs (no live `opencode` install) so CI does not require a live agent.

**Hook-instrumented agents** (Cursor, and others without native export) send spans via the OpenEval config file — no OTLP env vars in the agent shell.

**Agents with built-in OpenTelemetry** (OpenCode, Pi) can also export directly when not using hooks:

```bash
export OTEL_EXPORTER_OTLP_HEADERS="..."
export OTEL_EXPORTER_OTLP_ENDPOINT="..."
export OTEL_RESOURCE_ATTRIBUTES="user.name=$(whoami)"
```

OpenEval adapters normalize that output into the shared contract so runs compare cleanly across agents.

---

**License:** [Apache-2.0](./LICENSE) · **Version:** 0.0.0-dev · **Author:** Jonathan Gabor [jgabor.se](https://jgabor.se)
