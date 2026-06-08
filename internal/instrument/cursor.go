package instrument

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jgabor/openeval/internal/config"
)

func Install(agent string, cfg config.Config) error {
	switch agent {
	case "cursor":
		return installCursor(cfg)
	default:
		return fmt.Errorf("instrument not implemented for agent %q", agent)
	}
}

func installCursor(cfg config.Config) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".cursor", "openeval")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	hook := filepath.Join(dir, "otel-hook.json")
	content := fmt.Sprintf(`{
  "telemetry_endpoint": %q,
  "mask_prompts": %t,
  "mask_secrets": %t
}
`, cfg.Telemetry.Endpoint, cfg.Privacy.MaskPrompts, cfg.Privacy.MaskSecrets)
	return os.WriteFile(hook, []byte(content), 0o644)
}
