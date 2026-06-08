# OpenEval

Drop-in, platform-agnostic instrumentation and evaluation for CLI coding agents.

- Instrument: install hooks or adapters that export logs, metrics, and traces to an OTLP backend
- Run: execute scenario tasks against the agent under test
- Evaluate: score sessions, aggregate pass@k, compare cost and quality across variations
- Results: publish run output linked to session traces (skills, tools, context, cost)

Run the same scenarios on Cursor, OpenCode, Pi, and other supported agents. Compare pass rates and cost without rebuilding your setup for each one. Better numbers upstream mean cheaper, more reliable agents for the teams and end users downstream.

> **Status:** `0.0.0-dev` — CLI and hooks described here are the target interface. Example scenarios ship in this repo; the binary is not published yet.

### Terminology

| Term          | Meaning                                                                                            |
| ------------- | -------------------------------------------------------------------------------------------------- |
| **Eval**      | The overall activity: instrument an agent, run a scenario, score and compare results               |
| **Scenario**  | A pinned task set and verifiers (what problems run). Selected with `--scenario`                    |
| **Variation** | A labeled config arm under test (baseline, a version, a feature flag). Selected with `--variation` |
| **Run**       | One execution of a scenario, stored under `./scenarios/<scenario>/runs/…`                          |

Shipped examples live under `examples/`. Your scenarios live under `evals/` (or any path you pass to `--scenario`).

## Quick start

### Prerequisites

- Linux, macOS, or Windows
- Docker (optional, for packaged eval runs)
- An OTLP backend (Jaeger, Grafana Cloud, or any OTLP collector)
- A supported CLI agent installed locally (Cursor, OpenCode, or Pi)

### Install

```bash
go install openeval@latest
```

From source:

```bash
git clone https://github.com/jgabor/openeval.git
cd openeval
mage install
```

Instrument detected agents (writes hooks or plugins and default config):

```bash
openeval instrument --all

# Or a single agent
openeval instrument --agent cursor
```

`openeval instrument` writes agent-specific hooks or plugins and creates the config file if missing (see [Configuration](#configuration)).

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
scenarios:
  aliases:
    frontend-tasks: ./examples/scenarios/frontend-tasks/scenario.yaml
```

Telemetry settings follow standard OTLP environment variable names. Hook-instrumented agents read this file, so you do not need export variables in the agent process.

Export options (set in config or via OTLP env vars):

- Protocols: gRPC, HTTP/Protobuf, HTTP/JSON
- Generation-based batching for HTTP/JSON (spans grouped by `generation_id`, sent on stop event)
- Automated install scripts for Windows and macOS/Linux

### Scenarios and variations

- **`--scenario`** selects which tasks to run: a built-in id (`example-fixtures`, `frontend-tasks`), a path to your YAML (`./evals/my-scenario.yaml`), a registered alias, or an integrated external scenario (`deepswe`, `margin-eval`).
- **`--variation`** selects a named config arm defined in the scenario file (or `default` when omitted). Compare two runs that share the same scenario.
- **`--rounds`** sets attempts per task (default: `3`). Use multiple rounds for pass@k; treat a single round as exploratory only.
- **`--out`** is optional. Default layout:
  - no `--variation` (or `default`) → `./scenarios/<scenario>/runs/<n>` with auto-incrementing `n` per scenario; prior runs are kept
  - named `--variation` → `./scenarios/<scenario>/runs/<variation>`; re-running overwrites that directory unless you pass `--out`

`<scenario>` is the resolved scenario id (`example-fixtures`, `frontend-tasks`, `deepswe`, …), not the agent name. Agent, rounds, and timestamps live in `score.json` and traces.

**How variations apply:** each scenario YAML lists `variations` with the agent config for that arm — skills to enable, env vars, prompt pins, and similar. `openeval run --variation removed-design-tokens` loads that block before starting tasks. The variation label is stored in `score.json`; skill names and telemetry reflect what actually ran.

**Compare rules:** `openeval compare` requires the same `scenario_id` in both `score.json` files. It warns when `agent` differs. Pass two run directories; labels come from each run's `variation` field.

Drop instrumentation that never changes scenario scores or cost (see [Supported telemetry](#supported-telemetry) — capture only what informs comparisons).

### Scenario format

Minimal shape (see `examples/scenarios/example-fixtures/scenario.yaml` and `examples/scenarios/frontend-tasks/scenario.yaml`):

```yaml
id: frontend-tasks
tasks:
  - id: token-audit
    prompt: Audit design-token usage in src/components/Card.tsx
    verifier: { type: script, run: ./verifiers/token-audit.sh }
  - id: component-styles
    prompt: Fix spacing in Card to match the design system
    verifier: { type: script, run: ./verifiers/component-styles.sh }
variations:
  baseline:
    skills: [frontend-design]
  removed-design-tokens:
    skills: [frontend-design]
    env:
      DESIGN_TOKENS_ENABLED: "false"
```

### Run

```bash
# Example scenario shipped with the repo → ./scenarios/example-fixtures/runs/1
openeval run \
  --scenario example-fixtures \
  --agent cursor \
  --rounds 3

# Example scenario with variations → ./scenarios/frontend-tasks/runs/1
openeval run \
  --scenario frontend-tasks \
  --agent cursor

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
# Example scenario: design tokens on vs off
openeval run --scenario frontend-tasks --agent cursor \
  --variation baseline --rounds 3
# → ./scenarios/frontend-tasks/runs/baseline

openeval run --scenario frontend-tasks --agent cursor \
  --variation removed-design-tokens --rounds 3
# → ./scenarios/frontend-tasks/runs/removed-design-tokens

openeval compare \
  ./scenarios/frontend-tasks/runs/baseline \
  ./scenarios/frontend-tasks/runs/removed-design-tokens

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

From `./scenarios/frontend-tasks/runs/removed-design-tokens/score.json`:

```json
{
  "schema": "openeval.score.v1",
  "scenario_id": "frontend-tasks",
  "agent": "cursor",
  "variation": "removed-design-tokens",
  "rounds": 3,
  "tasks": 2,
  "summary": {
    "pass_at_1": 0.5,
    "pass_at_3": 0.5,
    "tasks_passed": 1,
    "tasks_total": 2,
    "cost_usd_total": 0.42,
    "cost_usd_per_passed_task": 0.42,
    "tokens_input_total": 18420,
    "tokens_output_total": 3920
  },
  "by_task": [
    {
      "task_id": "token-audit",
      "verifier": "pass",
      "rounds": [
        {
          "round": 1,
          "verifier": "pass",
          "cost_usd": 0.08,
          "trace_id": "a1b2c3..."
        },
        {
          "round": 2,
          "verifier": "pass",
          "cost_usd": 0.07,
          "trace_id": "d4e5f6..."
        },
        {
          "round": 3,
          "verifier": "fail",
          "cost_usd": 0.09,
          "trace_id": "789abc..."
        }
      ],
      "pass_at_k": { "1": 1.0, "3": 1.0 }
    },
    {
      "task_id": "component-styles",
      "verifier": "fail",
      "rounds": [
        { "round": 1, "verifier": "fail", "cost_usd": 0.06, "trace_id": "..." },
        { "round": 2, "verifier": "fail", "cost_usd": 0.06, "trace_id": "..." },
        { "round": 3, "verifier": "fail", "cost_usd": 0.06, "trace_id": "..." }
      ],
      "pass_at_k": { "1": 0.0, "3": 0.0 }
    }
  ],
  "telemetry": {
    "skills_active": ["frontend-design"],
    "cost_by_skill_usd": { "frontend-design": 0.42 },
    "sessions": 6
  }
}
```

### `openeval report` output (example)

```text
scenario: frontend-tasks
agent: cursor
variation: removed-design-tokens
rounds: 3

pass@1: 0.50  (1/2 tasks passed on first attempt)
pass@3: 0.50  (1/2 tasks passed within 3 attempts)
cost:   $0.42 total  ($0.42 per passed task)

tasks:
  token-audit       PASS  pass@3=1.00  cost=$0.24
  component-styles  FAIL  pass@3=0.00  cost=$0.18

traces: 6 sessions, OTLP service openeval-agent
        open Jaeger with trace_id from score.json or run:
        openeval traces ./scenarios/frontend-tasks/runs/removed-design-tokens --task token-audit --round 2
```

### `openeval compare` output (example)

```text
comparing: baseline to removed-design-tokens
scenario:  frontend-tasks
agent:     cursor

                    baseline    removed-design-tokens    delta
pass@1                 0.50                  0.50      0.00
pass@3                 0.50                  0.50      0.00
cost_usd_total         0.38                  0.42     +0.04
cost_per_passed        0.38                  0.42     +0.04

results:
  summary: pass@3 unchanged, cost up slightly
  logs:
    ./scenarios/frontend-tasks/runs/baseline/score.json
    ./scenarios/frontend-tasks/runs/removed-design-tokens/score.json
```

## Supported telemetry

OpenEval captures session metadata across supported agents. For resources, fields use the same names where applicable:

- Inventory: enabled resources in the session (listed per domain below)
- Metadata: descriptors (names, descriptions, hooks)
- Content: full captured text or payload
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

All captured data can be aggregated or broken down: average cost per session with a given skill on vs off, tool success and failure rates, total cost when a skill is used, and similar comparisons across any dimension above.

## Privacy

- Data masking before export (prompts and secrets)
- Rules configurable per deployment

## Scenarios

A scenario pins tasks, verifiers, and runtime versions so runs stay comparable. **`--variation`** is separate: it tags which agent config produced a given run. Compare only runs that share the same `--scenario`.

| Scenario                                           | `--scenario` value                               | Source                                                               |
| -------------------------------------------------- | ------------------------------------------------ | -------------------------------------------------------------------- |
| Example fixtures                                   | `example-fixtures`                               | `examples/scenarios/example-fixtures/`                               |
| Frontend tasks (example)                           | `frontend-tasks`                                 | `examples/scenarios/frontend-tasks/`; demonstrates variation compare |
| Custom                                             | `./evals/my-scenario.yaml` or a registered alias | Your `evals/` directory; not shipped by OpenEval                     |
| [DeepSWE](https://deepswe.datacurve.ai/)           | `deepswe`                                        | Integrated; long-horizon SWE tasks, behavior-based verification      |
| [Margin Eval](https://github.com/Margin-Lab/evals) | `margin-eval`                                    | Integrated; run bundles, resume on failure, side-by-side comparison  |

Host instrumentation is the default: hooks and plugins capture telemetry from the agent setup engineers actually use. For reproducible and repeatable measurement, OpenEval also supports optional Docker packaging that bundles the agent runtime, skill configuration, and scenario into one image. The same image runs across many rounds to sample pass@k, cost, and outcome variance under fixed pins. Docker is optional infrastructure: use it when isolation and repeatability matter, use host instrumentation when fidelity to a real desktop or local setup matters.

## Agents

| Agent    | Status                           |
| -------- | -------------------------------- |
| Cursor   | Hook-based OTEL export           |
| OpenCode | Native OTEL plus plugin adapters |
| Pi       | OTEL export via adapters         |

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
