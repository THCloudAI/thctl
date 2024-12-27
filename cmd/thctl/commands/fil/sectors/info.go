package sectors

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/THCloudAI/thctl/internal/lotus"
	"github.com/THCloudAI/thctl/pkg/output"
)

// InfoResult represents the sector info result
type InfoResult struct {
	MinerID    string   `json:"miner_id" yaml:"miner_id"`
	SectorID   uint64   `json:"sector_id" yaml:"sector_id"`
	State      string   `json:"state" yaml:"state"`
	SealedCID  string   `json:"sealed_cid" yaml:"sealed_cid"`
	DealIDs    []uint64 `json:"deal_ids" yaml:"deal_ids"`
}

// NewInfoCmd creates a new info command
func NewInfoCmd() *cobra.Command {
	var (
		minerID string
		format  string
	)

	cmd := &cobra.Command{
		Use:   "info [sector-id]",
		Short: "Get sector information",
		Long:  "Get detailed information about a specific sector",
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
			info, err := client.GetSectorInfo(cmd.Context(), minerID, sectorNumber)
			if err != nil {
				return fmt.Errorf("failed to get sector info: %v", err)
			}

			// Print output
			if err := output.Print(info, output.Format(format)); err != nil {
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
