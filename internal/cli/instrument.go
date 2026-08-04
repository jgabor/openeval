package cli

import (
	"fmt"

	"github.com/jgabor/openeval/internal/config"
	"github.com/jgabor/openeval/internal/instrument"
	"github.com/spf13/cobra"
)

var (
	instrumentAll   bool
	instrumentAgent string
)

var instrumentCmd = &cobra.Command{
	Use:   "instrument",
	Short: "Configure telemetry for supported agents",
	Run: func(cmd *cobra.Command, args []string) {
		var agents []string
		if instrumentAll {
			agents = []string{"opencode", "cursor"}
		} else if instrumentAgent != "" {
			agents = []string{instrumentAgent}
		} else {
			exitErr(fmt.Errorf("pass --all or --agent"))
		}
		cfg := config.Default()
		if existing, err := config.Load(); err == nil {
			cfg = existing
		}
		if err := config.Save(cfg); err != nil {
			exitErr(err)
		}
		results := make([]instrument.Result, 0, len(agents))
		for _, a := range agents {
			result, err := instrument.Install(a, &cfg)
			if err != nil {
				exitErr(err)
			}
			results = append(results, result)
		}
		if err := config.Save(cfg); err != nil {
			exitErr(err)
		}
		for _, result := range results {
			fmt.Printf("instrumented %s\n", result.Agent)
			for _, message := range result.Messages {
				fmt.Println(message)
			}
		}
		path, _ := config.Path()
		fmt.Printf("config: %s\n", path)
	},
}

func init() {
	instrumentCmd.Flags().BoolVar(&instrumentAll, "all", false, "Configure OpenCode and Cursor")
	instrumentCmd.Flags().StringVar(&instrumentAgent, "agent", "", "Configure one agent (opencode, cursor)")
}
