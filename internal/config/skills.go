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
	if cfg.Skills.Aliases == nil {
		return "", fmt.Errorf("skill %q: no skills.aliases in config", name)
	}
	raw, ok := cfg.Skills.Aliases[name]
	if !ok {
		return "", fmt.Errorf("skill %q: not found in skills.aliases", name)
	}
	path := expandHome(raw)
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
