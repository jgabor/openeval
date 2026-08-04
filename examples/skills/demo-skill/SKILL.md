---
name: demo-skill
description: Use this example skill when verifying that an OpenEval variation can load workspace-local agent instructions.
---

# demo-skill

Example skill shipped with OpenEval for skill-backed scenario variations.

## Reproduce

Run the smallest relevant test or verifier before editing. Record the failing behavior and expected behavior from the task.

## Inspect

Read the failing test and the implementation it exercises. Trace inputs through the code, and check nearby cases that must keep working.

## Fix

Change the narrowest source location that explains the failure. Preserve existing behavior outside the task, and use the project's current dependencies and style.

## Recheck

Run the focused test or verifier again. Then run any broader local test suite that covers the changed code, and review the diff for unrelated edits.
