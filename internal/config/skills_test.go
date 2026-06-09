package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveSkillPathExpandsHome(t *testing.T) {
	root := t.TempDir()
	skillDir := filepath.Join(root, "my-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}

	home := filepath.Join(root, "home")
	if err := os.MkdirAll(filepath.Join(home, "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(home, "skills", "linked")
	if err := os.Symlink(skillDir, link); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", home)
	cfg := Default()
	cfg.Skills.Aliases = map[string]string{"linked": "~/skills/linked"}

	got, err := cfg.ResolveSkillPath("linked")
	if err != nil {
		t.Fatal(err)
	}
	want, err := filepath.Abs(link)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("ResolveSkillPath = %q, want %q", got, want)
	}
}

func TestResolveSkillPathUsesShippedExample(t *testing.T) {
	chdirRepoRoot(t)
	cfg := Default()
	got, err := cfg.ResolveSkillPath("demo-skill")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(got, filepath.Join("examples", "skills", "demo-skill")) {
		t.Fatalf("ResolveSkillPath = %q", got)
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
