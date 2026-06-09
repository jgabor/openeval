package instrument

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jgabor/openeval/internal/config"
)

func TestInstallCursorMergesHooksWithoutRemovingExisting(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	cursorDir := filepath.Join(home, ".cursor")
	if err := os.MkdirAll(cursorDir, 0o755); err != nil {
		t.Fatal(err)
	}
	existing := []byte(`{
  "version": 1,
  "hooks": {
    "sessionStart": [
      {"command": "uv run hooks/custom.py", "timeout": 10}
    ]
  }
}`)
	if err := os.WriteFile(filepath.Join(cursorDir, "hooks.json"), existing, 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	if err := installCursor(cfg); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(cursorDir, "hooks.json"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(data)
	for _, want := range []string{"uv run hooks/custom.py", "hook --agent cursor", "sessionEnd", "postToolUse"} {
		if !strings.Contains(body, want) {
			t.Fatalf("merged hooks.json missing %q:\n%s", want, body)
		}
	}
}

func TestInstallCursorPermissionError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root can write anywhere")
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	cursorDir := filepath.Join(home, ".cursor")
	if err := os.MkdirAll(cursorDir, 0o444); err != nil {
		t.Fatal(err)
	}
	err := installCursor(config.Default())
	if err == nil {
		t.Fatal("expected permission error")
	}
}
