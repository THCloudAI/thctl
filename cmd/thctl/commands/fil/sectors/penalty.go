package sectors

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/THCloudAI/thctl/internal/lotus"
	"github.com/THCloudAI/thctl/internal/config"
	"github.com/THCloudAI/thctl/internal/output"
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

func newPenaltyCmd() *cobra.Command {
	var (
		minerID    string
		sectorNum  uint64
		apiURL     string
		authToken  string
	)

	cmd := &cobra.Command{
		Use:   "penalty",
		Short: "Get sector penalty",
		Long: `Get the penalty for a sector.

Example:
  thctl fil sectors penalty --miner f01234 --sector 1`,
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

			// Get sector penalty
			ctx := cmd.Context()

			// Get sector ID from flag
			sectorID, err := cmd.Flags().GetInt64("sector")
			if err != nil {
				return fmt.Errorf("failed to get sector ID: %v", err)
			}

			penalty, err := client.GetSectorPenalty(ctx, minerID, sectorID)
			if err != nil {
				return fmt.Errorf("failed to get sector penalty: %v", err)
			}

			// Get output format
			format := output.Format("table")
			if !format.IsValid() {
				format = output.FormatTable
			}

			// Print result
			printer := output.NewPrinter(format)
			if err := printer.Print(penalty); err != nil {
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
