package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/jgabor/openeval/internal/config"
	"github.com/jgabor/openeval/internal/runcontext"
)

const (
	defaultInputCostPerMillion  = 3.0
	defaultOutputCostPerMillion = 15.0
)

type Cursor struct {
	command string
	cost    config.CursorCostConfig
}

func newCursor(cfg config.Config) (Cursor, error) {
	cmd := strings.TrimSpace(cfg.Agents.Cursor.Command)
	if cmd == "" {
		path, err := exec.LookPath("cursor-agent")
		if err != nil {
			return Cursor{}, fmt.Errorf(
				"cursor-agent not found on PATH: install the Cursor CLI, run `cursor-agent login`, or set agents.cursor.command in %s",
				configHint(),
			)
		}
		cmd = path
	} else if _, err := os.Stat(cmd); err != nil {
		return Cursor{}, fmt.Errorf("agents.cursor.command %q: %w", cmd, err)
	}
	return Cursor{command: cmd, cost: cfg.Agents.Cursor.Cost}, nil
}

func configHint() string {
	path, err := config.Path()
	if err != nil {
		return "your OpenEval config"
	}
	return path
}

func (c Cursor) Run(ctx context.Context, s Session) (float64, string, error) {
	args := []string{
		"-p",
		"--trust",
		"--workspace", s.WorkDir,
		"--output-format", "json",
	}
	for _, dir := range s.PluginDirs {
		args = append(args, "--plugin-dir", dir)
	}
	args = append(args, s.Task.Prompt)
	cmd := exec.CommandContext(ctx, c.command, args...)
	cmd.Env = runcontext.MergeEnvironment(
		os.Environ(),
		s.Run.Env(),
		runcontext.FromMap(s.Variation.Env),
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(stdout.String())
		}
		if msg == "" {
			msg = err.Error()
		}
		return 0, "", fmt.Errorf(
			"cursor-agent failed: %s (run `cursor-agent login` if this is an authentication error)",
			msg,
		)
	}
	cursorSessionID, usage, err := parseCursorJSON(stdout.Bytes())
	if err != nil {
		return 0, "", err
	}
	if cursorSessionID == "" {
		return 0, "", fmt.Errorf("cursor-agent returned no session_id or request_id in JSON output")
	}
	traceID := cursorSessionID
	if s.Run.Active() {
		traceID = s.Run.TraceID
	}
	_ = cursorSessionID
	return estimateCost(usage, c.cost), traceID, nil
}

type cursorResponse struct {
	Type      string `json:"type"`
	Subtype   string `json:"subtype"`
	IsError   bool   `json:"is_error"`
	SessionID string `json:"session_id"`
	RequestID string `json:"request_id"`
	Usage     struct {
		InputTokens      int `json:"inputTokens"`
		OutputTokens     int `json:"outputTokens"`
		CacheReadTokens  int `json:"cacheReadTokens"`
		CacheWriteTokens int `json:"cacheWriteTokens"`
	} `json:"usage"`
}

type cursorUsage struct {
	InputTokens      int
	OutputTokens     int
	CacheReadTokens  int
	CacheWriteTokens int
}

func parseCursorJSON(data []byte) (traceID string, usage cursorUsage, err error) {
	line := lastNonEmptyLine(data)
	if line == "" {
		return "", cursorUsage{}, fmt.Errorf("cursor-agent produced empty JSON output")
	}
	var resp cursorResponse
	if err := json.Unmarshal([]byte(line), &resp); err != nil {
		return "", cursorUsage{}, fmt.Errorf("parse cursor-agent JSON: %w", err)
	}
	if resp.IsError || resp.Subtype == "error" {
		return "", cursorUsage{}, fmt.Errorf("cursor-agent reported an error in JSON output")
	}
	traceID = resp.SessionID
	if traceID == "" {
		traceID = resp.RequestID
	}
	usage = cursorUsage{
		InputTokens:      resp.Usage.InputTokens,
		OutputTokens:     resp.Usage.OutputTokens,
		CacheReadTokens:  resp.Usage.CacheReadTokens,
		CacheWriteTokens: resp.Usage.CacheWriteTokens,
	}
	return traceID, usage, nil
}

func estimateCost(u cursorUsage, cfg config.CursorCostConfig) float64 {
	inRate := cfg.InputPerMillion
	outRate := cfg.OutputPerMillion
	if inRate == 0 {
		inRate = defaultInputCostPerMillion
	}
	if outRate == 0 {
		outRate = defaultOutputCostPerMillion
	}
	billableInput := u.InputTokens + u.CacheReadTokens + u.CacheWriteTokens
	return (float64(billableInput)/1e6)*inRate + (float64(u.OutputTokens)/1e6)*outRate
}

func lastNonEmptyLine(data []byte) string {
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			return line
		}
	}
	return ""
}
