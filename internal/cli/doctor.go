package cli

import (
	"context"
	"fmt"

	"github.com/jgabor/openeval/internal/doctor"
	"github.com/spf13/cobra"
)

var (
	doctorAgent string
	doctorJSON  bool
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose agent, config, skill, and telemetry setup",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		report := doctor.Run(context.Background(), doctorAgent)
		if doctorJSON {
			if err := doctor.WriteJSON(cmd.OutOrStdout(), report); err != nil {
				exitErr(err)
			}
		} else {
			doctor.WriteHuman(cmd.OutOrStdout(), report)
		}
		if report.ExitCode != 0 {
			exitErr(fmt.Errorf("doctor found fatal setup errors"))
		}
	},
}

func init() {
	doctorCmd.Flags().StringVar(&doctorAgent, "agent", "opencode", "Agent runtime to diagnose (opencode, cursor)")
	doctorCmd.Flags().BoolVar(&doctorJSON, "json", false, "Write a stable JSON report")
}
