package scenario_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jgabor/openeval/internal/config"
	"github.com/jgabor/openeval/internal/scenario"
)

func TestScenarioSelectionsPrecedeShippedFallback(t *testing.T) {
	root := t.TempDir()
	explicit := writeScenario(t, root, "explicit.yaml", "explicit")
	alias := writeScenario(t, root, "alias.yaml", "alias")
	cfg := config.Default()
	cfg.Scenarios.Aliases["example-fixtures"] = alias

	tests := []struct {
		name      string
		selection string
		wantID    string
	}{
		{name: "explicit path", selection: explicit, wantID: "explicit"},
		{name: "alias", selection: "example-fixtures", wantID: "alias"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc, err := scenario.Load(tt.selection, cfg)
			if err != nil {
				t.Fatal(err)
			}
			if sc.ID != tt.wantID {
				t.Fatalf("scenario id = %q, want %q", sc.ID, tt.wantID)
			}
		})
	}
}

func TestInvalidExplicitScenarioDoesNotUseShippedFallback(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "example-fixtures.yaml")
	_, err := scenario.Load(missing, config.Default())
	if err == nil || !strings.Contains(err.Error(), missing) {
		t.Fatalf("Load error = %v, want missing explicit path", err)
	}
}

func writeScenario(t *testing.T, dir, name, id string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	body := "id: " + id + "\ntasks:\n  - id: task\n    prompt: test\n    verifier: {type: script, run: verify.sh}\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
