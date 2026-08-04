# Changelog

## [Unreleased]

### Added

- `opencode` agent driver: subprocess invocation (`opencode run --format json --dir <workspace> <prompt>`), JSON event-stream parser that sums `step_finish` tokens for accurate multi-step cost reporting, and config-driven `input_per_million` / `output_per_million` rates that fall back to the same defaults as the cursor driver.
- Driver stub tests assert flag shape, env injection, sessionID capture, non-zero cost from parsed usage, stderr capture on non-zero exit, an actionable error for empty output, and multi-step token summation against configured rates.
- OpenCode harness runs auto-approve non-interactively, rejects malformed or runtime-error JSONL before verification, reports authentication remediation, keeps OpenEval run correlation authoritative, and includes reasoning tokens in configured cost estimates.
- OpenCode skill variations run `opencode debug skill` before paid execution, require the exact seeded `.agents/skills` path, and reject same-named global skills; shipped example skills now include valid portable frontmatter.

### Documentation

- `OpenCode` subsection in `README.md` Agents section (prerequisites, install + auth + run, config override example) so `--agent opencode` is discoverable from the user-facing docs.
- `openeval run --help` `--agent` help text now lists `mock, cursor, opencode`.
