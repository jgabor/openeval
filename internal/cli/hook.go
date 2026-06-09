package cli

import (
	"context"
	"fmt"
	"os"

	cursorhook "github.com/jgabor/openeval/internal/hooks/cursor"
	"github.com/spf13/cobra"
)

var hookAgent string

var hookCmd = &cobra.Command{
	Use:    "hook",
	Short:  "Agent hook entrypoint for OTLP export",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		switch hookAgent {
		case "cursor":
			if err := cursorhook.Run(context.Background(), os.Stdin); err != nil {
				exitErr(err)
			}
		default:
			exitErr(fmt.Errorf("unsupported hook agent %q (supported: cursor)", hookAgent))
		}
	},
}

func init() {
	hookCmd.Flags().StringVar(&hookAgent, "agent", "", "Agent runtime for this hook invocation")
	_ = hookCmd.MarkFlagRequired("agent")
	rootCmd.AddCommand(hookCmd)
}
