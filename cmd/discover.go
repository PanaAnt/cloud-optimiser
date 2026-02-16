package cmd

import (
	"context"
	"fmt"

	"github.com/PanaAnt/cloud-optimiser/internal/awsclient"
	"github.com/PanaAnt/cloud-optimiser/internal/config"
	"github.com/PanaAnt/cloud-optimiser/internal/logging"
	"github.com/spf13/cobra"
)

var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover EC2 instances (mock or real AWS)",
	Long: `The discover command lists EC2 instances using either:
  • Mock data (default, safe)
  • Real AWS API (when enabled via config or flag)`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// 1. Load config & resolve initial mode
		cfg, err := config.LoadConfig()
		if err != nil {
			logging.Warn(fmt.Sprintf("Could not load config: %v, defaulting to MOCK mode", err))
			cfg = config.AppConfig{Mode: "mock"}
		}

		useMockMode := useMock || cfg.Mode == "mock"

		// 2. AWS readiness check (only if not already in mock mode)
		if !useMockMode {
			if err := awsclient.CanUseRealAWS(ctx, awsProfile); err != nil {
				fmt.Println("AWS unavailable – switching to MOCK mode.")
				logging.DebugErr("AWS readiness check failed", err)
				useMockMode = true
			}
		}

		// 3. Display mode
		fmt.Println("MODE:", map[bool]string{true: "Mock", false: "Real AWS"}[useMockMode])
		fmt.Println()

		// 4. Create EC2 client
		client, err := awsclient.New(ctx, awsclient.Config{
			UseMock: useMockMode,
			Profile: awsProfile,
		})
		if err != nil {
			fmt.Printf("Failed to create EC2 client: %v\n", err)
			logging.DebugErr("EC2 client creation failed", err)
			return
		}

		// 5. Fetch instances
		instances, err := client.ListInstances(ctx)
		if err != nil {
			fmt.Printf("Failed to list instances: %v\n", err)
			logging.DebugErr("EC2 ListInstances failed", err)
			return
		}

		if len(instances) == 0 {
			fmt.Println("No EC2 instances found.")
			return
		}

		// 6. Output
		fmt.Println("Discovered EC2 Instances:")
		for _, inst := range instances {
			fmt.Printf(" - %s (%s) [%s]\n", inst.ID, inst.InstanceType, inst.State)
		}
		
		if client.IsMock() {
			fmt.Println("\n[Note: Using mock data]")
		}
	},
}

func init() {
	rootCmd.AddCommand(discoverCmd)
}
