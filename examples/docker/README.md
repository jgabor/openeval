# Docker execution is deferred

OpenEval does not currently ship an agent image or implement `openeval run --image`. The CLI rejects the flag before scenario execution.

The intended opt-in interface remains:

```bash
openeval run \
  --scenario example-fixtures \
  --agent opencode \
  --image ./examples/docker/agent-eval \
  --rounds 5
```

Do not use this directory as evidence that packaged execution is available. Host execution with OpenCode, Cursor, or mock is the shipped path.
