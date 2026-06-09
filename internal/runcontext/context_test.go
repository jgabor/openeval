package runcontext

import (
	"strings"
	"testing"
)

func TestContextEnvEmptyWhenInactive(t *testing.T) {
	if got := (Context{}).Env(); got != nil {
		t.Fatalf("got %v, want nil", got)
	}
}

func TestContextEnvIncludesOpenEvalKeys(t *testing.T) {
	env := Context{
		ScenarioID: "demo",
		Variation:  "v1",
		TaskID:     "t1",
		Round:      3,
		TraceID:    "trace-1",
	}.Env()
	body := strings.Join(env, "\n")
	for _, want := range []string{
		"OPENEVAL_SCENARIO_ID=demo",
		"OPENEVAL_VARIATION=v1",
		"OPENEVAL_TASK_ID=t1",
		"OPENEVAL_ROUND=3",
		"OPENEVAL_TRACE_ID=trace-1",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in %q", want, body)
		}
	}
}
