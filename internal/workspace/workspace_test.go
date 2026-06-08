package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jgabor/openeval/internal/config"
	"github.com/jgabor/openeval/internal/scenario"
)

func TestSeedCopiesFixturesIntoWorkspace(t *testing.T) {
	root := t.TempDir()
	fixtures := filepath.Join(root, "fixtures")
	if err := os.MkdirAll(filepath.Join(fixtures, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixtures, "hello.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixtures, "src", "main.py"), []byte("print('hi')"), 0o644); err != nil {
		t.Fatal(err)
	}
	scenarioPath := filepath.Join(root, "scenario.yaml")
	if err := os.WriteFile(scenarioPath, []byte("id: demo\ntasks:\n  - id: t\n    prompt: p\n    verifier: {type: script, run: ./v.sh}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sc, err := scenario.Load(scenarioPath, config.Default())
	if err != nil {
		t.Fatal(err)
	}

	dest := filepath.Join(root, "run", "tasks", "t", "round-1")
	if err := Seed(sc, dest); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{"fixtures/hello.txt", "fixtures/src/main.py"} {
		if _, err := os.Stat(filepath.Join(dest, rel)); err != nil {
			t.Fatalf("missing %s: %v", rel, err)
		}
	}
}

func TestRoundWorkspacesAreIsolated(t *testing.T) {
	runDir := t.TempDir()
	dir1 := RoundDir(runDir, "edit-file", 1)
	dir2 := RoundDir(runDir, "edit-file", 2)
	if dir1 == dir2 {
		t.Fatalf("round dirs should differ: %q", dir1)
	}
	if err := os.MkdirAll(dir1, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir1, "marker"), []byte("r1"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir2, "marker")); err == nil {
		t.Fatal("round 2 should not see round 1 marker")
	}
}
