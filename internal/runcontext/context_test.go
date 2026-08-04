package runcontext

import (
	"slices"
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

func TestMergeEnvironmentUsesLastValueAndSortsKeys(t *testing.T) {
	got := MergeEnvironment(
		[]string{"SECOND=inherited", "FIRST=inherited"},
		[]string{"THIRD=variation", "SECOND=variation"},
	)
	want := []string{"FIRST=inherited", "SECOND=variation", "THIRD=variation"}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestFromMapReturnsMergeableEnvironment(t *testing.T) {
	got := MergeEnvironment(FromMap(map[string]string{
		"SECOND": "two",
		"FIRST":  "one",
	}))
	want := []string{"FIRST=one", "SECOND=two"}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestEnvironmentKeepsReservedRunContextAuthoritative(t *testing.T) {
	got := Environment(
		[]string{"SHARED=inherited", "VARIATION_ONLY=inherited", "OPENEVAL_TASK_ID=inherited"},
		map[string]string{
			"SHARED":           "variation",
			"VARIATION_ONLY":   "variation",
			"OPENEVAL_TASK_ID": "variation",
		},
		Context{TaskID: "reserved", TraceID: "trace-1"},
	)
	want := []string{
		"OPENEVAL_ROUND=0",
		"OPENEVAL_SCENARIO_ID=",
		"OPENEVAL_TASK_ID=reserved",
		"OPENEVAL_TRACE_ID=trace-1",
		"OPENEVAL_VARIATION=",
		"SHARED=variation",
		"VARIATION_ONLY=variation",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
