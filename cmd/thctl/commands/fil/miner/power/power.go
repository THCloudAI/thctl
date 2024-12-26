package power

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/thcloudai/thctl/internal/lotus"
	"github.com/thcloudai/thctl/pkg/framework/output"
)

// NewPowerCmd creates a new power command
func NewPowerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "power [miner_id]",
		Short: "Get miner power information",
		Long: `Get detailed power information about a Filecoin storage provider (miner), including:
- Raw byte power
- Quality adjusted power
- Network total power
- Relative power percentage
- Power growth over time`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			minerID := args[0]
			
			// Create Lotus client
			client := lotus.NewClient(lotus.Config{})
			
			// Get power info
			ctx := context.Background()
			info, err := client.GetMinerPower(ctx, minerID)
			if err != nil {
				return fmt.Errorf("failed to get miner power: %w", err)
			}
			
			// Format output based on the selected format
			format, _ := cmd.Flags().GetString("output")
			switch format {
			case "json":
				return output.JSON(info)
			case "yaml":
				return output.YAML(info)
			default:
				fmt.Printf("Miner Power Information for %s:\n\n", minerID)
				
				if rawPower, ok := info["MinerPower"].(map[string]interface{}); ok {
					fmt.Println("Raw Power:")
					fmt.Printf("  Raw Byte Power: %v\n", rawPower["RawBytePower"])
					fmt.Printf("  Quality Adjusted Power: %v\n", rawPower["QualityAdjPower"])
				}
				
				if totalPower, ok := info["TotalPower"].(map[string]interface{}); ok {
					fmt.Println("\nNetwork Total Power:")
					fmt.Printf("  Total Raw Byte Power: %v\n", totalPower["RawBytePower"])
					fmt.Printf("  Total Quality Adjusted Power: %v\n", totalPower["QualityAdjPower"])
				}
				
				// Calculate and display percentage of network power
				if minerRaw, ok := info["MinerPower"].(map[string]interface{})["RawBytePower"].(string); ok {
					if totalRaw, ok := info["TotalPower"].(map[string]interface{})["RawBytePower"].(string); ok {
						// Note: In real implementation, you'd need to properly parse these values
						fmt.Printf("\nNetwork Power Share: %.4f%%\n", calculatePowerShare(minerRaw, totalRaw))
					}
				}
			}
			
			return nil
		},
	}
	
	cmd.Flags().StringP("output", "o", "table", "Output format: json, yaml, or table")
	
	return cmd
}

// calculatePowerShare calculates the percentage of network power
// Note: This is a placeholder implementation. In real code, you'd need to properly parse the values
func calculatePowerShare(minerPower, totalPower string) float64 {
	// Implement proper calculation logic here
	return 0.0
}
