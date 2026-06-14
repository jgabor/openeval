# Changelog

## [Unreleased]

### Added

- `opencode` agent driver: subprocess invocation (`opencode run --format json --dir <workspace> <prompt>`), JSON event-stream parser that sums `step_finish` tokens for accurate multi-step cost reporting, and config-driven `input_per_million` / `output_per_million` rates that fall back to the same defaults as the cursor driver.
- Driver stub tests assert flag shape, env injection, sessionID capture, non-zero cost from parsed usage, stderr capture on non-zero exit, an actionable error for empty output, and multi-step token summation against configured rates.
