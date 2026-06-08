package config

import (
	"os"
	"path/filepath"
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
