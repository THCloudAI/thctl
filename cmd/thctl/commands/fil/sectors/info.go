package sectors

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/THCloudAI/thctl/internal/lotus"
	"github.com/THCloudAI/thctl/internal/config"
	"github.com/THCloudAI/thctl/internal/output"
)

// InfoResult represents the sector info result
type InfoResult struct {
	MinerID    string   `json:"miner_id" yaml:"miner_id"`
	SectorID   uint64   `json:"sector_id" yaml:"sector_id"`
	State      string   `json:"state" yaml:"state"`
	SealedCID  string   `json:"sealed_cid" yaml:"sealed_cid"`
	DealIDs    []uint64 `json:"deal_ids" yaml:"deal_ids"`
}

// TableHeaders returns the headers for table output
func (r InfoResult) TableHeaders() []string {
	return []string{"Miner ID", "Sector ID", "State", "Sealed CID", "Deal IDs"}
}

// TableRow returns the row data for table output
func (r InfoResult) TableRow() []string {
	dealIDs := ""
	for i, id := range r.DealIDs {
		if i > 0 {
			dealIDs += ","
		}
		dealIDs += fmt.Sprintf("%d", id)
	}
	return []string{
		r.MinerID,
		fmt.Sprintf("%d", r.SectorID),
		r.State,
		r.SealedCID,
		dealIDs,
	}
}

func newInfoCmd() *cobra.Command {
	var (
		minerID    string
		sectorNum  uint64
		apiURL     string
		authToken  string
	)

	cmd := &cobra.Command{
		Use:   "info",
		Short: "Get sector information",
		Long: `Get detailed information about a sector.

Example:
  thctl fil sectors info --miner f01234 --sector 1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load configuration: %v", err)
			}

			// Get miner ID from flag
			minerID := cmd.Flag("miner").Value.String()
			if minerID == "" {
				return fmt.Errorf("miner ID is required")
			}

			// Create Lotus client
			client := lotus.New(lotus.Config{
				APIURL:    cfg.Lotus.APIURL,
				AuthToken: cfg.Lotus.AuthToken,
			})

			// Get sector info
			ctx := cmd.Context()

			// Get sector ID from flag
			sectorID, err := cmd.Flags().GetInt64("sector")
			if err != nil {
				return fmt.Errorf("failed to get sector ID: %v", err)
			}

			info, err := client.GetSectorInfo(ctx, minerID, sectorID)
			if err != nil {
				return fmt.Errorf("failed to get sector info: %v", err)
			}

			// Get output format
			format := output.Format("table")
			if !format.IsValid() {
				format = output.FormatTable
			}

			// Print result
			printer := output.NewPrinter(format)
			if err := printer.Print(info); err != nil {
				return fmt.Errorf("failed to print result: %v", err)
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVar(&minerID, "miner", "", "Miner ID (required)")
	cmd.Flags().Uint64Var(&sectorNum, "sector", 0, "Sector number (required)")
	cmd.Flags().StringVar(&apiURL, "api-url", "", "Lotus API URL (overrides config)")
	cmd.Flags().StringVar(&authToken, "auth-token", "", "Lotus API token (overrides config)")

	// Mark required flags
	cmd.MarkFlagRequired("miner")
	cmd.MarkFlagRequired("sector")

	return cmd
}
