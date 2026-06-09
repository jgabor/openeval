package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jgabor/openeval/internal/config"
	"github.com/jgabor/openeval/internal/runcontext"
	"github.com/jgabor/openeval/internal/scenario"
)

type Session struct {
	WorkDir    string
	Agent      string
	Variation  scenario.Variation
	Task       scenario.Task
	Round      int
	PluginDirs []string
	Run        runcontext.Context
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
	default:
		return nil, fmt.Errorf("unsupported agent %q", name)
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
