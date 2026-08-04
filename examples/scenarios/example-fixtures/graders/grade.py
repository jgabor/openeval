"""Grade one example-fixtures task against an agent workspace."""

import importlib.util
import sys
from pathlib import Path
from types import ModuleType
from typing import Callable


def load_module(fixtures: Path, relative_path: str) -> ModuleType:
    path = fixtures / relative_path
    spec = importlib.util.spec_from_file_location("openeval_submission", path)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"cannot load submission: {path}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def grade_durations(module: ModuleType) -> None:
    assert module.parse_duration("250ms") == 0.25
    assert module.parse_duration("2s") == 2.0
    assert module.parse_duration("3m") == 180.0


def grade_urls(module: ModuleType) -> None:
    assert module.redact_credentials(
        "https://example.test/jobs?token=secret&limit=20&api_key=key"
    ) == "https://example.test/jobs?token=REDACTED&limit=20&api_key=REDACTED"
    assert (
        module.redact_credentials("https://example.test/?password=hunter2#status")
        == "https://example.test/?password=REDACTED#status"
    )


def grade_accounts(module: ModuleType) -> None:
    assert (
        module.normalize_account_name("  Release   Engineering  ")
        == "release-engineering"
    )
    assert module.normalize_account_name("Quality\tAssurance") == "quality-assurance"


def grade_logs(module: ModuleType) -> None:
    lines = ["[info] started", "[ERROR] failed", "plain output", "[Info] done"]
    assert module.count_log_levels(lines) == {"INFO": 2, "ERROR": 1}


GRADERS: dict[str, tuple[str, Callable[[ModuleType], None]]] = {
    "parse-duration-units": ("maintainer_tools/durations.py", grade_durations),
    "redact-url-credentials": ("maintainer_tools/urls.py", grade_urls),
    "normalize-account-names": ("maintainer_tools/accounts.py", grade_accounts),
    "summarize-log-levels": ("maintainer_tools/logs.py", grade_logs),
}


def main() -> None:
    if len(sys.argv) != 3 or sys.argv[1] not in GRADERS:
        raise SystemExit(f"usage: {Path(sys.argv[0]).name} TASK_ID FIXTURES_DIR")
    relative_path, grade = GRADERS[sys.argv[1]]
    grade(load_module(Path(sys.argv[2]).resolve(), relative_path))


if __name__ == "__main__":
    main()
