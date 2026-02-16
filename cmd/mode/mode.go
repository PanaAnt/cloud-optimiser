package mode

import "github.com/spf13/cobra"

var ModeCmd = &cobra.Command{
    Use:   "mode",
    Short: "Manage Cloud Optimiser mode (mock or real)",
    Long: `cloud-optimiser supports two operating modes:

  • mock: Uses local JSON files for EC2, CloudWatch, and Cost Explorer data.
            Safe for development and does not require AWS credentials.

  • real: Uses live AWS APIs to analyse your actual EC2 infrastructure.

The active mode is stored in the app configuration file so it persists
between command runs.`,
	Example: `
  # Display current mode
  cloud-optimiser mode show

  # Switch to mock mode
  cloud-optimiser mode set mock

  # Switch to real AWS mode
  cloud-optimiser mode set real`,
}


func init() {
    ModeCmd.AddCommand(SetModeCmd)
    ModeCmd.AddCommand(ShowModeCmd)
}
