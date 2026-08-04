package instrument

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jgabor/openeval/internal/config"
)

func installCursor(cfg config.Config) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	cursorDir := filepath.Join(home, ".cursor")
	if err := os.MkdirAll(cursorDir, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", cursorDir, err)
	}
	hookCmd, err := openEvalHookCommand()
	if err != nil {
		return err
	}
	hooksPath := filepath.Join(cursorDir, "hooks.json")
	existing, err := os.ReadFile(hooksPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read %s: %w", hooksPath, err)
	}
	merged, err := mergeOpenEvalHooks(existing, hookCmd)
	if err != nil {
		return err
	}
	if err := os.WriteFile(hooksPath, merged, 0o644); err != nil {
		return fmt.Errorf("write %s (check permissions): %w", hooksPath, err)
	}
	return writeLegacyHookConfig(cursorDir, cfg)
}

func openEvalHookCommand() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve openeval binary: %w", err)
	}
	exe, err = filepath.Abs(exe)
	if err != nil {
		return "", err
	}
	if strings.Contains(filepath.Base(exe), "go-build") {
		if path, err := exec.LookPath("openeval"); err == nil {
			exe = path
		}
	}
	return fmt.Sprintf("%q hook --agent cursor", exe), nil
}

func writeLegacyHookConfig(cursorDir string, cfg config.Config) error {
	dir := filepath.Join(cursorDir, "openeval")
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
