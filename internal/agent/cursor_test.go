package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jgabor/openeval/internal/config"
	"github.com/jgabor/openeval/internal/runcontext"
	"github.com/jgabor/openeval/internal/scenario"
)

func TestNewCursorMissingBinary(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	_, err := New("cursor", config.Default())
	if err == nil {
		t.Fatal("expected error when cursor-agent is missing")
	}
	if !strings.Contains(err.Error(), "cursor-agent not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCursorResolvesCursorAgentNotGUI(t *testing.T) {
	dir := t.TempDir()
	agentPath := filepath.Join(dir, "cursor-agent")
	if err := os.WriteFile(agentPath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	guiPath := filepath.Join(dir, "cursor")
	if err := os.WriteFile(guiPath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)

	d, err := New("cursor", config.Default())
	if err != nil {
		t.Fatal(err)
	}
	c, ok := d.(Cursor)
	if !ok {
		t.Fatalf("expected Cursor driver, got %T", d)
	}
	if c.command != agentPath {
		t.Fatalf("command = %q, want %q", c.command, agentPath)
	}
}

func TestCursorRunInvokesStubWithFlagsAndEnv(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "invocation.log")
	envPath := filepath.Join(dir, "env.log")
	stub := filepath.Join(dir, "cursor-agent-stub.sh")
	script := fmt.Sprintf(`#!/usr/bin/env bash
set -euo pipefail
printf '%%s\n' "$@" > %q
printf '%%s\n' "${DESIGN_TOKENS_ENABLED-}" > %q
printf '%%s\n' '{"type":"result","subtype":"success","is_error":false,"session_id":"sess-test-1","request_id":"req-test-1","usage":{"inputTokens":1000000,"outputTokens":1000000,"cacheReadTokens":0,"cacheWriteTokens":0}}'
`, logPath, envPath)
	if err := os.WriteFile(stub, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := config.Default()
	cfg.Agents.Cursor.Command = stub
	driver, err := New("cursor", cfg)
	if err != nil {
		t.Fatal(err)
	}

	workDir := filepath.Join(dir, "workspace")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cost, traceID, err := driver.Run(context.Background(), Session{
		WorkDir: workDir,
		Task: scenario.Task{
			ID:     "edit-file",
			Prompt: "add a comment",
		},
		Variation: scenario.Variation{
			Env: map[string]string{"DESIGN_TOKENS_ENABLED": "false"},
		},
		Round: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if traceID != "sess-test-1" {
		t.Fatalf("traceID = %q", traceID)
	}
	wantCost := 3.0 + 15.0
	if cost != wantCost {
		t.Fatalf("cost = %v, want %v", cost, wantCost)
	}

	got, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	args := strings.TrimSpace(string(got))
	for _, want := range []string{"-p", "--trust", "--workspace", workDir, "--output-format", "json", "add a comment"} {
		if !strings.Contains(args, want) {
			t.Fatalf("args %q missing %q", args, want)
		}
	}

	envGot, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(envGot)) != "false" {
		t.Fatalf("DESIGN_TOKENS_ENABLED = %q, want false", strings.TrimSpace(string(envGot)))
	}
}

func TestParseCursorJSONCostFromUsage(t *testing.T) {
	payload := `{"type":"result","subtype":"success","is_error":false,"session_id":"abc","usage":{"inputTokens":2000000,"outputTokens":500000,"cacheReadTokens":1000000,"cacheWriteTokens":0}}`
	traceID, usage, err := parseCursorJSON([]byte(payload))
	if err != nil {
		t.Fatal(err)
	}
	if traceID != "abc" {
		t.Fatalf("traceID = %q", traceID)
	}
	cost := estimateCost(usage, config.CursorCostConfig{InputPerMillion: 2, OutputPerMillion: 10})
	want := 11.0
	if cost != want {
		t.Fatalf("cost = %v, want %v", cost, want)
	}
}

func TestParseCursorJSONUsesRequestIDFallback(t *testing.T) {
	payload := `{"type":"result","subtype":"success","request_id":"req-only","usage":{"inputTokens":0,"outputTokens":0}}`
	traceID, _, err := parseCursorJSON([]byte(payload))
	if err != nil {
		t.Fatal(err)
	}
	if traceID != "req-only" {
		t.Fatalf("traceID = %q", traceID)
	}
}

func TestParseCursorJSONRejectsErrorSubtype(t *testing.T) {
	payload := `{"type":"result","subtype":"error","is_error":true}`
	_, _, err := parseCursorJSON([]byte(payload))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCursorJSONLastLine(t *testing.T) {
	payload := "noise\n{\"type\":\"result\",\"subtype\":\"success\",\"session_id\":\"last\",\"usage\":{}}\n"
	traceID, _, err := parseCursorJSON([]byte(payload))
	if err != nil {
		t.Fatal(err)
	}
	if traceID != "last" {
		t.Fatalf("traceID = %q", traceID)
	}
}

func TestCursorRunPassesPluginDirFlags(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "invocation.log")
	stub := filepath.Join(dir, "cursor-agent-stub.sh")
	script := fmt.Sprintf(`#!/usr/bin/env bash
set -euo pipefail
printf '%%s\n' "$@" > %q
printf '%%s\n' '{"type":"result","subtype":"success","session_id":"sess-plugin","usage":{}}'
`, logPath)
	if err := os.WriteFile(stub, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := config.Default()
	cfg.Agents.Cursor.Command = stub
	driver, err := New("cursor", cfg)
	if err != nil {
		t.Fatal(err)
	}

	workDir := filepath.Join(dir, "workspace")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatal(err)
	}
	pluginDir := filepath.Join(dir, "plugins", "example")

	_, _, err = driver.Run(context.Background(), Session{
		WorkDir:    workDir,
		PluginDirs: []string{pluginDir},
		Task:       scenario.Task{Prompt: "build"},
		Round:      1,
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	args := strings.TrimSpace(string(got))
	for _, want := range []string{"--plugin-dir", pluginDir} {
		if !strings.Contains(args, want) {
			t.Fatalf("args %q missing %q", args, want)
		}
	}
}

func TestParseCursorJSONRoundTrip(t *testing.T) {
	var resp cursorResponse
	resp.Type = "result"
	resp.Subtype = "success"
	resp.SessionID = "sess"
	resp.Usage.InputTokens = 1
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	traceID, usage, err := parseCursorJSON(data)
	if err != nil {
		t.Fatal(err)
	}
	if traceID != "sess" || usage.InputTokens != 1 {
		t.Fatalf("got trace=%q usage=%+v", traceID, usage)
	}
}

func TestCursorRunPropagatesOpenEvalContextEnv(t *testing.T) {
	t.Setenv("DESIGN_TOKENS_ENABLED", "inherited")
	dir := t.TempDir()
	envPath := filepath.Join(dir, "env.log")
	stub := filepath.Join(dir, "cursor-agent-stub.sh")
	script := "#!/usr/bin/env bash\nset -euo pipefail\nenv | sort > " + envPath + "\nprintf '%s\n' '{\"type\":\"result\",\"subtype\":\"success\",\"session_id\":\"sess-test-1\",\"usage\":{}}'\n"
	if err := os.WriteFile(stub, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := config.Default()
	cfg.Agents.Cursor.Command = stub
	driver, err := New("cursor", cfg)
	if err != nil {
		t.Fatal(err)
	}

	workDir := filepath.Join(dir, "workspace")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatal(err)
	}

	correlationID := "abc123correlationtraceid00000001"
	_cost, traceID, err := driver.Run(context.Background(), Session{
		WorkDir: workDir,
		Task:    scenario.Task{ID: "edit-file", Prompt: "go"},
		Variation: scenario.Variation{
			Env: map[string]string{
				"DESIGN_TOKENS_ENABLED": "false",
				"OPENEVAL_TASK_ID":      "variation-must-not-win",
			},
		},
		Round: 2,
		Run: runcontext.Context{
			ScenarioID: "example-fixtures",
			Variation:  "default",
			TaskID:     "edit-file",
			Round:      2,
			TraceID:    correlationID,
		},
	})
	_ = _cost
	if err != nil {
		t.Fatal(err)
	}
	if traceID != correlationID {
		t.Fatalf("traceID = %q, want correlation id", traceID)
	}

	envGot, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}
	body := string(envGot)
	for _, want := range []string{
		"OPENEVAL_SCENARIO_ID=example-fixtures",
		"OPENEVAL_VARIATION=default",
		"OPENEVAL_TASK_ID=edit-file",
		"OPENEVAL_ROUND=2",
		"OPENEVAL_TRACE_ID=" + correlationID,
		"DESIGN_TOKENS_ENABLED=false",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("env missing %q:\n%s", want, body)
		}
	}
	if strings.Contains(body, "OPENEVAL_TASK_ID=variation-must-not-win") {
		t.Fatalf("variation replaced reserved task identity:\n%s", body)
	}
}
