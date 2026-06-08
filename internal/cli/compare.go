package cli

import (
	"fmt"
	"os"

	"github.com/jgabor/openeval/internal/compare"
	"github.com/spf13/cobra"
)

var compareCmd = &cobra.Command{
	Use:   "compare <run-dir-a> <run-dir-b>",
	Short: "Compare two runs with the same scenario",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		out, err := compare.Run(args[0], args[1])
		if err != nil {
			exitErr(err)
		}
		fmt.Fprint(os.Stdout, out)
	},
}
