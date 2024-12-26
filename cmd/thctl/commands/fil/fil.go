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

			// If no subcommand is specified, show help
			return cmd.Help()
		},
	}

	// Add subcommands
	minerCmd := miner.NewMinerCmd()
	sectorsCmd := sectors.NewSectorsCmd()

	// Set custom help template for all commands to not show global flags
	helpTemplate := `{{.Long | trimTrailingWhitespaces}}

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
`

	// Apply template to fil command and all subcommands
	cmd.SetHelpTemplate(helpTemplate)
	for _, subcmd := range []*cobra.Command{minerCmd, sectorsCmd} {
		subcmd.SetHelpTemplate(helpTemplate)
	}

	cmd.AddCommand(sectorsCmd, minerCmd)

	// Add persistent flags for API configuration
	cmd.PersistentFlags().String("api-url", "", "Lotus API URL (overrides config)")
	cmd.PersistentFlags().String("auth-token", "", "Lotus API token (overrides config)")
	cmd.PersistentFlags().StringP("output", "o", "", "Output format: json, yaml, or table (default \"json\")")

	return cmd
}
