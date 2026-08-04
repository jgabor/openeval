package cli

import (
	"context"
	"fmt"

	"github.com/jgabor/openeval/internal/runner"
	"github.com/spf13/cobra"
)

var (
	runScenario  string
	runAgent     string
	runVariation string
	runRounds    int
	runOut       string
	runImage     string
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute a scenario against an agent",
	Run: func(cmd *cobra.Command, args []string) {
		if runScenario == "" {
			exitErr(fmt.Errorf("--scenario is required"))
		}
		if runAgent == "" {
			exitErr(fmt.Errorf("--agent is required"))
		}
		if runImage != "" {
			exitErr(fmt.Errorf("docker --image runs are not implemented yet"))
		}
		switch runScenario {
		case "deepswe", "margin-eval":
			exitErr(fmt.Errorf("external scenario %q is not integrated yet", runScenario))
		}
		res, err := runner.Run(context.Background(), runner.Options{
			Scenario:  runScenario,
			Agent:     runAgent,
			Variation: runVariation,
			Rounds:    runRounds,
			Out:       runOut,
			Image:     runImage,
		})
		if err != nil {
			exitErr(err)
		}
		fmt.Printf("run complete: %s\n", res.RunDir)
	},
}

func init() {
	runCmd.Flags().StringVar(&runScenario, "scenario", "", "Scenario id or path")
	runCmd.Flags().StringVar(&runAgent, "agent", "", "Agent runtime (opencode, cursor, mock for CI)")
	runCmd.Flags().StringVar(&runVariation, "variation", "", "Config variation label")
	runCmd.Flags().IntVar(&runRounds, "rounds", 3, "Attempts per task")
	runCmd.Flags().StringVar(&runOut, "out", "", "Override output directory")
	runCmd.Flags().StringVar(&runImage, "image", "", "Docker image for packaged runs (planned; not implemented)")
	_ = runCmd.MarkFlagRequired("scenario")
	_ = runCmd.MarkFlagRequired("agent")
}
