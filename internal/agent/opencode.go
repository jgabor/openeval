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
	}
	args = append(args, s.Task.Prompt)
	cmd := exec.CommandContext(ctx, o.command, args...)
	cmd.Env = runcontext.MergeEnvironment(
		os.Environ(),
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
		return 0, "", fmt.Errorf("opencode failed: %s", msg)
	}
	traceID, usage, err := parseOpenCodeJSON(stdout.Bytes())
	if err != nil {
		return 0, "", err
	}
	if traceID == "" {
		return 0, "", fmt.Errorf("opencode returned no sessionID in JSON output")
	}
	return estimateOpenCodeCost(usage, o.cost), traceID, nil
}

type openCodeResponse struct {
	Type      string `json:"type"`
	SessionID string `json:"sessionID"`
	Part      struct {
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
	CacheReadTokens  int
	CacheWriteTokens int
}

func parseOpenCodeJSON(data []byte) (traceID string, usage openCodeUsage, err error) {
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 64*1024), 16*1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var resp openCodeResponse
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			continue
		}
		if resp.SessionID != "" {
			traceID = resp.SessionID
		}
		if resp.Type == "step_finish" {
			usage.InputTokens += resp.Part.Tokens.Input
			usage.OutputTokens += resp.Part.Tokens.Output
			usage.CacheReadTokens += resp.Part.Tokens.Cache.Read
			usage.CacheWriteTokens += resp.Part.Tokens.Cache.Write
		}
	}
	if err := sc.Err(); err != nil {
		return "", openCodeUsage{}, fmt.Errorf("scan opencode JSON: %w", err)
	}
	return traceID, usage, nil
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
	return (float64(billableInput)/1e6)*inRate + (float64(u.OutputTokens)/1e6)*outRate
}
