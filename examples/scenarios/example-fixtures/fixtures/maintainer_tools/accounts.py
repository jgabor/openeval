"""Normalize account names used in generated file paths."""


def normalize_account_name(name: str) -> str:
    """Return the stable path form of an account display name."""
    return name.lower().replace(" ", "-")
