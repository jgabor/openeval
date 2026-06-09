package instrument

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMergeOpenEvalHooksPreservesExisting(t *testing.T) {
	existing := []byte(`{
  "version": 1,
  "hooks": {
    "sessionStart": [
      {"command": "uv run hooks/custom.py", "timeout": 10}
    ]
  }
}`)
	merged, err := mergeOpenEvalHooks(existing, "/usr/local/bin/openeval hook --agent cursor")
	if err != nil {
		t.Fatal(err)
	}
	var doc hooksFile
	if err := json.Unmarshal(merged, &doc); err != nil {
		t.Fatal(err)
	}
	starts := doc.Hooks["sessionStart"]
	if len(starts) != 2 {
		t.Fatalf("sessionStart hooks = %d, want 2", len(starts))
	}
	if starts[0].Command != "uv run hooks/custom.py" {
		t.Fatalf("first hook = %q", starts[0].Command)
	}
	if !strings.Contains(starts[1].Command, "hook --agent cursor") {
		t.Fatalf("second hook = %q", starts[1].Command)
	}
}

func TestMergeOpenEvalHooksIdempotent(t *testing.T) {
	cmd := "/bin/openeval hook --agent cursor"
	first, err := mergeOpenEvalHooks(nil, cmd)
	if err != nil {
		t.Fatal(err)
	}
	second, err := mergeOpenEvalHooks(first, cmd)
	if err != nil {
		t.Fatal(err)
	}
	var a, b hooksFile
	if err := json.Unmarshal(first, &a); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(second, &b); err != nil {
		t.Fatal(err)
	}
	for _, event := range openEvalHookEvents {
		if len(a.Hooks[event]) != len(b.Hooks[event]) {
			t.Fatalf("event %s count changed: %d -> %d", event, len(a.Hooks[event]), len(b.Hooks[event]))
		}
	}
}
