package workspace

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jgabor/openeval/internal/config"
)

// SeedSkills copies registered skills into workDir/.agents/skills/<name>/ for workspace-local discovery.
func SeedSkills(workDir string, skillNames []string, cfg config.Config) error {
	if len(skillNames) == 0 {
		return nil
	}
	for _, name := range skillNames {
		src, _, err := cfg.ResolveSkill(name)
		if err != nil {
			return err
		}
		dest := filepath.Join(workDir, ".agents", "skills", name)
		if err := copyFSTree(src, dest); err != nil {
			return fmt.Errorf("seed skill %q: %w", name, err)
		}
	}
	return nil
}

// SkillPluginDirs returns plugin directories for skills that ship a .claude-plugin manifest.
func SkillPluginDirs(skillNames []string, cfg config.Config) ([]string, error) {
	var dirs []string
	for _, name := range skillNames {
		_, src, err := cfg.ResolveSkill(name)
		if err != nil {
			return nil, err
		}
		if src == "" {
			continue
		}
		pluginDir := filepath.Join(src, ".claude-plugin")
		if info, err := os.Stat(pluginDir); err == nil && info.IsDir() {
			dirs = append(dirs, src)
		}
	}
	return dirs, nil
}
