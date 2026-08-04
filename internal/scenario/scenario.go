package scenario

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jgabor/openeval/internal/config"
	"gopkg.in/yaml.v3"
)

type Scenario struct {
	ID         string               `yaml:"id"`
	Model      string               `yaml:"model"`
	Tasks      []Task               `yaml:"tasks"`
	Variations map[string]Variation `yaml:"variations"`
	sourcePath string
}

type Task struct {
	ID       string   `yaml:"id"`
	Prompt   string   `yaml:"prompt"`
	Verifier Verifier `yaml:"verifier"`
}

type Verifier struct {
	Type string `yaml:"type"`
	Run  string `yaml:"run"`
}

type Variation struct {
	Skills []string          `yaml:"skills"`
	Env    map[string]string `yaml:"env"`
}

func Load(nameOrPath string, cfg config.Config) (Scenario, error) {
	path, idHint, err := resolvePath(nameOrPath, cfg)
	if err != nil {
		return Scenario{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Scenario{}, fmt.Errorf("read scenario %s: %w", path, err)
	}
	var sc Scenario
	if err := yaml.Unmarshal(data, &sc); err != nil {
		return Scenario{}, fmt.Errorf("parse scenario %s: %w", path, err)
	}
	if sc.ID == "" {
		if idHint != "" {
			sc.ID = idHint
		} else {
			sc.ID = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		}
	}
	if sc.Variations == nil {
		sc.Variations = map[string]Variation{"default": {}}
	}
	sc.sourcePath = path
	if err := sc.validate(); err != nil {
		return Scenario{}, err
	}
	return sc, nil
}

func resolvePath(nameOrPath string, cfg config.Config) (path string, idHint string, err error) {
	if strings.Contains(nameOrPath, string(os.PathSeparator)) || strings.HasSuffix(nameOrPath, ".yaml") || strings.HasSuffix(nameOrPath, ".yml") {
		abs, err := filepath.Abs(nameOrPath)
		return abs, "", err
	}
	if alias, ok := config.ResolveScenarioAlias(cfg, nameOrPath); ok {
		abs, err := filepath.Abs(alias)
		return abs, nameOrPath, err
	}
	builtins := []string{
		filepath.Join("examples", "scenarios", nameOrPath, "scenario.yaml"),
		filepath.Join("examples", "scenarios", nameOrPath+".yaml"),
	}
	for _, candidate := range builtins {
		if _, err := os.Stat(candidate); err == nil {
			abs, err := filepath.Abs(candidate)
			return abs, nameOrPath, err
		}
	}
	return "", "", fmt.Errorf("unknown scenario %q (not a path, alias, or built-in)", nameOrPath)
}

func (s Scenario) validate() error {
	if len(s.Tasks) == 0 {
		return fmt.Errorf("scenario %s has no tasks", s.ID)
	}
	for _, t := range s.Tasks {
		if t.ID == "" {
			return fmt.Errorf("scenario %s has a task without id", s.ID)
		}
		if t.Verifier.Type != "script" || t.Verifier.Run == "" {
			return fmt.Errorf("task %s requires verifier.type=script with run path", t.ID)
		}
	}
	return nil
}

func (s Scenario) Variation(name string) (Variation, error) {
	if name == "" {
		name = "default"
	}
	v, ok := s.Variations[name]
	if !ok {
		return Variation{}, fmt.Errorf("variation %q not defined in scenario %s", name, s.ID)
	}
	return v, nil
}

func (s Scenario) SourceDir() string {
	return filepath.Dir(s.sourcePath)
}
