# Changelog

## [Unreleased]

### Added

- `opencode` agent driver: subprocess invocation (`opencode run --format json --dir <workspace> <prompt>`), JSON event-stream parser that sums `step_finish` tokens for accurate multi-step cost reporting, and config-driven `input_per_million` / `output_per_million` rates that fall back to the same defaults as the cursor driver.
- Driver stub tests assert flag shape, env injection, sessionID capture, non-zero cost from parsed usage, stderr capture on non-zero exit, an actionable error for empty output, and multi-step token summation against configured rates.
- OpenCode harness runs auto-approve non-interactively, rejects malformed or runtime-error JSONL before verification, reports authentication remediation, keeps OpenEval run correlation authoritative, and includes reasoning tokens in configured cost estimates.
- OpenCode skill variations run `opencode debug skill` before paid execution, require the exact seeded `.agents/skills` path, and reject same-named global skills; shipped example skills now include valid portable frontmatter.
- `openeval instrument --agent opencode` explicitly enables native OTEL, derives OpenCode's base endpoint without dropping inherited exporter headers or resource attributes, appends encoded run correlation, and states the native-payload privacy boundary; `traces` distinguishes direct summary trace IDs from attribute-correlated native traces.
- `openeval doctor` defaults to OpenCode and checks config, runtime/version, `opencode auth list`, skill aliases, OTLP reachability, and native telemetry without a model call; `--agent cursor` adds secondary runtime and hook checks, while `--json` returns a versioned report with warning/fatal exit semantics and remediation.
- `openeval demo` defaults to OpenCode and one round per arm, diagnoses setup, runs retained baseline and demo-skill arms, and compares the actual returned paths; each invocation uses a new evidence root, `--dry-run` has no filesystem side effects, and unavailable OTLP remains nonfatal.
- `openeval run --agent opencode` now passes `--model opencode/big-pickle` by default; CLI, scenario, and config overrides follow explicit precedence, while malformed identifiers and unsupported Cursor or mock CLI overrides fail before execution.
- The shipped `example-fixtures` scenario now provides four independent standard-library Python maintenance tasks with focused scenario-owned graders that cannot be replaced from the agent workspace, plus a one-task credential-free mock subset.
- OpenCode scores retain the resolved provider/model; reports show known model identity, and comparisons warn about known model mismatches without blocking cross-model analysis or legacy scores.

### Documentation

- `OpenCode` subsection in `README.md` Agents section (prerequisites, install + auth + run, config override example) so `--agent opencode` is discoverable from the user-facing docs.
- `openeval run --help` `--agent` help text now lists `mock, cursor, opencode`.
- Current README, help, minimal config, and Docker example now lead with the executable OpenCode comparison path and state Cursor, mock, Pi, external scenarios, packaging, telemetry, native trace correlation, and privacy boundaries according to shipped behavior.
- Document the OpenCode model default and override precedence, non-paid catalog diagnosis, controlled comparison guidance, scenario-author selection factors, measured onboarding envelope, retained model evidence, and local-only claim limits.
