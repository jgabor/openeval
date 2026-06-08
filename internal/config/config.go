package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

const Version = "0.0.0-dev"

type Config struct {
	Version   int             `yaml:"version"`
	Telemetry TelemetryConfig `yaml:"telemetry"`
	Privacy   PrivacyConfig   `yaml:"privacy"`
	Scenarios ScenarioAliases `yaml:"scenarios"`
	Skills    SkillsConfig    `yaml:"skills"`
	Agents    AgentsConfig    `yaml:"agents"`
}

type AgentsConfig struct {
	Cursor CursorAgentConfig `yaml:"cursor"`
}

type CursorAgentConfig struct {
	Command string           `yaml:"command"`
	Cost    CursorCostConfig `yaml:"cost"`
}

type CursorCostConfig struct {
	InputPerMillion  float64 `yaml:"input_per_million"`
	OutputPerMillion float64 `yaml:"output_per_million"`
}

type TelemetryConfig struct {
	Protocol string `yaml:"protocol"`
	Endpoint string `yaml:"endpoint"`
	Insecure bool   `yaml:"insecure"`
}

type PrivacyConfig struct {
	MaskPrompts bool `yaml:"mask_prompts"`
	MaskSecrets bool `yaml:"mask_secrets"`
}

type ScenarioAliases struct {
	Aliases map[string]string `yaml:"aliases"`
}

func Default() Config {
	return Config{
		Version: 1,
		Telemetry: TelemetryConfig{
			Protocol: "http/json",
			Endpoint: "http://localhost:4318/v1/traces",
			Insecure: true,
		},
		Privacy: PrivacyConfig{
			MaskPrompts: true,
			MaskSecrets: true,
		},
		Scenarios: ScenarioAliases{Aliases: map[string]string{}},
		Skills:    SkillsConfig{Aliases: map[string]string{}},
	}
}

func Dir() (string, error) {
	if runtime.GOOS == "windows" {
		local := os.Getenv("LOCALAPPDATA")
		if local == "" {
			return "", fmt.Errorf("LOCALAPPDATA is not set")
		}
		return filepath.Join(local, "openeval"), nil
	}
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "openeval"), nil
}

func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

func Load() (Config, error) {
	path, err := Path()
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return Config{}, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config %s: %w", path, err)
	}
	if cfg.Scenarios.Aliases == nil {
		cfg.Scenarios.Aliases = map[string]string{}
	}
	if cfg.Skills.Aliases == nil {
		cfg.Skills.Aliases = map[string]string{}
	}
	return cfg, nil
}

func Save(cfg Config) error {
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func ResolveScenarioAlias(cfg Config, name string) (string, bool) {
	path, ok := cfg.Scenarios.Aliases[name]
	return path, ok
}
