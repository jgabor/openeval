package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jgabor/openeval/internal/config"
)

func TestSeedSkillsCopiesIntoWorkspaceAgentsDir(t *testing.T) {
	root := t.TempDir()
	skillSrc := filepath.Join(root, "skills", "demo-skill")
	if err := os.MkdirAll(skillSrc, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillSrc, "SKILL.md"), []byte("# demo"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := config.Default()
	cfg.Skills.Aliases = map[string]string{"demo-skill": skillSrc}

	workDir := filepath.Join(root, "workspace")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := SeedSkills(workDir, []string{"demo-skill"}, cfg); err != nil {
		t.Fatal(err)
	}
	got := filepath.Join(workDir, ".agents", "skills", "demo-skill", "SKILL.md")
	if _, err := os.Stat(got); err != nil {
		t.Fatalf("missing seeded skill: %v", err)
	}
}

func TestSkillPluginDirsFindsClaudePlugin(t *testing.T) {
	root := t.TempDir()
	skillSrc := filepath.Join(root, "plugin-skill")
	pluginDir := filepath.Join(skillSrc, ".claude-plugin")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "plugin.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := config.Default()
	cfg.Skills.Aliases = map[string]string{"plugin-skill": skillSrc}

	dirs, err := SkillPluginDirs([]string{"plugin-skill"}, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(dirs) != 1 || dirs[0] != skillSrc {
		t.Fatalf("dirs = %v, want [%s]", dirs, skillSrc)
	}
}

func TestShippedSkillsHaveOpenCodeFrontmatter(t *testing.T) {
	root := repositoryRoot(t)
	for _, name := range []string{"demo-skill", "plugin-skill"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(root, "examples", "skills", name, "SKILL.md")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			body := string(data)
			for _, want := range []string{"---\nname: " + name + "\n", "description:"} {
				if !strings.Contains(body, want) {
					t.Fatalf("%s missing %q", path, want)
				}
			}
		})
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}
