# Docker packaging

Optional images bundle an agent runtime, skill configuration, and scenario into one reproducible package.

Target interface:

```bash
openeval run \
  --scenario example-fixtures \
  --image ./examples/docker/agent-eval \
  --rounds 5
```

The `agent-eval` image definition will live here when implemented.