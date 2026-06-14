package agent

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/jgabor/openeval/internal/config"
)

type OpenCode struct {
	command string
	cost    config.OpenCodeCostConfig
}

func newOpenCode(cfg config.Config) (OpenCode, error) {
	cmd := strings.TrimSpace(cfg.Agents.OpenCode.Command)
	if cmd == "" {
		path, err := exec.LookPath("opencode")
		if err != nil {
			return OpenCode{}, fmt.Errorf(
				"opencode not found on PATH: install the opencode CLI and authenticate, or set agents.opencode.command in %s",
				configHint(),
			)
		}
		cmd = path
	} else if _, err := os.Stat(cmd); err != nil {
		return OpenCode{}, fmt.Errorf("agents.opencode.command %q: %w", cmd, err)
	}
	return OpenCode{command: cmd, cost: cfg.Agents.OpenCode.Cost}, nil
}

func (o OpenCode) Run(ctx context.Context, s Session) (float64, string, error) {
	_ = ctx
	_ = s
	_ = o
	return 0, "", fmt.Errorf("opencode driver is not yet implemented")
}
