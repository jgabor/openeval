package runner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/jgabor/openeval/internal/paths"
	"github.com/jgabor/openeval/internal/score"
)

func TestRunSkillVariationSeedsWorkspaceAndScore(t *testing.T) {
	chdirRepoRoot(t)
	runDir := filepath.Join(t.TempDir(), "run")
	result, err := Run(context.Background(), Options{
		Scenario:  "example-fixtures",
		Agent:     "mock",
		Variation: "with-demo-skill",
		Rounds:    1,
		Out:       runDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Score.Telemetry.SkillsActive) != 1 || result.Score.Telemetry.SkillsActive[0] != "demo-skill" {
		t.Fatalf("skills_active = %v", result.Score.Telemetry.SkillsActive)
	}
	seeded := filepath.Join(runDir, "tasks", "hello-verify", "round-1", ".agents", "skills", "demo-skill", "SKILL.md")
	if _, err := os.Stat(seeded); err != nil {
		t.Fatalf("missing seeded skill: %v", err)
	}
	doc, err := score.Load(paths.ScorePath(runDir))
	if err != nil {
		t.Fatal(err)
	}
	if doc.Variation != "with-demo-skill" {
		t.Fatalf("variation = %q", doc.Variation)
	}
}

func chdirRepoRoot(t *testing.T) {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			if err := os.Chdir(dir); err != nil {
				t.Fatal(err)
			}
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}
