package scenario_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/jgabor/openeval/internal/config"
	"github.com/jgabor/openeval/internal/scenario"
)

func TestExampleFixturesTasksFailBeforeMaintenance(t *testing.T) {
	chdirRepoRoot(t)
	sc, err := scenario.Load("example-fixtures", config.Default())
	if err != nil {
		t.Fatal(err)
	}

	wantIDs := []string{
		"parse-duration-units",
		"redact-url-credentials",
		"normalize-account-names",
		"summarize-log-levels",
	}
	if len(sc.Tasks) != len(wantIDs) {
		t.Fatalf("task count = %d, want %d", len(sc.Tasks), len(wantIDs))
	}
	for i, task := range sc.Tasks {
		if task.ID != wantIDs[i] {
			t.Errorf("task %d id = %q, want %q", i, task.ID, wantIDs[i])
		}
		t.Run(task.ID, func(t *testing.T) {
			script := filepath.Join(sc.SourceDir(), task.Verifier.Run)
			cmd := exec.Command(script)
			cmd.Dir = sc.SourceDir()
			output, err := cmd.CombinedOutput()
			if err == nil {
				t.Fatalf("untouched fixture passed verifier; output:\n%s", output)
			}
			if !bytes.Contains(output, []byte("FAILED (")) {
				t.Fatalf("verifier did not reach its failing test; output:\n%s", output)
			}
		})
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
