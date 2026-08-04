"""Parse human-readable command timeouts."""


def parse_duration(value: str) -> float:
    """Return a duration in seconds."""
    if value.endswith("s"):
        return float(value[:-1])
    if value.endswith("m"):
        return float(value[:-1]) * 60
    raise ValueError(f"unsupported duration: {value}")
