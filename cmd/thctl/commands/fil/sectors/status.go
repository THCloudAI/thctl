package sectors

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/THCloudAI/thctl/internal/lotus"
	"github.com/THCloudAI/thctl/pkg/output"
)

// NewStatusCmd creates a new status command
func NewStatusCmd() *cobra.Command {
	var (
		minerID string
		format  string
	)

	cmd := &cobra.Command{
		Use:   "status [sector-id]",
		Short: "Get sector status",
		Long:  "Get the current status of a specific sector",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create Lotus client
			client, err := lotus.NewFromEnv()
			if err != nil {
				return fmt.Errorf("failed to create Lotus client: %v", err)
			}

			// Parse sector ID
			sectorID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse sector ID: %v", err)
			}

			// Get sector info
			sectorNumber := uint64(sectorID)
			status, err := client.GetSectorInfo(cmd.Context(), minerID, sectorNumber)
			if err != nil {
				return fmt.Errorf("failed to get sector status: %v", err)
			}

			// Print output
			if err := output.Print(status, output.Format(format)); err != nil {
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
