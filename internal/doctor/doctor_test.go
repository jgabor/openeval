package doctor

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jgabor/openeval/internal/config"
)

func TestRunOpenCodePassesWithoutPaidModelCall(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))
	collector := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer collector.Close()

	logPath := filepath.Join(root, "agent.log")
	stub := writeStub(t, root, `#!/bin/sh
printf '%s\n' "$*" >> "`+logPath+`"
case "$*" in
  --version) printf '%s\n' '1.18.11' ;;
  'auth list') printf '%s\n' 'Credentials: test-provider' ;;
  models) printf '%s\n' 'opencode/big-pickle' ;;
  *) exit 99 ;;
esac
`)
	skillDir := filepath.Join(root, "skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: local-skill\ndescription: local test skill\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Agents.OpenCode.Command = stub
	cfg.Agents.OpenCode.NativeOTEL = true
	cfg.Telemetry.Endpoint = collector.URL + "/v1/traces"
	cfg.Skills.Aliases = map[string]string{"local-skill": skillDir}
	saveConfig(t, cfg)

	report := Run(context.Background(), "opencode", "")
	if report.Status != StatusPass || report.ExitCode != 0 {
		t.Fatalf("report = %+v, want pass with exit 0", report)
	}
	for _, check := range report.Checks {
		if check.Status != StatusPass {
			t.Fatalf("check = %+v, want pass", check)
		}
	}
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Split(strings.TrimSpace(string(data)), "\n"); len(got) != 3 || got[0] != "--version" || got[1] != "auth list" || got[2] != "models" {
		t.Fatalf("runtime calls = %v, want only version, auth list, and non-paid models listing", got)
	}
}

func TestRunTreatsUnreachableTelemetryAsWarning(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))
	stub := writeOpenCodeStub(t, root, true)
	cfg := config.Default()
	cfg.Agents.OpenCode.Command = stub
	cfg.Agents.OpenCode.NativeOTEL = true
	cfg.Telemetry.Endpoint = "http://127.0.0.1:1/v1/traces"
	saveConfig(t, cfg)

	report := Run(context.Background(), "opencode", "")
	if report.Status != StatusWarning || report.ExitCode != 0 {
		t.Fatalf("status = %s exit = %d, want warning exit 0", report.Status, report.ExitCode)
	}
	check := findCheck(t, report, "telemetry")
	if check.Status != StatusWarning || !strings.Contains(check.Remediation, "harness runs can continue") {
		t.Fatalf("telemetry check = %+v", check)
	}
}

func TestRunReportsAuthenticationFailureWithRemediation(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))
	collector := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer collector.Close()
	stub := writeOpenCodeStub(t, root, false)
	cfg := config.Default()
	cfg.Agents.OpenCode.Command = stub
	cfg.Agents.OpenCode.NativeOTEL = true
	cfg.Telemetry.Endpoint = collector.URL + "/v1/traces"
	saveConfig(t, cfg)

	report := Run(context.Background(), "opencode", "provider/intended")
	if report.Status != StatusFail || report.ExitCode != 1 {
		t.Fatalf("status = %s exit = %d, want fail exit 1", report.Status, report.ExitCode)
	}
	check := findCheck(t, report, "authentication")
	for _, want := range []string{"opencode auth login", "opencode auth list"} {
		if !strings.Contains(check.Remediation, want) {
			t.Fatalf("authentication remediation %q missing %q", check.Remediation, want)
		}
	}
	for _, check := range report.Checks {
		if check.ID == "model" {
			t.Fatalf("model catalog ran after authentication failed: %+v", check)
		}
	}
}

func TestRunReportsUnavailableResolvedModelWithSelectionAndAuthenticationRemediation(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))
	stub := writeOpenCodeStub(t, root, true)
	cfg := config.Default()
	cfg.Agents.OpenCode.Command = stub
	saveConfig(t, cfg)

	report := Run(context.Background(), "opencode", "missing-provider/missing-model")
	check := findCheck(t, report, "model")
	if check.Status != StatusFail || report.ExitCode != 1 {
		t.Fatalf("model check = %+v report=%+v, want fatal unavailable model", check, report)
	}
	for _, want := range []string{"missing-provider/missing-model", "--model", "agents.opencode.model", "opencode auth login", "opencode models"} {
		if !strings.Contains(check.Summary+" "+check.Remediation, want) {
			t.Fatalf("model diagnosis %+v missing %q", check, want)
		}
	}
}

func TestRunReportsModelCatalogCommandFailure(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))
	stub := writeStub(t, root, `#!/bin/sh
case "$*" in
  --version) printf '%s\n' '1.18.11' ;;
  'auth list') printf '%s\n' 'Credentials: test-provider' ;;
  models) echo 'catalog unavailable' >&2; exit 2 ;;
  *) exit 99 ;;
esac
`)
	cfg := config.Default()
	cfg.Agents.OpenCode.Command = stub
	saveConfig(t, cfg)

	report := Run(context.Background(), "opencode", "provider/model")
	check := findCheck(t, report, "model")
	if check.Status != StatusFail || !strings.Contains(check.Summary, "catalog unavailable") {
		t.Fatalf("model check = %+v, want catalog command failure", check)
	}
	for _, want := range []string{"opencode models", "opencode auth login"} {
		if !strings.Contains(check.Remediation, want) {
			t.Fatalf("model remediation %q missing %q", check.Remediation, want)
		}
	}
}

func TestRunCursorChecksRuntimeAndHooks(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))
	t.Setenv("HOME", filepath.Join(root, "home"))
	collector := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer collector.Close()
	stub := writeStub(t, root, "#!/bin/sh\nprintf '%s\\n' 'cursor-agent 2026.1'\n")
	hooksPath := filepath.Join(root, "home", ".cursor", "hooks.json")
	if err := os.MkdirAll(filepath.Dir(hooksPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(hooksPath, []byte(`{"version":1,"hooks":{"sessionStart":[{"command":"openeval hook --agent cursor"}]}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Agents.Cursor.Command = stub
	cfg.Telemetry.Endpoint = collector.URL + "/v1/traces"
	saveConfig(t, cfg)

	report := Run(context.Background(), "cursor", "")
	if report.Status != StatusPass || report.ExitCode != 0 {
		t.Fatalf("report = %+v, want Cursor pass", report)
	}
	if findCheck(t, report, "cursor_hooks").Status != StatusPass {
		t.Fatal("Cursor hooks did not pass")
	}
	for _, check := range report.Checks {
		if check.ID == "authentication" || check.ID == "native_otel" {
			t.Fatalf("unexpected OpenCode check in Cursor report: %+v", check)
		}
	}
}

func TestWriteJSONUsesStableSchemaAndStatuses(t *testing.T) {
	report := Report{
		SchemaVersion: SchemaVersion,
		Agent:         "opencode",
		Status:        StatusWarning,
		ExitCode:      0,
		Checks: []Check{{
			ID:          "telemetry",
			Status:      StatusWarning,
			Summary:     "unreachable",
			Remediation: "start collector",
		}},
	}
	var out bytes.Buffer
	if err := WriteJSON(&out, report); err != nil {
		t.Fatal(err)
	}
	body := out.String()
	for _, want := range []string{
		`"schema_version": "openeval.doctor.v1"`,
		`"status": "warning"`,
		`"exit_code": 0`,
		`"remediation": "start collector"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("JSON missing %q:\n%s", want, body)
		}
	}
}

func TestValidOpenCodeFrontmatterParsesYAML(t *testing.T) {
	body := "---\nname: \"quoted-skill\"\ndescription: |\n  A multiline description.\n---\n# Skill\n"
	if !validOpenCodeFrontmatter(body, "quoted-skill") {
		t.Fatal("valid quoted YAML frontmatter was rejected")
	}
	if validOpenCodeFrontmatter(body, "alias-name") {
		t.Fatal("frontmatter name mismatch was accepted")
	}
}

func TestEveryNonPassCheckHasRemediation(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))
	t.Setenv("PATH", t.TempDir())
	cfg := config.Default()
	cfg.Telemetry.Endpoint = ""
	saveConfig(t, cfg)

	report := Run(context.Background(), "opencode", "")
	if report.Status != StatusFail {
		t.Fatalf("status = %s, want fail", report.Status)
	}
	for _, check := range report.Checks {
		if check.Status != StatusPass && strings.TrimSpace(check.Remediation) == "" {
			t.Fatalf("non-pass check lacks remediation: %+v", check)
		}
	}
}

func writeOpenCodeStub(t *testing.T, root string, authOK bool) string {
	t.Helper()
	auth := "echo auth unavailable >&2; exit 1"
	if authOK {
		auth = "printf '%s\\n' 'Credentials: test-provider'"
	}
	return writeStub(t, root, `#!/bin/sh
case "$*" in
  --version) printf '%s\n' '1.18.11' ;;
  'auth list') `+auth+` ;;
  models) printf '%s\n' 'opencode/big-pickle' ;;
  *) exit 99 ;;
esac
`)
}

func writeStub(t *testing.T, root, body string) string {
	t.Helper()
	path := filepath.Join(root, "agent-stub.sh")
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func saveConfig(t *testing.T, cfg config.Config) {
	t.Helper()
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}
}

func findCheck(t *testing.T, report Report, id string) Check {
	t.Helper()
	for _, check := range report.Checks {
		if check.ID == id {
			return check
		}
	}
	t.Fatalf("check %q not found in %+v", id, report.Checks)
	return Check{}
}
