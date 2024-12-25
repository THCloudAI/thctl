// Copyright (c) 2024 THCloud.AI
// Author: OC
// Last Updated: 2024-12-25
// Description: Authentication command implementation.

package auth

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"github.com/THCloudAI/thctl/internal/auth"
)

const (
	defaultAuthURL = "https://console.thcloudai.com/auth/cli"
	credentialsFile = "th-credentials.json"
)

// NewAuthCmd creates a new auth command
func NewAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with THCloud.AI",
		Long: `The command launches a browser window where you can authorize thctl to access your thcloudai account.
After granting permissions, your credentials are saved locally to a configuration file named th-credentials.json,
enabling you to manage your account's projects from the command line.

Alternative authentication methods:
1. Use the global --api-key option when running a command
2. Set the THC_API_KEY environment variable

Example:
  # Start the authentication flow
  thctl auth

  # Use with API key
  thctl --api-key=your-api-key [command]
  
  # Use with environment variable
  export THC_API_KEY=your-api-key
  thctl [command]`,
		RunE: runAuth,
	}

	return cmd
}

func runAuth(cmd *cobra.Command, args []string) error {
	fmt.Println("Opening browser for authentication...")
	
	// Launch browser for authentication
	err := browser.OpenURL(defaultAuthURL)
	if err != nil {
		fmt.Printf("Failed to open browser automatically. Please visit %s manually.\n", defaultAuthURL)
	}

	// Start local server to receive callback
	authClient := auth.NewClient()
	credentials, err := authClient.WaitForCallback()
	if err != nil {
		return fmt.Errorf("authentication failed: %v", err)
	}

	// Save credentials
	err = saveCredentials(credentials)
	if err != nil {
		return fmt.Errorf("failed to save credentials: %v", err)
	}

	fmt.Println("Authentication successful! Credentials saved.")
	return nil
}

func saveCredentials(credentials *auth.Credentials) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	thcloudDir := filepath.Join(configDir, "thctl")
	if err := os.MkdirAll(thcloudDir, 0700); err != nil {
		return err
	}

	credPath := filepath.Join(thcloudDir, credentialsFile)
	return credentials.SaveToFile(credPath)
}
