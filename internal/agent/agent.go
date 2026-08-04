package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/jgabor/openeval/internal/config"
	"github.com/jgabor/openeval/internal/runcontext"
	"github.com/jgabor/openeval/internal/scenario"
)

const DefaultOpenCodeModel = "opencode/big-pickle"

type Session struct {
	WorkDir    string
	Agent      string
	Model      string
	Variation  scenario.Variation
	Task       scenario.Task
	Round      int
	PluginDirs []string
	Run        runcontext.Context
}

func ResolveModel(agentName, userModel, scenarioModel string, cfg config.Config) (string, error) {
	if agentName != "opencode" {
		if userModel != "" {
			return "", fmt.Errorf("--model is not supported for agent %q", agentName)
		}
		return "", nil
	}
	model := firstNonEmpty(userModel, scenarioModel, cfg.Agents.OpenCode.Model, DefaultOpenCodeModel)
	if err := validateModel(model); err != nil {
		return "", err
	}
	return model, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func validateModel(model string) error {
	parts := strings.Split(model, "/")
	if len(parts) < 2 || strings.ContainsFunc(model, unicode.IsSpace) {
		return fmt.Errorf("invalid model %q: expected provider/model", model)
	}
	for _, part := range parts {
		if part == "" {
			return fmt.Errorf("invalid model %q: expected provider/model", model)
		}
	}
	return nil
}

type Driver interface {
	Run(ctx context.Context, s Session) (costUSD float64, traceID string, err error)
}

func New(name string, cfg config.Config) (Driver, error) {
	switch name {
	case "", "mock":
		return Mock{}, nil
	case "cursor":
		return newCursor(cfg)
	case "opencode":
		return newOpenCode(cfg)
	default:
		return nil, fmt.Errorf("unsupported agent %q (supported: mock, cursor, opencode)", name)
	}
}

type Mock struct{}

func (Mock) Run(ctx context.Context, s Session) (float64, string, error) {
	_ = ctx
	if err := os.MkdirAll(s.WorkDir, 0o755); err != nil {
		return 0, "", err
	}
	promptPath := filepath.Join(s.WorkDir, "prompt.txt")
	if err := os.WriteFile(promptPath, []byte(s.Task.Prompt), 0o644); err != nil {
		return 0, "", err
	}
	return 0.07, fmt.Sprintf("mock-%s-r%d", s.Task.ID, s.Round), nil
}
