package sectors

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/THCloudAI/thctl/internal/lotus"
	"github.com/THCloudAI/thctl/internal/config"
	"github.com/THCloudAI/thctl/internal/output"
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

func newVestedCmd() *cobra.Command {
	var (
		minerID    string
		sectorNum  int
		apiURL     string
		authToken  string
	)

	cmd := &cobra.Command{
		Use:   "vested",
		Short: "Get vested funds",
		Long: `Get the total vested funds for a miner.

Example:
  thctl fil sectors vested --miner f01234 --sector 1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get configuration
			cfg := config.LoadFilConfig()
			if cfg == nil {
				return fmt.Errorf("failed to get configuration")
			}

			// Create Lotus client configuration
			lotusCfg := lotus.Config{
				APIURL:    cfg.GetString("lotus.api_url"),
				AuthToken: cfg.GetString("lotus.token"),
				Timeout:   cfg.GetDuration("lotus.timeout"),
			}

			// Override with command line flags
			if apiURL != "" {
				lotusCfg.APIURL = apiURL
			}
			if authToken != "" {
				lotusCfg.AuthToken = authToken
			}

			// Create Lotus client
			client := lotus.New(lotusCfg)

			// Get vested funds
			ctx := cmd.Context()
			vested, err := client.GetSectorVested(ctx, minerID, int64(sectorNum))
			if err != nil {
				return fmt.Errorf("failed to get vested funds: %v", err)
			}

			// Get output format
			format := output.Format(cfg.GetString("output"))
			if !format.IsValid() {
				format = output.FormatTable
			}

			// Print result
			printer := output.NewPrinter(format)
			if err := printer.Print(vested); err != nil {
				return fmt.Errorf("failed to print result: %v", err)
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVar(&minerID, "miner", "", "Miner ID (required)")
	cmd.Flags().IntVar(&sectorNum, "sector", 0, "Sector number (required)")
	cmd.Flags().StringVar(&apiURL, "api-url", "", "Lotus API URL (overrides config)")
	cmd.Flags().StringVar(&authToken, "auth-token", "", "Lotus API token (overrides config)")

	// Mark required flags
	cmd.MarkFlagRequired("miner")
	cmd.MarkFlagRequired("sector")

	return cmd
}
