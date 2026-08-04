package compare

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jgabor/openeval/internal/paths"
	"github.com/jgabor/openeval/internal/score"
)

func TestRunWarnsAboutKnownModelMismatchAndStillCompares(t *testing.T) {
	dirA := writeScore(t, "a", "provider/model-a")
	dirB := writeScore(t, "b", "provider/model-b")

	out, err := Run(dirA, dirB)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"model:     provider/model-a vs provider/model-b",
		"warning:   model mismatch (provider/model-a vs provider/model-b)",
		"pass@1",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("comparison missing %q:\n%s", want, out)
		}
	}
}

func TestRunSupportsLegacyScoreWithoutModel(t *testing.T) {
	dirA := writeScore(t, "legacy", "")
	dirB := writeScore(t, "known", "provider/model")

	out, err := Run(dirA, dirB)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "model:     unknown vs provider/model") {
		t.Fatalf("comparison did not identify missing legacy model evidence:\n%s", out)
	}
	if strings.Contains(out, "model mismatch") {
		t.Fatalf("comparison treated absent legacy evidence as a known mismatch:\n%s", out)
	}
}

func writeScore(t *testing.T, variation, model string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), variation)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	doc := score.Document{
		ScenarioID: "example",
		Agent:      "opencode",
		Model:      model,
		Variation:  variation,
		Rounds:     1,
		Summary:    score.Summary{TasksTotal: 1},
	}
	if err := score.Save(paths.ScorePath(dir), doc); err != nil {
		t.Fatal(err)
	}
	return dir
}
