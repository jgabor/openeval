"""Summarize bracketed levels in command logs."""

import re


LEVEL = re.compile(r"^\[(DEBUG|INFO|WARNING|ERROR)\]")


def count_log_levels(lines: list[str]) -> dict[str, int]:
    """Count recognized log levels."""
    counts: dict[str, int] = {}
    for line in lines:
        match = LEVEL.match(line)
        if match:
            level = match.group(1)
            counts[level] = counts.get(level, 0) + 1
    return counts
