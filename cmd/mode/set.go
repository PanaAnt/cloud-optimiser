package mode

import (
	"fmt"

	"github.com/PanaAnt/cloud-optimiser/internal/config"
	"github.com/spf13/cobra"
)

var SetModeCmd = &cobra.Command{
	Use:   "set [mock|real]",
	Short: "Set Cloud Optimiser mode",
	Args:  cobra.ExactArgs(1),
	Example: `  cloud-optimiser mode set mock
  cloud-optimiser mode set real`,
	Run: func(cmd *cobra.Command, args []string) {
		mode := args[0]
		
		// Validate mode
		if mode != "mock" && mode != "real" {
			fmt.Println("Error: Invalid mode. Use: mock or real")
			return
		}

		// Save configuration
		cfg := config.AppConfig{Mode: mode}
		if err := config.SaveConfig(cfg); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			return
		}
		
		fmt.Printf("Mode successfully updated to: %s\n", mode)
	},
}
