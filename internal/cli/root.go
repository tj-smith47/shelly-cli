// Package cli implements the command-line interface for shelly.
package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/cli/alias"
	"github.com/tj-smith47/shelly-cli/internal/cli/completion"
	"github.com/tj-smith47/shelly-cli/internal/cli/extension"
	"github.com/tj-smith47/shelly-cli/internal/cli/theme"
	"github.com/tj-smith47/shelly-cli/internal/cli/upgrade"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/version"
)

var (
	cfgFile   string
	output    string
	noColor   bool
	verbose   bool
	quiet     bool
	apiMode   string
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "shelly",
	Short: "A powerful CLI for managing Shelly smart home devices",
	Long: `Shelly CLI provides complete control over your Shelly smart home devices.

Features:
  - Device discovery (mDNS, BLE, CoIoT)
  - Full device control (switches, covers, lights, RGB)
  - Configuration management
  - Firmware updates
  - Script management (Gen2+)
  - TUI dashboard
  - Plugin system
  - Theme support

Get started:
  shelly discover          # Find devices on your network
  shelly device add        # Add a device to your registry
  shelly dash              # Launch the TUI dashboard

For more information, visit: https://github.com/tj-smith47/shelly-cli`,
	Version: version.Short(),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initConfig()
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Config file flag
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.config/shelly/config.yaml)")

	// Output format flag
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "table", "output format (json|yaml|table|text)")
	_ = viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))

	// No color flag
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable color output")
	_ = viper.BindPFlag("no_color", rootCmd.PersistentFlags().Lookup("no-color"))

	// Verbose flag
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	// Quiet flag
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress non-essential output")
	_ = viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))

	// API mode flag
	rootCmd.PersistentFlags().StringVar(&apiMode, "api-mode", "local", "API mode (local|cloud|auto)")
	_ = viper.BindPFlag("api_mode", rootCmd.PersistentFlags().Lookup("api-mode"))

	// Version template
	rootCmd.SetVersionTemplate(version.Long() + "\n")

	// Add subcommands
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(alias.NewCommand())
	rootCmd.AddCommand(extension.NewCommand())
	rootCmd.AddCommand(completion.NewCommand(rootCmd))
	rootCmd.AddCommand(theme.NewCommand())
	rootCmd.AddCommand(upgrade.NewCommand())
}

// newVersionCmd creates the version command.
func newVersionCmd() *cobra.Command {
	var short bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			if short {
				fmt.Println(version.Short())
			} else {
				fmt.Println(version.Long())
			}
		},
	}

	cmd.Flags().BoolVarP(&short, "short", "s", false, "print only the version number")

	return cmd
}

// initConfig reads in config file and ENV variables if set.
func initConfig() error {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		// Search config in ~/.config/shelly/ directory
		configDir := filepath.Join(home, ".config", "shelly")
		viper.AddConfigPath(configDir)
		viper.AddConfigPath(home)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")

		// Create config directory if it doesn't exist
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	// Environment variables
	viper.SetEnvPrefix("SHELLY")
	viper.AutomaticEnv()

	// Set defaults
	setDefaults()

	// Read config file (ignore if not found)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Load configuration into config package
	if _, err := config.Load(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	return nil
}

// setDefaults sets default configuration values.
func setDefaults() {
	viper.SetDefault("output", "table")
	viper.SetDefault("color", true)
	viper.SetDefault("theme", "dracula")
	viper.SetDefault("api_mode", "local")
	viper.SetDefault("verbose", false)
	viper.SetDefault("quiet", false)

	// Discovery defaults
	viper.SetDefault("discovery.timeout", "10s")
	viper.SetDefault("discovery.mdns", true)
	viper.SetDefault("discovery.ble", false)
	viper.SetDefault("discovery.coiot", true)

	// Devices registry (empty by default)
	viper.SetDefault("devices", map[string]any{})

	// Aliases (empty by default)
	viper.SetDefault("aliases", map[string]any{})
}

// GetRootCmd returns the root command for testing.
func GetRootCmd() *cobra.Command {
	return rootCmd
}
