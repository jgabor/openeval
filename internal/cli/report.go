package cli

import (
	"fmt"
	"os"

	"github.com/jgabor/openeval/internal/paths"
	"github.com/jgabor/openeval/internal/report"
	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report <run-dir>",
	Short: "Print a human-readable run summary",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		doc, err := report.LoadRun(args[0])
		if err != nil {
			exitErr(err)
		}
		fmt.Fprint(os.Stdout, report.Format(doc))
	},
}

func init() {
	_ = paths.ScorePath
}
