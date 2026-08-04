package cli

import (
	"fmt"
	"io"

	"github.com/jgabor/openeval/internal/config"
	"github.com/jgabor/openeval/internal/paths"
	"github.com/jgabor/openeval/internal/score"
	"github.com/jgabor/openeval/internal/telemetry"
	"github.com/spf13/cobra"
)

var (
	tracesTask  string
	tracesRound int
)

var tracesCmd = &cobra.Command{
	Use:   "traces <run-dir>",
	Short: "Print trace reference for a task round",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		doc, err := score.Load(paths.ScorePath(args[0]))
		if err != nil {
			exitErr(err)
		}
		traceID, err := lookupTraceID(doc, tracesTask, tracesRound)
		if err != nil {
			exitErr(err)
		}
		cfg, err := config.Load()
		if err != nil {
			exitErr(err)
		}
		writeTracesOutput(cmd.OutOrStdout(), traceID, doc.Agent, cfg)
	},
}

func writeTracesOutput(w io.Writer, traceID, agent string, cfg config.Config) {
	lines := []string{
		fmt.Sprintf("trace_id: %s", traceID),
		"summary_trace: OpenEval direct summary span",
		"otlp_service: openeval-agent",
	}
	if cfg.Telemetry.Endpoint != "" {
		lines = append(lines, fmt.Sprintf("otlp_endpoint: %s", cfg.Telemetry.Endpoint))
	}
	lines = append(lines,
		fmt.Sprintf("jaeger_ui: http://localhost:16686/trace/%s", telemetry.NormalizeTraceID(traceID)),
		"summary_hint: open the direct trace ID above in your collector UI",
	)
	if agent == "opencode" && cfg.Agents.OpenCode.NativeOTEL {
		baseEndpoint, err := telemetry.OpenCodeOTLPBaseEndpoint(cfg.Telemetry.Endpoint)
		if err != nil {
			lines = append(lines, "native_opencode_setup_error: "+err.Error())
		} else {
			lines = append(lines,
				"native_opencode_trace: separate runtime-generated trace, not the OpenEval summary trace",
				"native_correlation: search resource attribute openeval.trace_id="+traceID,
				"native_otlp_endpoint: "+telemetry.OpenCodeOTLPTraceEndpoint(baseEndpoint),
				"native_privacy: OpenCode generates this payload; OpenEval masking does not apply",
			)
		}
	}
	for _, line := range lines {
		_, _ = fmt.Fprintln(w, line)
	}
}

func lookupTraceID(doc score.Document, taskID string, round int) (string, error) {
	for _, t := range doc.ByTask {
		if t.TaskID != taskID {
			continue
		}
		for _, r := range t.Rounds {
			if r.Round == round {
				if r.TraceID == "" {
					return "", fmt.Errorf("task %q round %d has no trace_id", taskID, round)
				}
				return r.TraceID, nil
			}
		}
	}
	return "", fmt.Errorf("task %q round %d not found in score.json", taskID, round)
}

func init() {
	tracesCmd.Flags().StringVar(&tracesTask, "task", "", "Task id")
	tracesCmd.Flags().IntVar(&tracesRound, "round", 1, "Round number")
	_ = tracesCmd.MarkFlagRequired("task")
}
