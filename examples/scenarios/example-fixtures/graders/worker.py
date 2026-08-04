"""Import and call agent-owned Python code outside the grader process."""

import importlib.util
import json
import sys
from contextlib import redirect_stdout
from pathlib import Path
from types import ModuleType


def error_result(error: BaseException) -> dict[str, object]:
    return {
        "error": {
            "type": type(error).__name__,
            "message": str(error),
        }
    }


def load_module(path: Path) -> ModuleType:
    spec = importlib.util.spec_from_file_location("openeval_submission", path)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"cannot load submission: {path}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def execute(request: object) -> dict[str, object]:
    if not isinstance(request, dict) or set(request) != {"module_path", "calls"}:
        raise ValueError("invalid request")
    module_path = request["module_path"]
    calls = request["calls"]
    if not isinstance(module_path, str) or not isinstance(calls, list):
        raise ValueError("invalid request")

    module = load_module(Path(module_path))
    results: list[dict[str, object]] = []
    for call in calls:
        if not isinstance(call, dict) or set(call) != {"function", "args"}:
            raise ValueError("invalid call")
        function = call["function"]
        arguments = call["args"]
        if not isinstance(function, str) or not isinstance(arguments, list):
            raise ValueError("invalid call")
        try:
            value = getattr(module, function)(*arguments)
            json.dumps(value, allow_nan=False)
            results.append({"value": value})
        except BaseException as error:
            results.append(error_result(error))
    return {"results": results}


def main() -> None:
    protocol_output = sys.stdout
    try:
        request = json.load(sys.stdin)
        with redirect_stdout(sys.stderr):
            response = execute(request)
    except BaseException as error:
        response = error_result(error)
    json.dump(response, protocol_output, allow_nan=False, separators=(",", ":"))


if __name__ == "__main__":
    main()
