package sectors

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/THCloudAI/thctl/internal/lotus"
	"github.com/THCloudAI/thctl/internal/config"
	"github.com/THCloudAI/thctl/internal/output"
)

// StatusResult represents the sector status result
type StatusResult map[string]interface{}

// TableHeaders returns the headers for table output
func (r StatusResult) TableHeaders() []string {
	return []string{"Miner ID", "Sector", "State"}
}

// TableRow returns the row data for table output
func (r StatusResult) TableRow() []string {
	return []string{r["miner_id"].(string), fmt.Sprintf("%d", r["sector"]), r["state"].(string)}
}

func newStatusCmd() *cobra.Command {
	var (
		minerID    string
		sectorNum  uint64
		apiURL     string
		authToken  string
	)

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Get sector status",
		Long: `Get the current status of a sector.

Example:
  thctl fil sectors status --miner f01234 --sector 1`,
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

			// Create Lotus client
			client := lotus.New(lotusCfg)

			// Get sector status
			ctx := cmd.Context()
			status, err := client.GetSectorInfo(ctx, minerID, int64(sectorNum))
			if err != nil {
				return fmt.Errorf("failed to get sector status: %v", err)
			}

			// Get output format
			format := output.Format(cfg.GetString("output"))
			if !format.IsValid() {
				format = output.FormatTable
			}

			// Print result
			printer := output.NewPrinter(format)
			if err := printer.Print(status); err != nil {
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
