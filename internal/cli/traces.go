package cli

import (
	"fmt"

	"github.com/jgabor/openeval/internal/paths"
	"github.com/jgabor/openeval/internal/score"
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
		for _, t := range doc.ByTask {
			if t.TaskID != tracesTask {
				continue
			}
			for _, r := range t.Rounds {
				if r.Round == tracesRound {
					fmt.Printf("trace_id: %s\n", r.TraceID)
					fmt.Printf("otlp_service: openeval-agent\n")
					return
				}
			}
		}
		exitErr(fmt.Errorf("task %q round %d not found in score.json", tracesTask, tracesRound))
	},
}

func init() {
	tracesCmd.Flags().StringVar(&tracesTask, "task", "", "Task id")
	tracesCmd.Flags().IntVar(&tracesRound, "round", 1, "Round number")
	_ = tracesCmd.MarkFlagRequired("task")
}
