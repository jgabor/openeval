package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jgabor/openeval/internal/config"
)

func TestNewOpenCodeMissingBinary(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	_, err := New("opencode", config.Default())
	if err == nil {
		t.Fatal("expected error when opencode is missing")
	}
	if !strings.Contains(err.Error(), "opencode not found") {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(err.Error(), "agents.opencode.command") {
		t.Fatalf("error should mention config hint, got: %v", err)
	}
}

func TestNewOpenCodeResolvesFromPath(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "opencode")
	if err := os.WriteFile(binary, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)

	d, err := New("opencode", config.Default())
	if err != nil {
		t.Fatal(err)
	}
	oc, ok := d.(OpenCode)
	if !ok {
		t.Fatalf("expected OpenCode driver, got %T", d)
	}
	if oc.command != binary {
		t.Fatalf("command = %q, want %q", oc.command, binary)
	}
}

func TestNewOpenCodeRespectsCommandOverride(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "my-opencode")
	if err := os.WriteFile(binary, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := config.Default()
	cfg.Agents.OpenCode.Command = binary

	d, err := New("opencode", cfg)
	if err != nil {
		t.Fatal(err)
	}
	oc := d.(OpenCode)
	if oc.command != binary {
		t.Fatalf("command = %q, want %q", oc.command, binary)
	}
}

func TestNewOpenCodeRejectsMissingCommandOverride(t *testing.T) {
	cfg := config.Default()
	cfg.Agents.OpenCode.Command = "/nonexistent/path/to/opencode"

	_, err := New("opencode", cfg)
	if err == nil {
		t.Fatal("expected error when command override is missing")
	}
	if !strings.Contains(err.Error(), "agents.opencode.command") {
		t.Fatalf("error should mention agents.opencode.command, got: %v", err)
	}
}

func TestNewUnknownAgentListsOpencode(t *testing.T) {
	_, err := New("foo", config.Default())
	if err == nil {
		t.Fatal("expected error for unknown agent")
	}
	msg := err.Error()
	for _, want := range []string{"mock", "cursor", "opencode"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("error %q should list %q as a supported agent", msg, want)
		}
	}
}
