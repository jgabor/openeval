package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type SkillsConfig struct {
	Aliases map[string]string `yaml:"aliases"`
}

func (cfg Config) ResolveSkillPath(name string) (string, error) {
	if cfg.Skills.Aliases != nil {
		if raw, ok := cfg.Skills.Aliases[name]; ok {
			return resolveSkillDir(name, expandHome(raw))
		}
	}
	if path, ok := shippedSkillPath(name); ok {
		return path, nil
	}
	return "", fmt.Errorf("skill %q: not found in skills.aliases or shipped examples/skills", name)
}

func shippedSkillPath(name string) (string, bool) {
	abs, err := filepath.Abs(filepath.Join("examples", "skills", name))
	if err != nil {
		return "", false
	}
	info, err := os.Stat(abs)
	if err != nil || !info.IsDir() {
		return "", false
	}
	return abs, true
}

func resolveSkillDir(name, path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(abs); err != nil {
		return "", fmt.Errorf("skill %q path %s: %w", name, abs, err)
	}
	return abs, nil
}

func expandHome(path string) string {
	if !strings.HasPrefix(path, "~/") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[2:])
}
