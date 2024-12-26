package sectors

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/THCloudAI/thctl/internal/lotus"
	"github.com/THCloudAI/thctl/pkg/framework/config"
	"github.com/THCloudAI/thctl/pkg/framework/output"
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

			// Get sector status
			ctx := cmd.Context()
			status, err := client.GetSectorStatus(ctx, minerID, sectorNum)
			if err != nil {
				return fmt.Errorf("failed to get sector status: %v", err)
			}

			// Create result
			result := StatusResult{
				"miner_id": minerID,
				"sector":   sectorNum,
				"state":    status["state"],
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
