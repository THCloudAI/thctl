package sectors

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/THCloudAI/thctl/internal/lotus"
	"github.com/THCloudAI/thctl/pkg/output"
)

// PenaltyResult represents the sector penalty result
type PenaltyResult struct {
	MinerID  string `json:"miner_id" yaml:"miner_id"`
	Sector   uint64 `json:"sector" yaml:"sector"`
	Penalty  string `json:"penalty" yaml:"penalty"`
}

// TableHeaders returns the headers for table output
func (r PenaltyResult) TableHeaders() []string {
	return []string{"Miner ID", "Sector", "Penalty"}
}

// TableRow returns the row data for table output
func (r PenaltyResult) TableRow() []string {
	return []string{r.MinerID, fmt.Sprintf("%d", r.Sector), r.Penalty}
}

// NewPenaltyCmd creates a new penalty command
func NewPenaltyCmd() *cobra.Command {
	var (
		minerID string
		format  string
	)

	cmd := &cobra.Command{
		Use:   "penalty [sector-id]",
		Short: "Get sector penalty",
		Long:  "Get penalty information for a specific sector",
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

			// Get sector penalty
			sectorNumber := uint64(sectorID)
			penalty, err := client.GetSectorPenalty(cmd.Context(), minerID, sectorNumber)
			if err != nil {
				return fmt.Errorf("failed to get sector penalty: %v", err)
			}

			// Print output
			if err := output.Print(penalty, output.Format(format)); err != nil {
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
