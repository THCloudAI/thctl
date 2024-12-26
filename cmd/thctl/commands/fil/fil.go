package fil

import (
	"github.com/spf13/cobra"
	"github.com/THCloudAI/thctl/cmd/thctl/commands/fil/sectors"
)

// NewFilCmd creates a new fil command
func NewFilCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fil",
		Short: "Filecoin operations",
		Long: `Manage Filecoin operations including sector management and financial calculations.

Examples:
  # Calculate sector penalty
  thctl fil sectors penalty --miner f01234 --sector 1

  # Query vested funds
  thctl fil sectors vested --miner f01234`,
	}

	// Add subcommands
	cmd.AddCommand(sectors.NewSectorsCmd())

	return cmd
}
