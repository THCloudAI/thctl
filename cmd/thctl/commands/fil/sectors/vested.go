package sectors

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/THCloudAI/thctl/internal/lotus"
	"github.com/THCloudAI/thctl/pkg/output"
)

// VestedResult represents the vested funds result
type VestedResult struct {
	MinerID string `json:"miner_id" yaml:"miner_id"`
	Vested  string `json:"vested" yaml:"vested"`
}

// TableHeaders returns the headers for table output
func (r VestedResult) TableHeaders() []string {
	return []string{"Miner ID", "Vested"}
}

// TableRow returns the row data for table output
func (r VestedResult) TableRow() []string {
	return []string{r.MinerID, r.Vested}
}

// NewVestedCmd creates a new vested command
func NewVestedCmd() *cobra.Command {
	var (
		minerID string
		format  string
	)

	cmd := &cobra.Command{
		Use:   "vested [sector-id]",
		Short: "Get sector vested funds",
		Long:  "Get vested funds information for a specific sector",
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

			// Get sector vested info
			sectorNumber := uint64(sectorID)
			vested, err := client.GetSectorVested(cmd.Context(), minerID, sectorNumber)
			if err != nil {
				return fmt.Errorf("failed to get vested funds: %v", err)
			}

			// Print output
			if err := output.Print(vested, output.Format(format)); err != nil {
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
