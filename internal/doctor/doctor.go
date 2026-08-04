package doctor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jgabor/openeval/internal/agent"
	"github.com/jgabor/openeval/internal/config"
	"gopkg.in/yaml.v3"
)

const (
	SchemaVersion            = "openeval.doctor.v1"
	supportedOpenCodeVersion = "1.18.11"
)

type Status string

const (
	StatusPass    Status = "pass"
	StatusWarning Status = "warning"
	StatusFail    Status = "fail"
)

type Check struct {
	ID          string `json:"id"`
	Status      Status `json:"status"`
	Summary     string `json:"summary"`
	Remediation string `json:"remediation,omitempty"`
}

type Report struct {
	SchemaVersion string  `json:"schema_version"`
	Agent         string  `json:"agent"`
	Status        Status  `json:"status"`
	ExitCode      int     `json:"exit_code"`
	Checks        []Check `json:"checks"`
}

func Run(ctx context.Context, agentName, modelOverride string) Report {
	if agentName == "" {
		agentName = "opencode"
	}
	report := Report{SchemaVersion: SchemaVersion, Agent: agentName}
	cfg := checkConfig(&report, agentName)
	checkModelOverride(&report, agentName, modelOverride, cfg)
	command, ok := checkRuntime(ctx, &report, agentName, cfg)
	if ok {
		checkVersion(ctx, &report, agentName, command)
		if agentName == "opencode" && checkOpenCodeAuth(ctx, &report, command) {
			checkOpenCodeModel(ctx, &report, command, modelOverride, cfg)
		}
	}
	checkSkills(&report, agentName, cfg)
	checkTelemetry(ctx, &report, cfg)
	switch agentName {
	case "opencode":
		checkOpenCodeNativeOTEL(&report, cfg)
	case "cursor":
		checkCursorHooks(&report)
	default:
		report.add(Check{
			ID:          "agent",
			Status:      StatusFail,
			Summary:     fmt.Sprintf("unsupported agent %q", agentName),
			Remediation: "use --agent opencode or --agent cursor",
		})
	}
	report.finish()
	return report
}

func checkModelOverride(report *Report, agentName, modelOverride string, cfg config.Config) {
	if modelOverride == "" || agentName == "opencode" {
		return
	}
	if _, err := agent.ResolveModel(agentName, modelOverride, "", cfg); err != nil {
		report.add(Check{
			ID:          "model",
			Status:      StatusFail,
			Summary:     err.Error(),
			Remediation: "remove `--model` or use `--agent opencode`",
		})
	}
}

func (r *Report) add(check Check) {
	r.Checks = append(r.Checks, check)
}

func (r *Report) finish() {
	r.Status = StatusPass
	for _, check := range r.Checks {
		if check.Status == StatusFail {
			r.Status = StatusFail
			r.ExitCode = 1
			return
		}
		if check.Status == StatusWarning {
			r.Status = StatusWarning
		}
	}
}

func checkConfig(report *Report, agent string) config.Config {
	cfg := config.Default()
	path, err := config.Path()
	if err != nil {
		report.add(Check{ID: "config", Status: StatusFail, Summary: err.Error(), Remediation: "set the platform configuration home and retry"})
		return cfg
	}
	_, statErr := os.Stat(path)
	loaded, loadErr := config.Load()
	if loadErr != nil {
		report.add(Check{
			ID:          "config",
			Status:      StatusFail,
			Summary:     loadErr.Error(),
			Remediation: fmt.Sprintf("fix %s or replace it with examples/config.minimal.yaml", path),
		})
		return cfg
	}
	cfg = loaded
	if os.IsNotExist(statErr) {
		report.add(Check{
			ID:          "config",
			Status:      StatusWarning,
			Summary:     fmt.Sprintf("using defaults; %s does not exist", path),
			Remediation: fmt.Sprintf("run `openeval instrument --agent %s` or copy examples/config.minimal.yaml to that path", agent),
		})
		return cfg
	}
	if statErr != nil {
		report.add(Check{ID: "config", Status: StatusFail, Summary: statErr.Error(), Remediation: fmt.Sprintf("make %s readable", path)})
		return cfg
	}
	report.add(Check{ID: "config", Status: StatusPass, Summary: path})
	return cfg
}

func checkRuntime(ctx context.Context, report *Report, agent string, cfg config.Config) (string, bool) {
	var command, binary, remediation string
	switch agent {
	case "opencode":
		command = strings.TrimSpace(cfg.Agents.OpenCode.Command)
		binary = "opencode"
		remediation = "install OpenCode 1.18.11 or set agents.opencode.command in the OpenEval config"
	case "cursor":
		command = strings.TrimSpace(cfg.Agents.Cursor.Command)
		binary = "cursor-agent"
		remediation = "install cursor-agent or set agents.cursor.command in the OpenEval config"
	default:
		return "", false
	}
	if command == "" {
		path, err := exec.LookPath(binary)
		if err != nil {
			report.add(Check{ID: "runtime", Status: StatusFail, Summary: fmt.Sprintf("%s not found on PATH", binary), Remediation: remediation})
			return "", false
		}
		command = path
	} else if _, err := os.Stat(command); err != nil {
		report.add(Check{ID: "runtime", Status: StatusFail, Summary: fmt.Sprintf("configured command %s: %v", command, err), Remediation: remediation})
		return "", false
	}
	if err := ctx.Err(); err != nil {
		report.add(Check{ID: "runtime", Status: StatusFail, Summary: err.Error(), Remediation: "retry doctor without an expired deadline"})
		return "", false
	}
	report.add(Check{ID: "runtime", Status: StatusPass, Summary: command})
	return command, true
}

func checkVersion(ctx context.Context, report *Report, agent, command string) {
	stdout, stderr, err := runCommand(ctx, command, "--version")
	if err != nil {
		report.add(Check{ID: "version", Status: StatusFail, Summary: commandFailure(err, stdout, stderr), Remediation: "verify the configured runtime command can execute `--version`"})
		return
	}
	versionText := strings.TrimSpace(stdout)
	if agent == "cursor" {
		if versionText == "" {
			report.add(Check{ID: "version", Status: StatusFail, Summary: "cursor-agent returned an empty version", Remediation: "verify the configured cursor-agent installation"})
			return
		}
		report.add(Check{ID: "version", Status: StatusPass, Summary: versionText})
		return
	}
	got, ok := parseVersion(versionText)
	if !ok {
		report.add(Check{ID: "version", Status: StatusFail, Summary: fmt.Sprintf("could not parse OpenCode version from %q", versionText), Remediation: "install OpenCode 1.18.11 and retry"})
		return
	}
	want, _ := parseVersion(supportedOpenCodeVersion)
	comparison := compareVersion(got, want)
	switch {
	case comparison == 0:
		report.add(Check{ID: "version", Status: StatusPass, Summary: "OpenCode " + got.String()})
	case got.major == want.major && comparison > 0:
		report.add(Check{
			ID:          "version",
			Status:      StatusWarning,
			Summary:     fmt.Sprintf("OpenCode %s is newer than the validated %s contract", got, want),
			Remediation: "if diagnosis or JSONL parsing fails, install OpenCode 1.18.11",
		})
	default:
		report.add(Check{
			ID:          "version",
			Status:      StatusFail,
			Summary:     fmt.Sprintf("OpenCode %s does not satisfy the validated %s contract", got, want),
			Remediation: "install OpenCode 1.18.11 and retry",
		})
	}
}

func checkOpenCodeAuth(ctx context.Context, report *Report, command string) bool {
	stdout, stderr, err := runCommand(ctx, command, "auth", "list")
	if err != nil {
		report.add(Check{ID: "authentication", Status: StatusFail, Summary: commandFailure(err, stdout, stderr), Remediation: "run `opencode auth login`, verify with `opencode auth list`, then retry"})
		return false
	}
	clean := strings.TrimSpace(stripANSI(stdout))
	lower := strings.ToLower(clean)
	if clean == "" || strings.Contains(lower, "no credentials") || strings.Contains(lower, "0 credentials") {
		report.add(Check{ID: "authentication", Status: StatusFail, Summary: "`opencode auth list` found no usable credentials", Remediation: "run `opencode auth login`, verify with `opencode auth list`, then retry"})
		return false
	}
	report.add(Check{ID: "authentication", Status: StatusPass, Summary: "`opencode auth list` completed with configured credentials"})
	return true
}

func checkOpenCodeModel(ctx context.Context, report *Report, command, modelOverride string, cfg config.Config) {
	model, err := agent.ResolveModel("opencode", modelOverride, "", cfg)
	if err != nil {
		report.add(Check{ID: "model", Status: StatusFail, Summary: err.Error(), Remediation: "select a provider/model shown by `opencode models`"})
		return
	}
	stdout, stderr, err := runCommand(ctx, command, "models")
	if err != nil {
		report.add(Check{
			ID:          "model",
			Status:      StatusFail,
			Summary:     fmt.Sprintf("could not list OpenCode models while checking %s: %s", model, commandFailure(err, stdout, stderr)),
			Remediation: "verify `opencode models` succeeds; if provider credentials are missing, run `opencode auth login` and retry",
		})
		return
	}
	for _, available := range strings.Split(stripANSI(stdout), "\n") {
		if strings.TrimSpace(available) == model {
			report.add(Check{ID: "model", Status: StatusPass, Summary: fmt.Sprintf("resolved OpenCode model %s is available", model)})
			return
		}
	}
	report.add(Check{
		ID:      "model",
		Status:  StatusFail,
		Summary: fmt.Sprintf("resolved OpenCode model %s is not available in `opencode models`", model),
		Remediation: fmt.Sprintf(
			"select a listed provider/model with `--model` or agents.opencode.model; to use provider %s, run `opencode auth login` and verify the model appears in `opencode models`",
			strings.SplitN(model, "/", 2)[0],
		),
	})
}

func checkSkills(report *Report, agent string, cfg config.Config) {
	names := make([]string, 0, len(cfg.Skills.Aliases))
	for name := range cfg.Skills.Aliases {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		dir, err := cfg.ResolveSkillPath(name)
		if err != nil {
			report.add(Check{ID: "skills", Status: StatusFail, Summary: err.Error(), Remediation: fmt.Sprintf("fix skills.aliases.%s to name a readable skill directory", name)})
			return
		}
		skillPath := filepath.Join(dir, "SKILL.md")
		data, err := os.ReadFile(skillPath)
		if err != nil {
			report.add(Check{ID: "skills", Status: StatusFail, Summary: fmt.Sprintf("skill %q: %v", name, err), Remediation: fmt.Sprintf("add a readable SKILL.md under %s", dir)})
			return
		}
		if agent == "opencode" && !validOpenCodeFrontmatter(string(data), name) {
			report.add(Check{ID: "skills", Status: StatusFail, Summary: fmt.Sprintf("skill %q has invalid OpenCode frontmatter", name), Remediation: "add frontmatter with matching name and a non-empty description"})
			return
		}
	}
	if len(names) == 0 {
		report.add(Check{ID: "skills", Status: StatusPass, Summary: "no skills.aliases entries configured"})
		return
	}
	report.add(Check{ID: "skills", Status: StatusPass, Summary: fmt.Sprintf("resolved %d skills.aliases entries", len(names))})
}

func validOpenCodeFrontmatter(body, name string) bool {
	if !strings.HasPrefix(body, "---\n") {
		return false
	}
	end := strings.Index(body[4:], "\n---")
	if end < 0 {
		return false
	}
	frontmatter := body[4 : 4+end]
	var metadata struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
	}
	if err := yaml.Unmarshal([]byte(frontmatter), &metadata); err != nil {
		return false
	}
	return metadata.Name == name && strings.TrimSpace(metadata.Description) != ""
}

func checkTelemetry(ctx context.Context, report *Report, cfg config.Config) {
	endpoint := strings.TrimSpace(cfg.Telemetry.Endpoint)
	if endpoint == "" {
		report.add(Check{ID: "telemetry", Status: StatusPass, Summary: "OTLP export is disabled; harness evaluation does not require it"})
		return
	}
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		report.add(Check{ID: "telemetry", Status: StatusFail, Summary: fmt.Sprintf("invalid telemetry.endpoint %q", endpoint), Remediation: "set telemetry.endpoint to an absolute HTTP(S) OTLP trace endpoint such as http://localhost:4318/v1/traces"})
		return
	}
	address := parsed.Host
	if parsed.Port() == "" {
		port := "80"
		if parsed.Scheme == "https" {
			port = "443"
		}
		address = net.JoinHostPort(parsed.Hostname(), port)
	}
	dialer := net.Dialer{Timeout: time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		report.add(Check{
			ID:          "telemetry",
			Status:      StatusWarning,
			Summary:     fmt.Sprintf("OTLP endpoint %s is unreachable: %v", endpoint, err),
			Remediation: "start the collector or update telemetry.endpoint; harness runs can continue without OTLP",
		})
		return
	}
	_ = conn.Close()
	report.add(Check{ID: "telemetry", Status: StatusPass, Summary: "OTLP endpoint is reachable at " + address})
}

func checkOpenCodeNativeOTEL(report *Report, cfg config.Config) {
	if !cfg.Agents.OpenCode.NativeOTEL {
		report.add(Check{
			ID:          "native_otel",
			Status:      StatusWarning,
			Summary:     "native OpenCode OTEL is disabled; OpenEval summary spans remain available",
			Remediation: "run `openeval instrument --agent opencode` to opt in; review the runtime-generated payload privacy boundary first",
		})
		return
	}
	report.add(Check{
		ID:      "native_otel",
		Status:  StatusPass,
		Summary: "native OpenCode OTEL is enabled; harness runs attach OpenEval correlation attributes; external sessions require OTEL_EXPORTER_OTLP_ENDPOINT",
	})
}

func checkCursorHooks(report *Report) {
	home, err := os.UserHomeDir()
	if err != nil {
		report.add(Check{ID: "cursor_hooks", Status: StatusFail, Summary: err.Error(), Remediation: "set HOME and retry"})
		return
	}
	path := filepath.Join(home, ".cursor", "hooks.json")
	data, err := os.ReadFile(path)
	if err != nil {
		report.add(Check{ID: "cursor_hooks", Status: StatusWarning, Summary: fmt.Sprintf("%s: %v", path, err), Remediation: "run `openeval instrument --agent cursor` to merge OpenEval hooks"})
		return
	}
	var doc any
	if json.Unmarshal(data, &doc) != nil || !bytes.Contains(data, []byte("hook --agent cursor")) {
		report.add(Check{ID: "cursor_hooks", Status: StatusWarning, Summary: fmt.Sprintf("%s has no valid OpenEval hook entries", path), Remediation: "run `openeval instrument --agent cursor` to merge OpenEval hooks"})
		return
	}
	report.add(Check{ID: "cursor_hooks", Status: StatusPass, Summary: path})
}

func runCommand(ctx context.Context, command string, args ...string) (stdout, stderr string, err error) {
	commandCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(commandCtx, command, args...)
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err = cmd.Run()
	if commandCtx.Err() != nil {
		err = commandCtx.Err()
	}
	return strings.TrimSpace(out.String()), strings.TrimSpace(errOut.String()), err
}

func commandFailure(err error, stdout, stderr string) string {
	detail := stderr
	if detail == "" {
		detail = stdout
	}
	if detail == "" {
		detail = err.Error()
	}
	return detail
}

var (
	versionPattern = regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)
	ansiPattern    = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)
)

type version struct {
	major int
	minor int
	patch int
}

func parseVersion(raw string) (version, bool) {
	match := versionPattern.FindStringSubmatch(raw)
	if len(match) != 4 {
		return version{}, false
	}
	major, _ := strconv.Atoi(match[1])
	minor, _ := strconv.Atoi(match[2])
	patch, _ := strconv.Atoi(match[3])
	return version{major: major, minor: minor, patch: patch}, true
}

func compareVersion(a, b version) int {
	for _, pair := range [][2]int{{a.major, b.major}, {a.minor, b.minor}, {a.patch, b.patch}} {
		if pair[0] < pair[1] {
			return -1
		}
		if pair[0] > pair[1] {
			return 1
		}
	}
	return 0
}

func (v version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.major, v.minor, v.patch)
}

func stripANSI(value string) string {
	return ansiPattern.ReplaceAllString(value, "")
}
