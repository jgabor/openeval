package report

import (
	"strings"
	"testing"

	"github.com/jgabor/openeval/internal/score"
)

func TestFormatShowsKnownModelAndSupportsLegacyScore(t *testing.T) {
	legacy := Format(score.Document{ScenarioID: "example", Agent: "cursor", Rounds: 1})
	if strings.Contains(legacy, "model:") {
		t.Fatalf("legacy report unexpectedly contains model identity:\n%s", legacy)
	}

	known := Format(score.Document{ScenarioID: "example", Agent: "opencode", Model: "provider/model", Rounds: 1})
	if !strings.Contains(known, "model: provider/model\n") {
		t.Fatalf("report missing model identity:\n%s", known)
	}
	for _, claim := range []string{"deterministic", "frozen", "reproducible"} {
		if strings.Contains(strings.ToLower(known), claim) {
			t.Fatalf("report makes unsupported %q claim:\n%s", claim, known)
		}
	}
}
