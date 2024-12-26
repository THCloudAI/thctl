package fil

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/THCloudAI/thctl/cmd/thctl/commands/fil/sectors"
	"github.com/THCloudAI/thctl/internal/config"
	"github.com/THCloudAI/thctl/internal/lotus"
	"github.com/THCloudAI/thctl/internal/output"
)

// NewFilCmd creates a new fil command
func NewFilCmd() *cobra.Command {
	var (
		minerID   string
		apiURL    string
		authToken string
	)

	cmd := &cobra.Command{
		Use:   "fil",
		Short: "Filecoin commands",
		Long:  `Commands for interacting with Filecoin network.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get configuration
			cfg := config.LoadFilConfig()
			if apiURL != "" {
				cfg.Set("lotus.api_url", apiURL)
			}
			if authToken != "" {
				cfg.Set("lotus.token", authToken)
			}

			// Create Lotus client configuration
			lotusCfg := lotus.Config{
				APIURL:    cfg.GetString("lotus.api_url"),
				AuthToken: cfg.GetString("lotus.token"),
				Timeout:   cfg.GetDuration("lotus.timeout"),
			}

			// Create Lotus client
			client := lotus.New(lotusCfg)

			// Get miner info
			ctx := cmd.Context()
			info, err := client.GetMinerInfo(ctx, minerID)
			if err != nil {
				return fmt.Errorf("failed to get miner info: %v", err)
			}

			// Get miner power
			power, err := client.GetMinerPower(ctx, minerID)
			if err != nil {
				return fmt.Errorf("failed to get miner power: %v", err)
			}

			// Combine results
			result := map[string]interface{}{
				"miner_id": minerID,
				"info":     info,
				"power":    power,
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

	cmd.Flags().StringVar(&minerID, "miner", "", "Miner ID (required)")
	cmd.Flags().StringVar(&apiURL, "api-url", "", "Lotus API URL (overrides config)")
	cmd.Flags().StringVar(&authToken, "auth-token", "", "Lotus API token (overrides config)")

	cmd.MarkFlagRequired("miner")

	cmd.AddCommand(
		sectors.NewSectorsCmd(),
	)

	return cmd
}
