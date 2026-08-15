package scenario

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/jgabor/openeval/examples"
	"github.com/jgabor/openeval/internal/config"
	"gopkg.in/yaml.v3"
)

type Scenario struct {
	ID         string               `yaml:"id"`
	Model      string               `yaml:"model"`
	Tasks      []Task               `yaml:"tasks"`
	Variations map[string]Variation `yaml:"variations"`
	sourcePath string
	sourceFS   fs.FS
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
	path, idHint, source, err := resolvePath(nameOrPath, cfg)
	if err != nil {
		return Scenario{}, err
	}
	var data []byte
	if path != "" {
		data, err = os.ReadFile(path)
	} else {
		data, err = fs.ReadFile(source, "scenario.yaml")
	}
	if err != nil {
		return Scenario{}, fmt.Errorf("read scenario %q: %w", nameOrPath, err)
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
	sc.sourceFS = source
	if err := sc.validate(); err != nil {
		return Scenario{}, err
	}
	return sc, nil
}

func resolvePath(nameOrPath string, cfg config.Config) (path string, idHint string, source fs.FS, err error) {
	if strings.Contains(nameOrPath, string(os.PathSeparator)) || strings.HasSuffix(nameOrPath, ".yaml") || strings.HasSuffix(nameOrPath, ".yml") {
		abs, err := filepath.Abs(nameOrPath)
		return abs, "", nil, err
	}
	if alias, ok := config.ResolveScenarioAlias(cfg, nameOrPath); ok {
		abs, err := filepath.Abs(alias)
		return abs, nameOrPath, nil, err
	}
	if nameOrPath == "example-fixtures" {
		source, err := fs.Sub(examples.Files, "scenarios/example-fixtures")
		return "", nameOrPath, source, err
	}
	return "", "", nil, fmt.Errorf("unknown scenario %q (not a path, alias, or shipped scenario)", nameOrPath)
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
	if s.sourcePath == "" {
		return ""
	}
	return filepath.Dir(s.sourcePath)
}

func (s Scenario) SourceFS() fs.FS {
	if s.sourceFS != nil {
		return s.sourceFS
	}
	return os.DirFS(s.SourceDir())
}

// MaterializeSource returns a directory for tools that must execute scenario files.
func (s Scenario) MaterializeSource() (string, func(), error) {
	if s.sourceFS == nil {
		return s.SourceDir(), func() {}, nil
	}
	dir, err := os.MkdirTemp("", "openeval-scenario-")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { _ = os.RemoveAll(dir) }
	if err := copyFS(s.sourceFS, dir); err != nil {
		cleanup()
		return "", nil, err
	}
	return dir, cleanup, nil
}

func copyFS(source fs.FS, dest string) error {
	return fs.WalkDir(source, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := filepath.Join(dest, filepath.FromSlash(path))
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := fs.ReadFile(source, path)
		if err != nil {
			return err
		}
		mode := fs.FileMode(0o644)
		if strings.HasSuffix(path, ".sh") {
			mode = 0o755
		}
		return os.WriteFile(target, data, mode)
	})
}
