package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jgabor/openeval/internal/config"
	"github.com/jgabor/openeval/internal/scenario"
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

func TestOpenCodeRunInvokesStubAndParsesUsage(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "invocation.log")
	envPath := filepath.Join(dir, "env.log")
	stub := filepath.Join(dir, "opencode-stub.sh")
	script := fmt.Sprintf(`#!/usr/bin/env bash
set -euo pipefail
printf '%%s\n' "$@" > %q
printf '%%s\n' "${OPENCODE_VARIATION-}" > %q
printf '{"type":"step_start","sessionID":"ses-test-1","part":{"type":"step-start"}}\n'
printf '{"type":"step_finish","sessionID":"ses-test-1","part":{"type":"step-finish","tokens":{"input":1000,"output":500,"cache":{"read":0,"write":0}},"cost":0}}\n'
`, logPath, envPath)
	if err := os.WriteFile(stub, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := config.Default()
	cfg.Agents.OpenCode.Command = stub
	driver, err := New("opencode", cfg)
	if err != nil {
		t.Fatal(err)
	}

	workDir := filepath.Join(dir, "workspace")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatal(err)
	}

	sess := Session{
		WorkDir: workDir,
		Task:    scenario.Task{ID: "t1", Prompt: "hello world"},
		Variation: scenario.Variation{
			Env: map[string]string{"OPENCODE_VARIATION": "yes"},
		},
	}

	cost, traceID, err := driver.Run(context.Background(), sess)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if traceID != "ses-test-1" {
		t.Fatalf("traceID = %q, want %q", traceID, "ses-test-1")
	}
	if cost <= 0 {
		t.Fatalf("cost = %f, want positive value reflecting parsed usage", cost)
	}

	argsRaw, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	argv := strings.Split(strings.TrimSpace(string(argsRaw)), "\n")
	wantFlags := []string{"run", "--format", "json", "--dir", workDir, "hello world"}
	if len(argv) != len(wantFlags) {
		t.Fatalf("argv = %v, want %v", argv, wantFlags)
	}
	for i, want := range wantFlags {
		if argv[i] != want {
			t.Fatalf("argv[%d] = %q, want %q", i, argv[i], want)
		}
	}

	envRaw, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(envRaw)); got != "yes" {
		t.Fatalf("OPENCODE_VARIATION = %q, want %q", got, "yes")
	}
}
