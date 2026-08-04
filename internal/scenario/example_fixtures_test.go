package scenario_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jgabor/openeval/internal/config"
	"github.com/jgabor/openeval/internal/scenario"
	"github.com/jgabor/openeval/internal/verifier"
	"github.com/jgabor/openeval/internal/workspace"
)

func TestExampleFixturesTasksFailBeforeMaintenance(t *testing.T) {
	chdirRepoRoot(t)
	sc, err := scenario.Load("example-fixtures", config.Default())
	if err != nil {
		t.Fatal(err)
	}

	wantIDs := []string{
		"parse-duration-units",
		"redact-url-credentials",
		"normalize-account-names",
		"summarize-log-levels",
	}
	if len(sc.Tasks) != len(wantIDs) {
		t.Fatalf("task count = %d, want %d", len(sc.Tasks), len(wantIDs))
	}
	for i, task := range sc.Tasks {
		if task.ID != wantIDs[i] {
			t.Errorf("task %d id = %q, want %q", i, task.ID, wantIDs[i])
		}
		t.Run(task.ID, func(t *testing.T) {
			script := filepath.Join(sc.SourceDir(), task.Verifier.Run)
			cmd := exec.Command(script)
			cmd.Dir = sc.SourceDir()
			output, err := cmd.CombinedOutput()
			if err == nil {
				t.Fatalf("untouched fixture passed verifier; output:\n%s", output)
			}
			if !bytes.Contains(output, []byte("FAILED (")) {
				t.Fatalf("verifier did not reach its failing test; output:\n%s", output)
			}
		})
	}
}

func TestExampleFixturesVerifiersAcceptReferencesAndRejectPlausibleErrors(t *testing.T) {
	chdirRepoRoot(t)
	sc, err := scenario.Load("example-fixtures", config.Default())
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		taskID    string
		target    string
		reference string
		incorrect string
	}{
		{
			taskID: "parse-duration-units",
			target: "maintainer_tools/durations.py",
			reference: `"""Parse human-readable command timeouts."""


def parse_duration(value: str) -> float:
    """Return a duration in seconds."""
    if value.endswith("ms"):
        return float(value[:-2]) / 1000
    if value.endswith("s"):
        return float(value[:-1])
    if value.endswith("m"):
        return float(value[:-1]) * 60
    raise ValueError(f"unsupported duration: {value}")
`,
			incorrect: `"""Parse human-readable command timeouts."""


def parse_duration(value: str) -> float:
    """Return a duration in seconds."""
    if value.endswith("ms"):
        return float(value[:-2]) / 100
    if value.endswith("s"):
        return float(value[:-1])
    if value.endswith("m"):
        return float(value[:-1]) * 60
    raise ValueError(f"unsupported duration: {value}")
`,
		},
		{
			taskID: "redact-url-credentials",
			target: "maintainer_tools/urls.py",
			reference: `"""Prepare URLs for safe diagnostic output."""

import re


def redact_credentials(url: str) -> str:
    """Redact credential values from a URL query string."""
    return re.sub(r"([?&](?:token|api_key|password)=)[^&#]*", r"\1REDACTED", url)
`,
			incorrect: `"""Prepare URLs for safe diagnostic output."""

import re


def redact_credentials(url: str) -> str:
    """Redact credential values from a URL query string."""
    return re.sub(r"([?&](?:token|api_key|password)=)[^&]*", r"\1REDACTED", url)
`,
		},
		{
			taskID: "normalize-account-names",
			target: "maintainer_tools/accounts.py",
			reference: `"""Normalize account names used in generated file paths."""


def normalize_account_name(name: str) -> str:
    """Return the stable path form of an account display name."""
    return "-".join(name.split()).lower()
`,
			incorrect: `"""Normalize account names used in generated file paths."""


def normalize_account_name(name: str) -> str:
    """Return the stable path form of an account display name."""
    return name.strip().lower().replace(" ", "-")
`,
		},
		{
			taskID: "summarize-log-levels",
			target: "maintainer_tools/logs.py",
			reference: `"""Summarize bracketed levels in command logs."""

import re


LEVEL = re.compile(r"^\[(DEBUG|INFO|WARNING|ERROR)\]", re.IGNORECASE)


def count_log_levels(lines: list[str]) -> dict[str, int]:
    """Count recognized log levels."""
    counts: dict[str, int] = {}
    for line in lines:
        match = LEVEL.match(line)
        if match:
            level = match.group(1).upper()
            counts[level] = counts.get(level, 0) + 1
    return counts
`,
			incorrect: `"""Summarize bracketed levels in command logs."""

import re


LEVEL = re.compile(r"^\[(DEBUG|INFO|WARNING|ERROR)\]", re.IGNORECASE)


def count_log_levels(lines: list[str]) -> dict[str, int]:
    """Count recognized log levels."""
    counts: dict[str, int] = {}
    for line in lines:
        match = LEVEL.match(line)
        if match:
            level = match.group(1)
            counts[level] = counts.get(level, 0) + 1
    return counts
`,
		},
	}

	for _, tc := range cases {
		task := taskByID(t, sc, tc.taskID)
		for _, correction := range []struct {
			name        string
			content     string
			wantVerdict string
		}{
			{name: "reference", content: tc.reference, wantVerdict: "pass"},
			{name: "plausible-incorrect", content: tc.incorrect, wantVerdict: "fail"},
		} {
			t.Run(tc.taskID+"/"+correction.name, func(t *testing.T) {
				workDir := t.TempDir()
				if err := workspace.Seed(sc, workDir); err != nil {
					t.Fatal(err)
				}
				target := filepath.Join(workDir, "fixtures", filepath.FromSlash(tc.target))
				if err := os.WriteFile(target, []byte(correction.content), 0o644); err != nil {
					t.Fatal(err)
				}
				got, err := verifier.Run(context.Background(), sc, task, workDir, scenario.Variation{})
				if err != nil {
					t.Fatal(err)
				}
				if got != correction.wantVerdict {
					t.Fatalf("verdict = %q, want %q", got, correction.wantVerdict)
				}
			})
		}
	}
}

func TestDemoSkillProvidesGeneralMaintenanceGuidance(t *testing.T) {
	chdirRepoRoot(t)
	content, err := os.ReadFile(filepath.Join("examples", "skills", "demo-skill", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ToLower(string(content))
	for _, stage := range []string{"reproduce", "inspect", "fix", "recheck"} {
		if !strings.Contains(text, "## "+stage) {
			t.Errorf("demo skill does not include a %q stage", stage)
		}
	}
	for _, taskID := range []string{
		"parse-duration-units",
		"redact-url-credentials",
		"normalize-account-names",
		"summarize-log-levels",
	} {
		if strings.Contains(text, taskID) {
			t.Errorf("demo skill includes task-specific identifier %q", taskID)
		}
	}
}

func taskByID(t *testing.T, sc scenario.Scenario, id string) scenario.Task {
	t.Helper()
	for _, task := range sc.Tasks {
		if task.ID == id {
			return task
		}
	}
	t.Fatalf("task %q not found", id)
	return scenario.Task{}
}

func chdirRepoRoot(t *testing.T) {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			if err := os.Chdir(dir); err != nil {
				t.Fatal(err)
			}
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}
