package agent

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jgabor/openeval/internal/config"
	"github.com/jgabor/openeval/internal/runcontext"
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
	wantFlags := []string{"run", "--format", "json", "--dir", workDir, "--auto", "hello world"}
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

func TestOpenCodeRunCapturesStderrOnNonZeroExit(t *testing.T) {
	dir := t.TempDir()
	stub := filepath.Join(dir, "opencode-stub.sh")
	script := `#!/usr/bin/env bash
set -euo pipefail
echo "auth: opencode session expired" >&2
exit 1
`
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

	_, _, err = driver.Run(context.Background(), Session{
		WorkDir: workDir,
		Task:    scenario.Task{ID: "t1", Prompt: "hi"},
	})
	if err == nil {
		t.Fatal("expected error when opencode exits non-zero")
	}
	msg := err.Error()
	var runErr *OpenCodeRunError
	if !errors.As(err, &runErr) {
		t.Fatalf("error %T should be an OpenCodeRunError", err)
	}
	if runErr.Kind != "execution" {
		t.Fatalf("error kind = %q, want execution", runErr.Kind)
	}
	if !strings.Contains(msg, "auth: opencode session expired") {
		t.Fatalf("error %q should include the captured stderr", msg)
	}
	for _, want := range []string{"opencode auth list", "opencode auth login"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("error %q should include remediation %q", msg, want)
		}
	}
}

func TestOpenCodeRunReportsEmptyOutput(t *testing.T) {
	dir := t.TempDir()
	stub := filepath.Join(dir, "opencode-stub.sh")
	script := `#!/usr/bin/env bash
set -euo pipefail
exit 0
`
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

	_, _, err = driver.Run(context.Background(), Session{
		WorkDir: workDir,
		Task:    scenario.Task{ID: "t1", Prompt: "hi"},
	})
	if err == nil {
		t.Fatal("expected error when opencode returns no events")
	}
	msg := err.Error()
	for _, want := range []string{"produced no JSON events", "OpenCode 1.18.11"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("error %q should include %q", msg, want)
		}
	}
}

func TestParseOpenCodeJSONRejectsMalformedLines(t *testing.T) {
	data := []byte("not-json\n" +
		`{"type":"step_finish","sessionID":"ses-valid","part":{"tokens":{}}}` + "\n")
	_, _, err := parseOpenCodeJSON(data)
	if err == nil {
		t.Fatal("expected malformed JSONL to fail")
	}
	var runErr *OpenCodeRunError
	if !errors.As(err, &runErr) || runErr.Kind != "output" {
		t.Fatalf("error = %#v, want structured output error", err)
	}
	if !strings.Contains(err.Error(), "JSONL line 1 is invalid") {
		t.Fatalf("error %q should identify the malformed line", err)
	}
}

func TestParseOpenCodeJSONReportsRuntimeErrorEvent(t *testing.T) {
	data := []byte(`{"type":"error","sessionID":"ses-error","error":{"name":"ProviderAuthError","data":{"message":"login required"}}}` + "\n")
	_, _, err := parseOpenCodeJSON(data)
	if err == nil {
		t.Fatal("expected runtime error event to fail")
	}
	var runErr *OpenCodeRunError
	if !errors.As(err, &runErr) || runErr.Kind != "runtime" {
		t.Fatalf("error = %#v, want structured runtime error", err)
	}
	for _, want := range []string{"login required", "opencode auth list", "opencode auth login"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q should include %q", err, want)
		}
	}
}

func TestOpenCodeRunSumsMultiStepTokens(t *testing.T) {
	dir := t.TempDir()
	stub := filepath.Join(dir, "opencode-stub.sh")
	script := `#!/usr/bin/env bash
set -euo pipefail
printf '{"type":"step_start","sessionID":"ses-multi","part":{"type":"step-start"}}\n'
	printf '{"type":"step_finish","sessionID":"ses-multi","part":{"type":"step-finish","tokens":{"input":1000,"output":200,"reasoning":300,"cache":{"read":400,"write":500}},"cost":999}}\n'
	printf '{"type":"step_start","sessionID":"ses-multi","part":{"type":"step-start"}}\n'
	printf '{"type":"step_finish","sessionID":"ses-multi","part":{"type":"step-finish","tokens":{"input":3000,"output":800,"reasoning":700,"cache":{"read":600,"write":700}},"cost":999}}\n'
`
	if err := os.WriteFile(stub, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := config.Default()
	cfg.Agents.OpenCode.Command = stub
	cfg.Agents.OpenCode.Cost = config.OpenCodeCostConfig{InputPerMillion: 2, OutputPerMillion: 10}
	driver, err := New("opencode", cfg)
	if err != nil {
		t.Fatal(err)
	}

	workDir := filepath.Join(dir, "workspace")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cost, traceID, err := driver.Run(context.Background(), Session{
		WorkDir: workDir,
		Task:    scenario.Task{ID: "t1", Prompt: "multi"},
		Run:     runcontext.Context{TraceID: "run-correlation-id"},
	})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if traceID != "run-correlation-id" {
		t.Fatalf("traceID = %q, want run correlation ID", traceID)
	}
	wantCost := (float64(1000+3000+400+500+600+700)/1e6)*2 + (float64(200+800+300+700)/1e6)*10
	if cost != wantCost {
		t.Fatalf("cost = %f, want %f (configured rates across all billable token classes)", cost, wantCost)
	}
}
