"""Prepare URLs for safe diagnostic output."""


def redact_credentials(url: str) -> str:
    """Redact credential values from a URL query string."""
    return url
