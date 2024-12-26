package fil

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/THCloudAI/thctl/cmd/thctl/commands/fil/miner"
	"github.com/THCloudAI/thctl/cmd/thctl/commands/fil/sectors"
	"github.com/THCloudAI/thctl/internal/lotus"
)

func NewFilCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fil",
		Short: "Commands for interacting with Filecoin network",
		Long:  `Commands for interacting with Filecoin network.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get flags
			apiURL, _ := cmd.Flags().GetString("api-url")
			authToken, _ := cmd.Flags().GetString("auth-token")

			// Create Lotus client configuration
			cfg := lotus.Config{
				APIURL:    apiURL,
				AuthToken: authToken,
			}

			// Create Lotus client
			client := lotus.New(cfg)
			if client == nil {
				return fmt.Errorf("failed to create Lotus client")
			}

			// TODO: Add default behavior when no subcommand is specified
			return nil
		},
	}

	// Add subcommands
	minerCmd := miner.NewMinerCmd()
	sectorsCmd := sectors.NewSectorsCmd()

	// Configure subcommands to not show global flags
	for _, subcmd := range []*cobra.Command{minerCmd, sectorsCmd} {
		subcmd.SetHelpTemplate(`{{.Long | trimTrailingWhitespaces}}

Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}
{{end}}
`)
	}

	cmd.AddCommand(sectorsCmd, minerCmd)

	// Add persistent flags
	cmd.PersistentFlags().String("api-url", "", "Lotus API URL (overrides config)")
	cmd.PersistentFlags().String("auth-token", "", "Lotus API token (overrides config)")
	cmd.PersistentFlags().String("miner", "", "Miner ID (required)")

	return cmd
}
