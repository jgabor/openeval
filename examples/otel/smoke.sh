#!/usr/bin/env bash
set -euo pipefail

ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
COMPOSE_FILE="$ROOT/examples/otel/compose.yaml"
EVIDENCE_DIR=${OPENEVAL_OTEL_SMOKE_DIR:-"$ROOT/scenarios/example-fixtures/otel-smoke-$(date -u +%Y%m%dT%H%M%SZ)"}
RUN_DIR="$EVIDENCE_DIR/run"
CONFIG_HOME="$EVIDENCE_DIR/config"

mkdir -p "$CONFIG_HOME/openeval"
cat >"$CONFIG_HOME/openeval/config.yaml" <<'YAML'
version: 1
telemetry:
  protocol: http/json
  endpoint: http://localhost:4318/v1/traces
  insecure: true
YAML

cleanup() {
  status=$?
  trap - EXIT
  docker compose -f "$COMPOSE_FILE" logs --no-color >"$EVIDENCE_DIR/jaeger.log" 2>&1 || true
  docker compose -f "$COMPOSE_FILE" down || true
  printf 'smoke evidence: %s\n' "$EVIDENCE_DIR"
  exit "$status"
}
trap cleanup EXIT

cd "$ROOT"
GOFLAGS="${GOFLAGS:+$GOFLAGS }-buildvcs=false" mage build
docker compose -f "$COMPOSE_FILE" up -d --wait --wait-timeout 60

XDG_CONFIG_HOME="$CONFIG_HOME" ./bin/openeval run \
  --scenario ./examples/scenarios/example-fixtures/edit-file-only.yaml \
  --agent mock --rounds 1 --out "$RUN_DIR"

SCORE="$RUN_DIR/score.json"
read -r TASK_ID SOURCE_TRACE_ID NORMALIZED_TRACE_ID < <(python3 - "$SCORE" <<'PY'
import hashlib
import json
import re
import sys

with open(sys.argv[1], encoding="utf-8") as score_file:
    score = json.load(score_file)
task = score["by_task"][0]
source = task["rounds"][0]["trace_id"].strip()
normalized = source.lower() if re.fullmatch(r"[0-9a-fA-F]{32}", source) else hashlib.sha256(source.encode()).hexdigest()[:32]
print(task["task_id"], source, normalized)
PY
)

python3 - "$NORMALIZED_TRACE_ID" "$EVIDENCE_DIR/jaeger-trace.json" <<'PY'
import json
import sys
import time
import urllib.request

trace_id, evidence_path = sys.argv[1:]
url = "http://localhost:16686/api/traces/" + trace_id
deadline = time.monotonic() + 30
last_error = "trace not returned"
while time.monotonic() < deadline:
    try:
        with urllib.request.urlopen(url, timeout=2) as response:
            body = response.read()
        with open(evidence_path, "wb") as evidence:
            evidence.write(body)
        payload = json.loads(body)
        for trace in payload.get("data", []):
            processes = trace.get("processes", {})
            for span in trace.get("spans", []):
                process = processes.get(span.get("processID"), {})
                if span.get("operationName") == "openeval.session" and process.get("serviceName") == "openeval-agent":
                    sys.exit(0)
        last_error = "exact trace lacked openeval.session from openeval-agent"
    except (OSError, ValueError) as error:
        last_error = str(error)
    time.sleep(1)
raise SystemExit(f"Jaeger proof failed for {trace_id}: {last_error}")
PY

XDG_CONFIG_HOME="$CONFIG_HOME" ./bin/openeval traces "$RUN_DIR" \
  --task "$TASK_ID" --round 1 | tee "$EVIDENCE_DIR/traces.txt"

python3 - "$EVIDENCE_DIR/traces.txt" "$NORMALIZED_TRACE_ID" <<'PY'
import sys

with open(sys.argv[1], encoding="utf-8") as output_file:
    output = output_file.read().splitlines()
want = "jaeger_ui: http://localhost:16686/trace/" + sys.argv[2]
if want not in output:
    raise SystemExit(f"traces output missing {want!r}")
PY

printf 'source trace ID: %s\nnormalized trace ID: %s\n' "$SOURCE_TRACE_ID" "$NORMALIZED_TRACE_ID"
