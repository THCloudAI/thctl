// Copyright (c) 2024 THCloud.AI
// Author: OC
// Last Updated: 2024-12-25
// Description: Main entry point for the thctl command line tool.

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/THCloudAI/thctl/cmd/thctl/commands/auth"
	"github.com/THCloudAI/thctl/cmd/thctl/commands/cos"
	"github.com/THCloudAI/thctl/cmd/thctl/commands/doctor"
	"github.com/THCloudAI/thctl/cmd/thctl/commands/fil"
	"github.com/THCloudAI/thctl/cmd/thctl/commands/oss"
	"github.com/THCloudAI/thctl/cmd/thctl/commands/s3"
	"github.com/THCloudAI/thctl/pkg/framework/output"
	"github.com/THCloudAI/thctl/pkg/version"
)

const (
	// Version is the version of thctl
	Version = "1.3.0"
)

var (
	// Global flags
	outputFormat string
	configDir    string
	apiKey       string
	showVersion  bool

	// Root command
	rootCmd = &cobra.Command{
		Use:   "thctl",
		Short: "THCloud.AI Command Line Tool",
		Long: `thctl is a command line tool for managing THCloud.AI services.
It provides commands for managing Filecoin storage providers and cloud storage services.

Authentication:
  You can authenticate using either:
  1. Interactive browser-based auth:   thctl auth
  2. API key (--api-key flag):        thctl --api-key=your-key [command]
  3. API key (environment variable):   export THC_API_KEY=your-key

Global Options:
  -o, --output     Output format (json, yaml, table)
  --config-dir     Path to config directory
  --api-key        API key for authentication
  -v, --version    Show version number
  -h, --help       Show help`,
		Version: version.Version,
		Run: func(cmd *cobra.Command, args []string) {
			if showVersion {
				fmt.Printf("thctl version %s\n", version.Version)
				os.Exit(0)
			}
			cmd.Help()
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Check for API key in flag or environment
			if apiKey == "" {
				apiKey = os.Getenv("THC_API_KEY")
			}

			// Set default config directory if not specified
			if configDir == "" {
				userConfigDir, err := os.UserConfigDir()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error getting user config directory: %v\n", err)
					os.Exit(1)
				}
				configDir = filepath.Join(userConfigDir, "thctl")
			}

			// Create config directory if it doesn't exist
			if err := os.MkdirAll(configDir, 0700); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating config directory: %v\n", err)
				os.Exit(1)
			}

			// Validate output format
			if !output.Format(outputFormat).IsValid() {
				fmt.Fprintf(os.Stderr, "Invalid output format: %s\n", outputFormat)
				os.Exit(1)
			}
		},
	}
)

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "json", "Output format (json, yaml, table)")
	rootCmd.PersistentFlags().StringVar(&configDir, "config-dir", "", "Path to config directory")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "THCloud.AI API key")
	rootCmd.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "Show version number")

	// Add commands
	rootCmd.AddCommand(
		auth.NewAuthCmd(),
		doctor.NewDoctorCmd(),
		fil.NewFilCmd(),
		cos.NewCosCmd(),
		oss.NewOssCmd(),
		s3.NewS3Cmd(),
	)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
