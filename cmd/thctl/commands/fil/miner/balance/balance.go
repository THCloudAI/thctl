package balance

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/thcloudai/thctl/internal/lotus"
	"github.com/thcloudai/thctl/pkg/framework/output"
)

// NewBalanceCmd creates a new balance command
func NewBalanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balance [miner_id]",
		Short: "Get miner balance information",
		Long: `Get detailed balance information about a Filecoin storage provider (miner), including:
- Available balance
- Vesting balance
- Initial pledge
- Pre-commit deposits
- Total locked funds`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			minerID := args[0]
			
			// Create Lotus client
			client := lotus.NewClient(lotus.Config{})
			ctx := context.Background()
			
			// Get available balance
			available, err := client.GetMinerAvailableBalance(ctx, minerID)
			if err != nil {
				return fmt.Errorf("failed to get available balance: %w", err)
			}
			
			// Create balance info structure
			balanceInfo := map[string]interface{}{
				"MinerID":          minerID,
				"AvailableBalance": available,
			}
			
			// Format output based on the selected format
			format, _ := cmd.Flags().GetString("output")
			switch format {
			case "json":
				return output.JSON(balanceInfo)
			case "yaml":
				return output.YAML(balanceInfo)
			default:
				fmt.Printf("Miner Balance Information for %s:\n\n", minerID)
				fmt.Printf("Available Balance: %s\n", available)
				
				// You could add more balance-related information here:
				// - Initial pledge
				// - Pre-commit deposits
				// - Vesting funds
				// - Total locked funds
			}
			
			return nil
		},
	}
	
	cmd.Flags().StringP("output", "o", "table", "Output format: json, yaml, or table")
	
	return cmd
}
