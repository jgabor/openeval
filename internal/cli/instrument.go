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
	Short: "Install hooks and default config for supported agents",
	Run: func(cmd *cobra.Command, args []string) {
		var agents []string
		if instrumentAll {
			agents = []string{"cursor"}
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
		for _, a := range agents {
			if err := instrument.Install(a, cfg); err != nil {
				exitErr(err)
			}
			fmt.Printf("instrumented %s\n", a)
		}
		path, _ := config.Path()
		fmt.Printf("config: %s\n", path)
	},
}

func init() {
	instrumentCmd.Flags().BoolVar(&instrumentAll, "all", false, "Instrument all detected agents")
	instrumentCmd.Flags().StringVar(&instrumentAgent, "agent", "", "Instrument one agent")
}
