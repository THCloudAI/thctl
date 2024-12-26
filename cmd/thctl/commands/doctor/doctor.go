package doctor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/THCloudAI/thctl/internal/config"
)

// NewDoctorCmd creates a new doctor command
func NewDoctorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Run diagnostic checks",
		Long:  "Run diagnostic checks to verify configuration and environment settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := runDoctorChecks(); err != nil {
				cmd.SilenceUsage = true
				return err
			}
			return nil
		},
	}

	return cmd
}

func runDoctorChecks() error {
	fmt.Println("🔍 Running diagnostic checks...")
	fmt.Println()

	// First, check for .thctl.env file
	fmt.Println("📁 Checking configuration file...")
	envFile := ".thctl.env"
	configFound := false
	var configPath string

	// Check in current directory
	if _, err := os.Stat(envFile); err == nil {
		configPath = envFile
		fmt.Printf("✅ Found .thctl.env in current directory: %s\n", envFile)
		configFound = true
	} else {
		// Check in home directory
		home, err := os.UserHomeDir()
		if err == nil {
			homeEnvFile := filepath.Join(home, ".thctl.env")
			if _, err := os.Stat(homeEnvFile); err == nil {
				configPath = homeEnvFile
				fmt.Printf("✅ Found .thctl.env in home directory: %s\n", homeEnvFile)
				configFound = true
			}
		}
	}

	if !configFound {
		fmt.Println()
		fmt.Println("📊 Diagnostic Summary:")
		fmt.Println(" Configuration file (.thctl.env) not found!")
		fmt.Println("   Please create a .thctl.env file in either:")
		fmt.Printf("   - Current directory: %s\n", envFile)
		if home, err := os.UserHomeDir(); err == nil {
			fmt.Printf("   - Home directory: %s\n", filepath.Join(home, ".thctl.env"))
		}
		fmt.Println()
		fmt.Println("   The .thctl.env file should contain:")
		fmt.Println("   LOTUS_API_URL=http://your-lotus-node:1234/rpc/v0    # Required for fil commands")
		fmt.Println("   LOTUS_API_TOKEN=your-api-token                      # Required for fil commands")
		fmt.Println("   THC_API_KEY=your-api-key                           # Optional")
		return fmt.Errorf("configuration file not found")
	}

	fmt.Println()

	// Now check the configuration values
	fmt.Println("🔧 Checking configuration values...")
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf(" Failed to load configuration: %v\n", err)
		return fmt.Errorf("failed to load configuration from %s", configPath)
	}

	var missingLotusConfig []string

	// Check Lotus configuration (required for fil commands)
	if cfg.Lotus.APIURL == "" || cfg.Lotus.APIURL == "http://127.0.0.1:1234/rpc/v0" {
		missingLotusConfig = append(missingLotusConfig, "LOTUS_API_URL")
		fmt.Println("❌ LOTUS_API_URL is not set")
	} else {
		fmt.Printf("✅ LOTUS_API_URL: %s\n", cfg.Lotus.APIURL)
	}

	if cfg.Lotus.AuthToken == "" {
		missingLotusConfig = append(missingLotusConfig, "LOTUS_API_TOKEN")
		fmt.Println("❌ LOTUS_API_TOKEN is not set")
	} else {
		fmt.Println("✅ LOTUS_API_TOKEN is configured")
	}

	// Check optional THCloud configuration
	if cfg.THCloud.APIKey == "" {
		fmt.Println("ℹ️ THC_API_KEY is not set (optional)")
	} else {
		fmt.Println("✅ THC_API_KEY is configured")
	}

	fmt.Println()
	fmt.Println("📊 Diagnostic Summary:")
	if len(missingLotusConfig) > 0 {
		fmt.Println("⚠️  The following Lotus configuration values are missing:")
		for _, value := range missingLotusConfig {
			fmt.Printf("   - %s\n", value)
		}
		fmt.Println("   These values are required for using fil commands.")
		return fmt.Errorf("missing required Lotus configuration values")
	}

	fmt.Println("✅ All configuration checks passed!")
	return nil
}
