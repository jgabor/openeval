"""Grade one example-fixtures task against an agent workspace."""

import json
import subprocess
import sys
from pathlib import Path
from typing import Callable

Call = tuple[str, list[object]]
Grader = tuple[str, list[Call], Callable[[list[object]], None]]


def worker_error(value: object) -> str:
    if not isinstance(value, dict) or set(value) != {"type", "message"}:
        raise AssertionError("submission worker returned an invalid error")
    error_type = value["type"]
    message = value["message"]
    if not isinstance(error_type, str) or not isinstance(message, str):
        raise AssertionError("submission worker returned an invalid error")
    return f"{error_type}: {message}"


def call_submission(module_path: Path, calls: list[Call]) -> list[object]:
    request = {
        "module_path": str(module_path),
        "calls": [
            {"function": function, "args": arguments}
            for function, arguments in calls
        ],
    }
    try:
        completed = subprocess.run(
            [sys.executable, "-I", str(Path(__file__).with_name("worker.py"))],
            input=json.dumps(request),
            capture_output=True,
            check=False,
            text=True,
            timeout=10,
        )
    except subprocess.TimeoutExpired as error:
        raise AssertionError("submission worker timed out") from error
    if completed.returncode != 0:
        raise AssertionError("submission worker exited without a valid response")
    try:
        response = json.loads(completed.stdout)
    except (json.JSONDecodeError, UnicodeError) as error:
        raise AssertionError("submission worker returned invalid JSON") from error
    if not isinstance(response, dict):
        raise AssertionError("submission worker returned an invalid response")
    if set(response) == {"error"}:
        raise AssertionError(f"submission worker failed: {worker_error(response['error'])}")
    if set(response) != {"results"} or not isinstance(response["results"], list):
        raise AssertionError("submission worker returned an invalid response")
    if len(response["results"]) != len(calls):
        raise AssertionError("submission worker returned the wrong result count")

    values: list[object] = []
    for outcome in response["results"]:
        if not isinstance(outcome, dict):
            raise AssertionError("submission worker returned an invalid result")
        if set(outcome) == {"error"}:
            raise AssertionError(f"submission call failed: {worker_error(outcome['error'])}")
        if set(outcome) != {"value"}:
            raise AssertionError("submission worker returned an invalid result")
        values.append(outcome["value"])
    return values


def grade_durations(results: list[object]) -> None:
    assert results == [0.25, 2.0, 180.0]


def grade_urls(results: list[object]) -> None:
    assert results == [
        "https://example.test/jobs?token=REDACTED&limit=20&api_key=REDACTED",
        "https://example.test/?password=REDACTED#status",
    ]


def grade_accounts(results: list[object]) -> None:
    assert results == ["release-engineering", "quality-assurance"]


def grade_logs(results: list[object]) -> None:
    assert results == [{"INFO": 2, "ERROR": 1}]


GRADERS: dict[str, Grader] = {
    "parse-duration-units": (
        "maintainer_tools/durations.py",
        [
            ("parse_duration", ["250ms"]),
            ("parse_duration", ["2s"]),
            ("parse_duration", ["3m"]),
        ],
        grade_durations,
    ),
    "redact-url-credentials": (
        "maintainer_tools/urls.py",
        [
            (
                "redact_credentials",
                ["https://example.test/jobs?token=secret&limit=20&api_key=key"],
            ),
            (
                "redact_credentials",
                ["https://example.test/?password=hunter2#status"],
            ),
        ],
        grade_urls,
    ),
    "normalize-account-names": (
        "maintainer_tools/accounts.py",
        [
            ("normalize_account_name", ["  Release   Engineering  "]),
            ("normalize_account_name", ["Quality\tAssurance"]),
        ],
        grade_accounts,
    ),
    "summarize-log-levels": (
        "maintainer_tools/logs.py",
        [
            (
                "count_log_levels",
                [["[info] started", "[ERROR] failed", "plain output", "[Info] done"]],
            )
        ],
        grade_logs,
    ),
}


def main() -> None:
    if len(sys.argv) != 3 or sys.argv[1] not in GRADERS:
        raise SystemExit(f"usage: {Path(sys.argv[0]).name} TASK_ID FIXTURES_DIR")
    relative_path, calls, grade = GRADERS[sys.argv[1]]
    module_path = Path(sys.argv[2]).resolve() / relative_path
    grade(call_submission(module_path, calls))


if __name__ == "__main__":
    main()
