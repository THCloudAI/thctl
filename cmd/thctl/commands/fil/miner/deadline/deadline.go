package deadline

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/THCloudAI/thctl/internal/lotus"
	"github.com/THCloudAI/thctl/pkg/framework/output"
)

// NewDeadlineCmd creates a new deadline command
func NewDeadlineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deadline [miner_id]",
		Short: "Get miner deadline information",
		Long: `Get detailed deadline information about a Filecoin storage provider (miner), including:
- Current proving period
- Current deadline
- Sectors due for proving
- Upcoming deadlines
- Proving window details`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			minerID := args[0]
			
			// Create Lotus client
			client := lotus.NewClient(lotus.Config{})
			ctx := context.Background()
			
			// Get proving deadline info
			deadline, err := client.GetMinerProvingDeadline(ctx, minerID)
			if err != nil {
				return fmt.Errorf("failed to get proving deadline: %w", err)
			}
			
			// Get deadlines info
			deadlines, err := client.GetMinerDeadlines(ctx, minerID)
			if err != nil {
				return fmt.Errorf("failed to get deadlines: %w", err)
			}
			
			// Combine information
			info := map[string]interface{}{
				"MinerID":         minerID,
				"ProvingDeadline": deadline,
				"Deadlines":       deadlines,
			}
			
			// Format output based on the selected format
			format, _ := cmd.Flags().GetString("output")
			switch format {
			case "json":
				return output.JSON(info)
			case "yaml":
				return output.YAML(info)
			default:
				fmt.Printf("Miner Deadline Information for %s:\n\n", minerID)
				
				if deadline != nil {
					fmt.Println("Current Proving Period:")
					if epoch, ok := deadline["CurrentEpoch"].(float64); ok {
						fmt.Printf("  Current Epoch: %d\n", int64(epoch))
					}
					if index, ok := deadline["Index"].(float64); ok {
						fmt.Printf("  Deadline Index: %d\n", int64(index))
					}
					// Add more deadline details here
				}
				
				fmt.Println("\nDeadline Schedule:")
				if deadlines != nil {
					// Format and display deadline schedule
					// This would need to be implemented based on the actual data structure
				}
			}
			
			return nil
		},
	}
	
	cmd.Flags().StringP("output", "o", "table", "Output format: json, yaml, or table")
	
	return cmd
}
