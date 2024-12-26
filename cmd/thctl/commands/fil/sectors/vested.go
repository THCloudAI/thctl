package sectors

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/THCloudAI/thctl/internal/lotus"
	"github.com/THCloudAI/thctl/pkg/framework/config"
	"github.com/THCloudAI/thctl/pkg/framework/output"
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
		apiURL     string
		authToken  string
	)

	cmd := &cobra.Command{
		Use:   "vested",
		Short: "Get vested funds",
		Long: `Get the total vested funds for a miner.

Example:
  thctl fil sectors vested --miner f01234`,
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

			// Get vested funds
			ctx := cmd.Context()
			vested, err := client.GetVestedFunds(ctx, minerID)
			if err != nil {
				return fmt.Errorf("failed to get vested funds: %v", err)
			}

			// Create result
			result := VestedResult{
				MinerID: minerID,
				Vested:  vested.AvailableBalance,
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
	cmd.Flags().StringVar(&apiURL, "api-url", "", "Lotus API URL (overrides config)")
	cmd.Flags().StringVar(&authToken, "auth-token", "", "Lotus API token (overrides config)")

	// Mark required flags
	cmd.MarkFlagRequired("miner")

	return cmd
}
