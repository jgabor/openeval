package demo

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jgabor/openeval/internal/agent"
	"github.com/jgabor/openeval/internal/config"
	"github.com/jgabor/openeval/internal/doctor"
	"github.com/jgabor/openeval/internal/paths"
	"github.com/jgabor/openeval/internal/score"
)

func TestDryRunPrintsExactWorkWithoutFilesystemMutation(t *testing.T) {
	chdirRepositoryRoot(t)
	base := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(base, "config"))
	before, err := os.ReadDir(base)
	if err != nil {
		t.Fatal(err)
	}
	result, err := Run(context.Background(), Options{
		Agent:  "opencode",
		Out:    base,
		DryRun: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.DryRun {
		t.Fatal("result was not marked dry-run")
	}
	after, err := os.ReadDir(base)
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != len(before) {
		t.Fatalf("dry-run mutated %s: before=%d after=%d", base, len(before), len(after))
	}
	body := result.Plan.Format()
	for _, want := range []string{
		"openeval doctor --agent opencode",
		"--model " + agent.DefaultOpenCodeModel,
		"--variation baseline",
		"--variation with-demo-skill",
		"--rounds 1",
		"openeval compare " + result.Plan.BaselineDir + " " + result.Plan.SkillDir,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("dry-run output missing %q:\n%s", want, body)
		}
	}
	if got := strings.Count(body, "--model "+agent.DefaultOpenCodeModel); got != 3 {
		t.Fatalf("dry-run model count = %d, want doctor and both arms:\n%s", got, body)
	}
}

func TestOpenCodeOverrideAppliesToBothArmsAndRetainedScores(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))
	logPath := filepath.Join(root, "opencode.log")
	stub := filepath.Join(root, "opencode-stub.sh")
	stubBody := `#!/bin/sh
printf '%s\n' "$*" >> "` + logPath + `"
case "$*" in
  --version) printf '%s\n' '1.18.11' ;;
  'auth list') printf '%s\n' 'Credentials: override-provider' ;;
  models) printf '%s\n' 'override-provider/override-model' ;;
  run*) printf '%s\n' '{"type":"step_finish","sessionID":"ses-demo","part":{"tokens":{}}}' ;;
  *) exit 99 ;;
esac
`
	if err := os.WriteFile(stub, []byte(stubBody), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Agents.OpenCode.Command = stub
	cfg.Telemetry.Endpoint = ""
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}
	verifier := filepath.Join(root, "verify.sh")
	if err := os.WriteFile(verifier, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	scenarioPath := filepath.Join(root, "scenario.yaml")
	scenarioBody := `id: demo-model
model: scenario/provider-model
tasks:
  - id: edit
    prompt: edit the workspace
    verifier:
      type: script
      run: verify.sh
variations:
  baseline: {}
  with-demo-skill: {}
`
	if err := os.WriteFile(scenarioPath, []byte(scenarioBody), 0o644); err != nil {
		t.Fatal(err)
	}

	const model = "override-provider/override-model"
	result, err := Run(context.Background(), Options{
		Scenario: scenarioPath,
		Agent:    "opencode",
		Model:    model,
		Rounds:   1,
		Out:      filepath.Join(root, "evidence"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Plan.Model != model || result.Baseline.Score.Model != model || result.Skill.Score.Model != model {
		t.Fatalf("resolved models: plan=%q baseline=%q skill=%q, want %q", result.Plan.Model, result.Baseline.Score.Model, result.Skill.Score.Model, model)
	}
	for _, runDir := range []string{result.Baseline.RunDir, result.Skill.RunDir} {
		doc, err := score.Load(paths.ScorePath(runDir))
		if err != nil {
			t.Fatal(err)
		}
		if doc.Model != model {
			t.Fatalf("retained score %s model = %q, want %q", runDir, doc.Model, model)
		}
	}
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(string(data), "--model "+model); got != 2 {
		t.Fatalf("agent model invocation count = %d, want both arms:\n%s", got, data)
	}
}

func TestRunMockRetainsUniqueEvidenceAndComparesReturnedPaths(t *testing.T) {
	chdirRepositoryRoot(t)
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))
	cfg := config.Default()
	cfg.Telemetry.Endpoint = "http://127.0.0.1:1/v1/traces"
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}
	base := filepath.Join(root, "demos")
	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(base, "existing-evidence")
	if err := os.WriteFile(marker, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}

	first, err := Run(context.Background(), Options{Agent: "mock", Rounds: 1, Out: base})
	if err != nil {
		t.Fatal(err)
	}
	second, err := Run(context.Background(), Options{Agent: "mock", Rounds: 1, Out: base})
	if err != nil {
		t.Fatal(err)
	}
	if first.Plan.Root == second.Plan.Root {
		t.Fatalf("rerun reused evidence root %s", first.Plan.Root)
	}
	if data, err := os.ReadFile(marker); err != nil || string(data) != "keep" {
		t.Fatalf("existing evidence changed: data=%q err=%v", data, err)
	}
	for _, result := range []Result{first, second} {
		if result.Baseline.RunDir != result.Plan.BaselineDir || result.Skill.RunDir != result.Plan.SkillDir {
			t.Fatalf("runner paths differ from retained plan: %+v", result)
		}
		for _, want := range []string{
			"comparing: baseline to with-demo-skill",
			"cost_usd_total",
			filepath.Join(result.Baseline.RunDir, "score.json"),
			filepath.Join(result.Skill.RunDir, "score.json"),
		} {
			if !strings.Contains(result.Comparison, want) {
				t.Fatalf("comparison missing %q:\n%s", want, result.Comparison)
			}
		}
	}
}

func TestDiagnosisErrorIncludesAuthenticationRemediation(t *testing.T) {
	err := diagnosisError(doctor.Report{
		Checks: []doctor.Check{{
			ID:          "authentication",
			Status:      doctor.StatusFail,
			Summary:     "no credentials",
			Remediation: "run `opencode auth login`, then `opencode auth list`",
		}},
	})
	for _, want := range []string{"doctor failed before demo execution", "opencode auth login", "opencode auth list"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q missing %q", err, want)
		}
	}
}

func chdirRepositoryRoot(t *testing.T) {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			t.Chdir(dir)
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}
