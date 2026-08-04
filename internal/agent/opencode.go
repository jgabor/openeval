package agent

import (
	"bufio"
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

type OpenCode struct {
	command string
	cost    config.OpenCodeCostConfig
}

func newOpenCode(cfg config.Config) (OpenCode, error) {
	cmd := strings.TrimSpace(cfg.Agents.OpenCode.Command)
	if cmd == "" {
		path, err := exec.LookPath("opencode")
		if err != nil {
			return OpenCode{}, fmt.Errorf(
				"opencode not found on PATH: install the opencode CLI and authenticate, or set agents.opencode.command in %s",
				configHint(),
			)
		}
		cmd = path
	} else if _, err := os.Stat(cmd); err != nil {
		return OpenCode{}, fmt.Errorf("agents.opencode.command %q: %w", cmd, err)
	}
	return OpenCode{command: cmd, cost: cfg.Agents.OpenCode.Cost}, nil
}

func (o OpenCode) Run(ctx context.Context, s Session) (float64, string, error) {
	args := []string{
		"run",
		"--format", "json",
		"--dir", s.WorkDir,
		"--auto",
	}
	args = append(args, s.Task.Prompt)
	cmd := exec.CommandContext(ctx, o.command, args...)
	cmd.Env = runcontext.Environment(os.Environ(), s.Variation.Env, s.Run)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return 0, "", newOpenCodeRunError(
				"execution",
				ctx.Err().Error(),
				"retry with enough time for the OpenCode run to finish",
			)
		}
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(stdout.String())
		}
		if msg == "" {
			msg = err.Error()
		}
		return 0, "", newOpenCodeRunError("execution", msg, openCodeAuthRemediation)
	}
	openCodeSessionID, usage, err := parseOpenCodeJSON(stdout.Bytes())
	if err != nil {
		return 0, "", err
	}
	traceID := openCodeSessionID
	if s.Run.Active() {
		traceID = s.Run.TraceID
	}
	return estimateOpenCodeCost(usage, o.cost), traceID, nil
}

const (
	openCodeAuthRemediation   = "check credentials with `opencode auth list`; authenticate with `opencode auth login`, then retry"
	openCodeOutputRemediation = "verify the configured command is OpenCode 1.18.11 and supports `run --format json --dir <workspace> --auto`"
)

type OpenCodeRunError struct {
	Kind        string
	Detail      string
	Remediation string
}

func (e *OpenCodeRunError) Error() string {
	return fmt.Sprintf("opencode %s error: %s; remediation: %s", e.Kind, e.Detail, e.Remediation)
}

func newOpenCodeRunError(kind, detail, remediation string) *OpenCodeRunError {
	return &OpenCodeRunError{Kind: kind, Detail: detail, Remediation: remediation}
}

type openCodeResponse struct {
	Type      string          `json:"type"`
	SessionID string          `json:"sessionID"`
	Error     json.RawMessage `json:"error"`
	Info      struct {
		Error json.RawMessage `json:"error"`
	} `json:"info"`
	Part struct {
		Type   string `json:"type"`
		Tokens struct {
			Input     int `json:"input"`
			Output    int `json:"output"`
			Reasoning int `json:"reasoning"`
			Cache     struct {
				Read  int `json:"read"`
				Write int `json:"write"`
			} `json:"cache"`
		} `json:"tokens"`
		Cost float64 `json:"cost"`
	} `json:"part"`
}

type openCodeUsage struct {
	InputTokens      int
	OutputTokens     int
	ReasoningTokens  int
	CacheReadTokens  int
	CacheWriteTokens int
}

func parseOpenCodeJSON(data []byte) (traceID string, usage openCodeUsage, err error) {
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 64*1024), 16*1024*1024)
	lineNumber := 0
	eventCount := 0
	for sc.Scan() {
		lineNumber++
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var resp openCodeResponse
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			return "", openCodeUsage{}, newOpenCodeRunError(
				"output",
				fmt.Sprintf("JSONL line %d is invalid: %v", lineNumber, err),
				openCodeOutputRemediation,
			)
		}
		eventCount++
		if resp.Type == "" {
			return "", openCodeUsage{}, newOpenCodeRunError(
				"output",
				fmt.Sprintf("JSONL line %d has no event type", lineNumber),
				openCodeOutputRemediation,
			)
		}
		if resp.SessionID != "" {
			if traceID != "" && traceID != resp.SessionID {
				return "", openCodeUsage{}, newOpenCodeRunError(
					"output",
					fmt.Sprintf("JSONL contains multiple session IDs %q and %q", traceID, resp.SessionID),
					openCodeOutputRemediation,
				)
			}
			traceID = resp.SessionID
		}
		if detail := openCodeEventError(resp); detail != "" {
			return "", openCodeUsage{}, newOpenCodeRunError("runtime", detail, openCodeAuthRemediation)
		}
		if resp.Type == "step_finish" {
			usage.InputTokens += resp.Part.Tokens.Input
			usage.OutputTokens += resp.Part.Tokens.Output
			usage.ReasoningTokens += resp.Part.Tokens.Reasoning
			usage.CacheReadTokens += resp.Part.Tokens.Cache.Read
			usage.CacheWriteTokens += resp.Part.Tokens.Cache.Write
		}
	}
	if err := sc.Err(); err != nil {
		return "", openCodeUsage{}, newOpenCodeRunError("output", fmt.Sprintf("scan JSONL: %v", err), openCodeOutputRemediation)
	}
	if eventCount == 0 {
		return "", openCodeUsage{}, newOpenCodeRunError("output", "produced no JSON events", openCodeOutputRemediation)
	}
	if traceID == "" {
		return "", openCodeUsage{}, newOpenCodeRunError("output", "JSON events contain no sessionID", openCodeOutputRemediation)
	}
	return traceID, usage, nil
}

func openCodeEventError(resp openCodeResponse) string {
	if resp.Type == "error" {
		if detail := rawOpenCodeError(resp.Error); detail != "" {
			return detail
		}
		return "reported an error event without details"
	}
	return rawOpenCodeError(resp.Info.Error)
}

func rawOpenCodeError(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var message string
	if err := json.Unmarshal(raw, &message); err == nil {
		return message
	}
	return string(raw)
}

func estimateOpenCodeCost(u openCodeUsage, cfg config.OpenCodeCostConfig) float64 {
	inRate := cfg.InputPerMillion
	outRate := cfg.OutputPerMillion
	if inRate == 0 {
		inRate = defaultInputCostPerMillion
	}
	if outRate == 0 {
		outRate = defaultOutputCostPerMillion
	}
	billableInput := u.InputTokens + u.CacheReadTokens + u.CacheWriteTokens
	billableOutput := u.OutputTokens + u.ReasoningTokens
	return (float64(billableInput)/1e6)*inRate + (float64(billableOutput)/1e6)*outRate
}
