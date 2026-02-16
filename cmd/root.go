package cmd

import (
	"fmt"
	"os"

	modecmd "github.com/PanaAnt/cloud-optimiser/cmd/mode"
	"github.com/PanaAnt/cloud-optimiser/internal/logging"
	"github.com/spf13/cobra"
)

var (
	useMock    bool
	debug      bool
	awsProfile string
)

var rootCmd = &cobra.Command{
	Use:   "cloud-optimiser",
	Short: "A CLI tool for analysing EC2 cost and utilisation",
	Long: `cloud-optimiser analyses AWS EC2 instances using
CloudWatch CPU metrics and Cost Explorer cost data.

Supports:
  • Mock mode (safe, default)
  • Real AWS API mode
  • AWS CLI profiles
  • Filters, sorting, JSON output`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logging.Verbose = debug

		if debug {
			fmt.Println("Debug mode enabled")
		}

		if awsProfile != "" {
			logging.Debug("Using AWS profile: " + awsProfile)
		}

		if useMock {
			logging.Debug("Mode override: MOCK (via --use-mock flag)")
		}
	},
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(
		&useMock,
		"use-mock",
		false,
		"Force mock mode for this run (overrides config)",
	)

	rootCmd.PersistentFlags().BoolVar(
		&debug,
		"debug",
		false,
		"Enable debug logging output",
	)

	rootCmd.PersistentFlags().StringVar(
		&awsProfile,
		"profile",
		"",
		"AWS CLI profile to use (from ~/.aws/credentials)",
	)

	rootCmd.AddCommand(modecmd.ModeCmd)
}