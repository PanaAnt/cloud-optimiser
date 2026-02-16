package mode

import (
	"fmt"

	"github.com/PanaAnt/cloud-optimiser/internal/config"
	"github.com/spf13/cobra"
)

var ShowModeCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current Cloud Optimiser mode",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			fmt.Println("Current mode: mock (default)")
			return
		}
		fmt.Println("Current mode:", cfg.Mode)
	},
}
