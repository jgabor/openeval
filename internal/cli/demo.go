package cli

import (
	"fmt"

	"github.com/jgabor/openeval/internal/demo"
	"github.com/spf13/cobra"
)

var (
	demoAgent    string
	demoScenario string
	demoRounds   int
	demoOut      string
	demoDryRun   bool
)

var demoCmd = &cobra.Command{
	Use:   "demo",
	Short: "Run the OpenCode-first comparison demonstration",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		result, err := demo.Run(cmd.Context(), demo.Options{
			Agent:    demoAgent,
			Scenario: demoScenario,
			Rounds:   demoRounds,
			Out:      demoOut,
			DryRun:   demoDryRun,
		})
		if err != nil {
			exitErr(err)
		}
		if result.DryRun {
			_, _ = fmt.Fprint(cmd.OutOrStdout(), result.Plan.Format())
			return
		}
		_, _ = fmt.Fprint(cmd.OutOrStdout(), result.Comparison)
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "demo evidence root: %s\n", result.Plan.Root)
	},
}

func init() {
	demoCmd.Flags().StringVar(&demoAgent, "agent", "opencode", "Agent runtime (opencode, cursor, mock)")
	demoCmd.Flags().StringVar(&demoScenario, "scenario", "example-fixtures", "Scenario id or path")
	demoCmd.Flags().IntVar(&demoRounds, "rounds", 3, "Attempts per task in each arm")
	demoCmd.Flags().StringVar(&demoOut, "out", "", "Parent directory for unique demo evidence")
	demoCmd.Flags().BoolVar(&demoDryRun, "dry-run", false, "Print exact commands and paths without execution or filesystem mutation")
}
