package sectors

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/THCloudAI/thctl/internal/lotus"
	"github.com/THCloudAI/thctl/pkg/output"
)

// ListResult represents the result of listing sectors
type ListResult struct {
	MinerID string                   `json:"minerId"`
	Sectors []map[string]interface{} `json:"sectors"`
}

// NewListCmd creates a new list command
func NewListCmd() *cobra.Command {
	var (
		minerID string
		format  string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List sectors for a miner",
		Long:  "List all sectors for a specified miner",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create Lotus client
			client, err := lotus.NewFromEnv()
			if err != nil {
				return fmt.Errorf("failed to create Lotus client: %v", err)
			}

			// Get context
			ctx := cmd.Context()

			// List sectors
			sectors, err := client.ListSectors(ctx, minerID)
			if err != nil {
				return fmt.Errorf("failed to list sectors: %v", err)
			}

			// Convert []uint64 to []map[string]interface{}
			sectorMaps := make([]map[string]interface{}, len(sectors))
			for i, sector := range sectors {
				sectorMaps[i] = map[string]interface{}{
					"sectorNumber": sector,
				}
			}

			// Create result
			result := ListResult{
				MinerID: minerID,
				Sectors: sectorMaps,
			}

			// Print output
			if err := output.Print(result, output.Format(format)); err != nil {
				return fmt.Errorf("failed to print output: %v", err)
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&minerID, "miner", "m", "", "Miner ID (required)")
	cmd.Flags().StringVarP(&format, "format", "f", "json", "Output format (json|yaml|table)")

	// Mark required flags
	cmd.MarkFlagRequired("miner")

	return cmd
}
