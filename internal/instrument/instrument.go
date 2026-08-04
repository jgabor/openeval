package instrument

import (
	"fmt"

	"github.com/jgabor/openeval/internal/config"
)

type Result struct {
	Agent    string
	Messages []string
}

func Install(agent string, cfg *config.Config) (Result, error) {
	switch agent {
	case "cursor":
		if err := installCursor(*cfg); err != nil {
			return Result{}, err
		}
		return Result{Agent: agent}, nil
	case "opencode":
		return installOpenCode(cfg)
	default:
		return Result{}, fmt.Errorf("instrument not implemented for agent %q", agent)
	}
}
