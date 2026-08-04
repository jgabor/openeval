package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "openeval",
	Short: "Platform-agnostic instrumentation and evaluation for CLI coding agents",
	Long:  "Instrument agents, run scenarios, score sessions, and compare variations.",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(instrumentCmd)
	rootCmd.AddCommand(reportCmd)
	rootCmd.AddCommand(compareCmd)
	rootCmd.AddCommand(tracesCmd)
	rootCmd.AddCommand(doctorCmd)
}

func exitErr(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
