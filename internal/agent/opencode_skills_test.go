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

func TestOpenCodeRunValidatesWorkspaceSkillsBeforePaidRun(t *testing.T) {
	root := t.TempDir()
	workDir := filepath.Join(root, "workspace")
	skillPath := filepath.Join(workDir, ".agents", "skills", "demo-skill", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(skillPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(skillPath, []byte("---\nname: demo-skill\ndescription: demo\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	logPath := filepath.Join(root, "invocations.log")
	stub := filepath.Join(root, "opencode-stub.sh")
	script := fmt.Sprintf(`#!/bin/sh
set -eu
printf '%%s|%%s\n' "$PWD" "$*" >> %q
if [ "$1" = "debug" ]; then
  if [ "$PWD" = %q ]; then
    printf '%%s\n' '[{"name":"demo-skill","location":%q}]'
  else
    printf '%%s\n' '[]'
  fi
  exit 0
fi
printf '%%s\n' '{"type":"step_finish","sessionID":"ses-skill","part":{"tokens":{}}}'
`, logPath, workDir, skillPath)
	if err := os.WriteFile(stub, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := config.Default()
	cfg.Agents.OpenCode.Command = stub
	driver, err := New("opencode", cfg)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = driver.Run(context.Background(), Session{
		WorkDir: workDir,
		Task:    scenario.Task{ID: "edit", Prompt: "edit"},
		Variation: scenario.Variation{
			Skills: []string{"demo-skill"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Fatalf("invocations = %v, want global debug, workspace debug, then paid run", lines)
	}
	if !strings.Contains(lines[0], "|debug skill") || lines[1] != workDir+"|debug skill" {
		t.Fatalf("skill checks = %v, want isolated then exact workspace discovery", lines[:2])
	}
	if !strings.Contains(lines[2], "|run --format json --dir "+workDir+" --auto --model "+DefaultOpenCodeModel+" edit") {
		t.Fatalf("paid run invocation = %q", lines[2])
	}
}

func TestOpenCodeRunRejectsGlobalSkillCollisionBeforePaidRun(t *testing.T) {
	root := t.TempDir()
	logPath := filepath.Join(root, "invocations.log")
	stub := filepath.Join(root, "opencode-stub.sh")
	script := fmt.Sprintf(`#!/bin/sh
set -eu
printf '%%s\n' "$*" >> %q
if [ "$1" = "debug" ]; then
  printf '%%s\n' '[{"name":"demo-skill","location":"/global/.agents/skills/demo-skill/SKILL.md"}]'
  exit 0
fi
exit 99
`, logPath)
	if err := os.WriteFile(stub, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := config.Default()
	cfg.Agents.OpenCode.Command = stub
	driver, err := New("opencode", cfg)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = driver.Run(context.Background(), Session{
		WorkDir: filepath.Join(root, "workspace"),
		Task:    scenario.Task{ID: "edit", Prompt: "edit"},
		Variation: scenario.Variation{
			Skills: []string{"demo-skill"},
		},
	})
	if err == nil {
		t.Fatal("expected global skill collision")
	}
	for _, want := range []string{"skill error", "collides with global discovery", "/global/.agents/skills/demo-skill/SKILL.md"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q missing %q", err, want)
		}
	}
	data, readErr := os.ReadFile(logPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if got := strings.TrimSpace(string(data)); got != "debug skill" {
		t.Fatalf("invocations = %q, paid run must not start after a collision", got)
	}
}
