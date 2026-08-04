package runner

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jgabor/openeval/internal/config"
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
	seeded := filepath.Join(runDir, "tasks", "parse-duration-units", "round-1", ".agents", "skills", "demo-skill", "SKILL.md")
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

func TestRunDoesNotVerifyMalformedOpenCodeOutput(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

	stub := filepath.Join(root, "opencode-stub.sh")
	stubBody := `#!/bin/sh
printf '%s\n' 'not-json'
printf '%s\n' '{"type":"step_finish","sessionID":"ses-ignored","part":{"tokens":{}}}'
`
	if err := os.WriteFile(stub, []byte(stubBody), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Agents.OpenCode.Command = stub
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}

	marker := filepath.Join(root, "verifier-ran")
	verifierPath := filepath.Join(root, "verifier.sh")
	verifierBody := "#!/bin/sh\n: > " + marker + "\n"
	if err := os.WriteFile(verifierPath, []byte(verifierBody), 0o755); err != nil {
		t.Fatal(err)
	}
	scenarioPath := filepath.Join(root, "scenario.yaml")
	scenarioBody := `id: malformed-opencode
tasks:
  - id: edit
    prompt: edit the workspace
    verifier:
      type: script
      run: verifier.sh
variations:
  default: {}
`
	if err := os.WriteFile(scenarioPath, []byte(scenarioBody), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Run(context.Background(), Options{
		Scenario: scenarioPath,
		Agent:    "opencode",
		Rounds:   1,
		Out:      filepath.Join(root, "run"),
	})
	if err == nil {
		t.Fatal("expected malformed OpenCode output to stop the run")
	}
	if !strings.Contains(err.Error(), "JSONL line 1 is invalid") {
		t.Fatalf("error %q should identify malformed OpenCode output", err)
	}
	if _, statErr := os.Stat(marker); !os.IsNotExist(statErr) {
		t.Fatalf("verifier ran after agent failure: %v", statErr)
	}
}

func TestRunRetainsResolvedOpenCodeModelInScore(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

	stub := filepath.Join(root, "opencode-stub.sh")
	stubBody := `#!/bin/sh
printf '%s\n' '{"type":"step_finish","sessionID":"ses-model","part":{"tokens":{}}}'
`
	if err := os.WriteFile(stub, []byte(stubBody), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Agents.OpenCode.Command = stub
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}

	verifierPath := filepath.Join(root, "verifier.sh")
	if err := os.WriteFile(verifierPath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	scenarioPath := filepath.Join(root, "scenario.yaml")
	scenarioBody := `id: model-evidence
model: provider/selected-model
tasks:
  - id: edit
    prompt: edit the workspace
    verifier:
      type: script
      run: verifier.sh
variations:
  default: {}
`
	if err := os.WriteFile(scenarioPath, []byte(scenarioBody), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := Run(context.Background(), Options{
		Scenario: scenarioPath,
		Agent:    "opencode",
		Rounds:   1,
		Out:      filepath.Join(root, "run"),
	})
	if err != nil {
		t.Fatal(err)
	}
	doc, err := score.Load(paths.ScorePath(result.RunDir))
	if err != nil {
		t.Fatal(err)
	}
	if doc.Model != "provider/selected-model" {
		t.Fatalf("score model = %q, want resolved provider/model", doc.Model)
	}
}

func TestRunRejectsInvalidModelBeforeRunSideEffects(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))
	marker := filepath.Join(root, "agent-ran")
	stub := filepath.Join(root, "opencode-stub.sh")
	if err := os.WriteFile(stub, []byte("#!/bin/sh\n: > "+marker+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Agents.OpenCode.Command = stub
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}

	runDir := filepath.Join(root, "run")
	_, err := Run(context.Background(), Options{
		Scenario: "example-fixtures",
		Agent:    "opencode",
		Model:    "invalid",
		Rounds:   1,
		Out:      runDir,
	})
	if err == nil || !strings.Contains(err.Error(), "expected provider/model") {
		t.Fatalf("Run() error = %v, want provider/model syntax error", err)
	}
	for _, path := range []string{runDir, marker} {
		if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
			t.Fatalf("%s exists after model validation failure: %v", path, statErr)
		}
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
