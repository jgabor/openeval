package score

import (
	"os"
	"path/filepath"
	"testing"
)

func TestComputePassAtK(t *testing.T) {
	rounds := []RoundResult{
		{Round: 1, Verifier: "fail"},
		{Round: 2, Verifier: "pass"},
		{Round: 3, Verifier: "fail"},
	}
	if got := ComputePassAtK(rounds, 1); got != 0 {
		t.Fatalf("pass@1 = %v", got)
	}
	if got := ComputePassAtK(rounds, 3); got != 1 {
		t.Fatalf("pass@3 = %v", got)
	}
}

func TestLoadLegacyScoreWithoutModel(t *testing.T) {
	path := filepath.Join(t.TempDir(), "score.json")
	legacy := `{
  "schema": "openeval.score.v1",
  "scenario_id": "legacy",
  "agent": "cursor",
  "rounds": 1,
  "tasks": 0,
  "summary": {},
  "by_task": [],
  "telemetry": {}
}`
	if err := os.WriteFile(path, []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}

	doc, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Schema != Schema || doc.Agent != "cursor" || doc.Model != "" {
		t.Fatalf("legacy score = %+v, want cursor score with unknown model", doc)
	}
}
