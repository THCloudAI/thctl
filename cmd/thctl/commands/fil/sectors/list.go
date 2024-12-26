package sectors

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/THCloudAI/thctl/internal/lotus"
	"github.com/THCloudAI/thctl/internal/config"
	"github.com/THCloudAI/thctl/internal/output"
)

// ListResult represents the sector list result
type ListResult struct {
	MinerID string                   `json:"miner_id" yaml:"miner_id"`
	Sectors []map[string]interface{} `json:"sectors" yaml:"sectors"`
}

// TableHeaders returns the headers for table output
func (r ListResult) TableHeaders() []string {
	return []string{"Miner ID", "Sector ID", "State", "Sealed CID", "Deal Count"}
}

// TableRow returns the row data for table output
func (r ListResult) TableRow() []string {
	rows := make([][]string, 0, len(r.Sectors))
	for _, sector := range r.Sectors {
		dealIDs, _ := sector["deal_ids"].([]uint64)
		rows = append(rows, []string{
			r.MinerID,
			fmt.Sprintf("%d", sector["sector_id"].(uint64)),
			sector["state"].(string),
			sector["sealed_cid"].(string),
			fmt.Sprintf("%d", len(dealIDs)),
		})
	}
	return rows[0] // output.Printer will handle multiple rows
}

func newListCmd() *cobra.Command {
	var (
		minerID    string
		apiURL     string
		authToken  string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all sectors",
		Long: `List all sectors for a miner.

Example:
  thctl fil sectors list --miner f01234`,
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

			// List sectors
			ctx := cmd.Context()
			sectors, err := client.ListSectors(ctx, minerID)
			if err != nil {
				return fmt.Errorf("failed to list sectors: %v", err)
			}

			// Create result
			result := ListResult{
				MinerID: minerID,
				Sectors: sectors,
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
