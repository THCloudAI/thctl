package sectors

import (
	"github.com/spf13/cobra"
)

// NewSectorsCmd creates a new sectors command
func NewSectorsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sectors",
		Short: "Manage Filecoin sectors",
		Long: `Manage Filecoin sectors including listing, status check, and information retrieval.

Examples:
  # List all sectors
  thctl fil sectors list --miner f01234

  # Check sector status
  thctl fil sectors status --miner f01234 --sector 1

  # Get sector information
  thctl fil sectors info --miner f01234 --sector 1

  # Calculate sector penalty
  thctl fil sectors penalty --miner f01234 --sector 1

  # Query vested funds
  thctl fil sectors vested --miner f01234`,
	}

	// Add subcommands
	cmd.AddCommand(
		newListCmd(),
		newStatusCmd(),
		newInfoCmd(),
		newPenaltyCmd(),
		newVestedCmd(),
	)

	return cmd
}
