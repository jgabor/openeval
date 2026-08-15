package config

import (
	"io/fs"
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

func TestResolveSkillUsesShippedExampleOutsideCheckout(t *testing.T) {
	t.Chdir(t.TempDir())
	cfg := Default()
	source, path, err := cfg.ResolveSkill("demo-skill")
	if err != nil {
		t.Fatal(err)
	}
	if path != "" {
		t.Fatalf("shipped skill path = %q, want embedded source", path)
	}
	data, err := fs.ReadFile(source, "SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "name: demo-skill") {
		t.Fatalf("embedded SKILL.md missing demo-skill frontmatter")
	}
}

func TestResolveSkillAliasPrecedesShippedFallback(t *testing.T) {
	root := t.TempDir()
	alias := filepath.Join(root, "alias")
	if err := os.Mkdir(alias, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(alias, "SKILL.md"), []byte("alias bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := Default()
	cfg.Skills.Aliases["demo-skill"] = alias

	source, path, err := cfg.ResolveSkill("demo-skill")
	if err != nil {
		t.Fatal(err)
	}
	data, err := fs.ReadFile(source, "SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	if path != alias || string(data) != "alias bytes" {
		t.Fatalf("resolved path=%q data=%q, want alias %q", path, data, alias)
	}
}

func TestResolveSkillInvalidAliasDoesNotUseShippedFallback(t *testing.T) {
	cfg := Default()
	cfg.Skills.Aliases["demo-skill"] = filepath.Join(t.TempDir(), "missing")
	_, _, err := cfg.ResolveSkill("demo-skill")
	if err == nil || !strings.Contains(err.Error(), `skill "demo-skill" path`) {
		t.Fatalf("ResolveSkill error = %v, want invalid alias path", err)
	}
}
