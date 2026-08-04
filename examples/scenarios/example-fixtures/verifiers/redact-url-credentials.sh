#!/usr/bin/env bash
set -euo pipefail
scenario_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
python3 -I "$scenario_dir/graders/grade.py" redact-url-credentials "$PWD/fixtures"
