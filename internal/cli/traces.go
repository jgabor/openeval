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
		writeTracesOutput(cmd.OutOrStdout(), traceID, cfg)
	},
}

func writeTracesOutput(w io.Writer, traceID string, cfg config.Config) {
	lines := []string{
		fmt.Sprintf("trace_id: %s", traceID),
		"otlp_service: openeval-agent",
	}
	if cfg.Telemetry.Endpoint != "" {
		lines = append(lines, fmt.Sprintf("otlp_endpoint: %s", cfg.Telemetry.Endpoint))
	}
	lines = append(lines,
		fmt.Sprintf("jaeger_ui: http://localhost:16686/trace/%s", telemetry.NormalizeTraceID(traceID)),
		"hint: open Jaeger UI and paste trace_id if your collector uses a different UI host",
	)
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
