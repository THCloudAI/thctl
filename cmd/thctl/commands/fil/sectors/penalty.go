package sectors

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/THCloudAI/thctl/internal/lotus"
	"github.com/THCloudAI/thctl/pkg/framework/config"
	"github.com/THCloudAI/thctl/pkg/framework/output"
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
			// Get configuration
			cfg := config.Global()
			if cfg == nil {
				return fmt.Errorf("failed to get configuration")
			}

			// Create Lotus client configuration
			lotusCfg := lotus.DefaultConfig()
			
			// Override with config file values
			if err := cfg.UnmarshalKey("fil.lotus", lotusCfg); err != nil {
				return fmt.Errorf("failed to unmarshal lotus config: %v", err)
			}

			// Override with command line flags
			if apiURL != "" {
				lotusCfg.APIURL = apiURL
			}
			if authToken != "" {
				lotusCfg.AuthToken = authToken
			}

			// Create Lotus client
			client := lotus.NewClient(lotusCfg)

			// Get sector penalty
			ctx := cmd.Context()
			penalty, err := client.GetSectorPenalty(ctx, minerID, sectorNum)
			if err != nil {
				return fmt.Errorf("failed to get sector penalty: %v", err)
			}

			// Create result
			result := PenaltyResult{
				MinerID: minerID,
				Sector:  sectorNum,
				Penalty: penalty.InitialPledge,
			}

			// Get output format
			format := output.Format(cfg.GetString("output"))
			if !format.IsValid() {
				format = output.FormatTable
			}

			// Print result
			printer := output.NewPrinter(format)
			if err := printer.Print(result); err != nil {
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
