// Package cmd provides the root command and command wiring for the CLI.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/cmd/auth"
	"github.com/tj-smith47/shelly-cli/internal/cmd/backup"
	"github.com/tj-smith47/shelly-cli/internal/cmd/batch"
	"github.com/tj-smith47/shelly-cli/internal/cmd/cloud"
	configcmd "github.com/tj-smith47/shelly-cli/internal/cmd/config"
	"github.com/tj-smith47/shelly-cli/internal/cmd/cover"
	"github.com/tj-smith47/shelly-cli/internal/cmd/device"
	"github.com/tj-smith47/shelly-cli/internal/cmd/discover"
	"github.com/tj-smith47/shelly-cli/internal/cmd/ethernet"
	"github.com/tj-smith47/shelly-cli/internal/cmd/firmware"
	"github.com/tj-smith47/shelly-cli/internal/cmd/group"
	"github.com/tj-smith47/shelly-cli/internal/cmd/input"
	"github.com/tj-smith47/shelly-cli/internal/cmd/light"
	"github.com/tj-smith47/shelly-cli/internal/cmd/migrate"
	"github.com/tj-smith47/shelly-cli/internal/cmd/monitor"
	"github.com/tj-smith47/shelly-cli/internal/cmd/mqtt"
	"github.com/tj-smith47/shelly-cli/internal/cmd/rgb"
	"github.com/tj-smith47/shelly-cli/internal/cmd/scene"
	"github.com/tj-smith47/shelly-cli/internal/cmd/schedule"
	"github.com/tj-smith47/shelly-cli/internal/cmd/script"
	"github.com/tj-smith47/shelly-cli/internal/cmd/switchcmd"
	"github.com/tj-smith47/shelly-cli/internal/cmd/webhook"
	"github.com/tj-smith47/shelly-cli/internal/cmd/wifi"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/version"
)

var rootCmd = &cobra.Command{
	Use:   "shelly",
	Short: "CLI for controlling Shelly smart home devices",
	Long: `Shelly CLI - Control your Shelly smart home devices from the command line.

This tool provides a comprehensive interface for discovering, monitoring,
and controlling Shelly devices on your local network.`,
}

// Execute runs the root command with signal-aware context.
// The context is cancelled on SIGINT (Ctrl+C) or SIGTERM, enabling graceful
// shutdown of in-flight HTTP requests and other operations.
func Execute() {
	os.Exit(execute())
}

// execute runs the root command and returns an exit code.
// Separating this allows proper cleanup via defer before exit.
func execute() int {
	// Create a cancellable context that responds to signals
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		// Check if we were cancelled by signal
		if ctx.Err() != nil {
			// Exit quietly for signal-based cancellation
			return 130 // 128 + SIGINT (2)
		}
		return 1
	}
	return 0
}

func init() {
	// Set pre-run hook
	rootCmd.PersistentPreRunE = initializeConfig

	// Global flags
	rootCmd.PersistentFlags().StringP("output", "o", "table", "Output format (table, json, yaml, template)")
	rootCmd.PersistentFlags().String("template", "", "Go template string for output (use with -o template)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Suppress non-essential output")
	rootCmd.PersistentFlags().String("config", "", "Config file (default $HOME/.config/shelly/config.yaml)")
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable colored output")

	// Bind to viper - errors indicate programming bugs, panic is appropriate
	must(viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output")))
	must(viper.BindPFlag("template", rootCmd.PersistentFlags().Lookup("template")))
	must(viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose")))
	must(viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet")))
	must(viper.BindPFlag("no-color", rootCmd.PersistentFlags().Lookup("no-color")))

	// Add commands
	rootCmd.AddCommand(discover.NewCommand())
	rootCmd.AddCommand(device.NewCommand())
	rootCmd.AddCommand(group.NewCommand())
	rootCmd.AddCommand(batch.NewCommand())
	rootCmd.AddCommand(scene.NewCommand())
	rootCmd.AddCommand(switchcmd.NewCommand())
	rootCmd.AddCommand(cover.NewCommand())
	rootCmd.AddCommand(light.NewCommand())
	rootCmd.AddCommand(rgb.NewCommand())
	rootCmd.AddCommand(input.NewCommand())
	rootCmd.AddCommand(configcmd.NewCommand())
	rootCmd.AddCommand(wifi.NewCommand())
	rootCmd.AddCommand(ethernet.NewCommand())
	rootCmd.AddCommand(cloud.NewCommand())
	rootCmd.AddCommand(auth.NewCommand())
	rootCmd.AddCommand(mqtt.NewCommand())
	rootCmd.AddCommand(webhook.NewCommand())
	rootCmd.AddCommand(firmware.NewCommand())
	rootCmd.AddCommand(script.NewCommand())
	rootCmd.AddCommand(schedule.NewCommand())
	rootCmd.AddCommand(backup.NewCommand())
	rootCmd.AddCommand(migrate.NewCommand())
	rootCmd.AddCommand(monitor.NewCommand())
	rootCmd.AddCommand(versionCmd())
}

func initializeConfig(_ *cobra.Command, _ []string) error {
	// Load config
	configFile, err := rootCmd.Flags().GetString("config")
	if err != nil {
		return fmt.Errorf("failed to get config flag: %w", err)
	}

	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		var home string
		if home, err = os.UserHomeDir(); err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		viper.AddConfigPath(home + "/.config/shelly")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.SetEnvPrefix("SHELLY")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		var notFoundErr viper.ConfigFileNotFoundError
		if !errors.As(err, &notFoundErr) {
			return fmt.Errorf("failed to read config: %w", err)
		}
	}

	// Load config into struct
	if _, err := config.Load(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize theme
	themeName := viper.GetString("theme")
	if themeName != "" {
		theme.SetTheme(themeName)
	}

	// Handle color settings
	// Priority: --no-color flag > NO_COLOR env > SHELLY_NO_COLOR env
	if shouldDisableColor() {
		lipgloss.SetColorProfile(termenv.Ascii)
	}

	return nil
}

// shouldDisableColor checks if color output should be disabled.
// Returns true if --no-color flag is set, or NO_COLOR or SHELLY_NO_COLOR env vars are set.
func shouldDisableColor() bool {
	// Check if flag was explicitly set
	if viper.GetBool("no-color") {
		return true
	}

	// Check NO_COLOR env var (standard convention: https://no-color.org/)
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return true
	}

	// Check SHELLY_NO_COLOR env var (app-specific)
	if _, ok := os.LookupEnv("SHELLY_NO_COLOR"); ok {
		return true
	}

	return false
}

// must panics if err is not nil.
// Use for errors that indicate programming bugs, not runtime errors.
func must(err error) {
	if err != nil {
		panic(err)
	}
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("shelly version %s\n", version.Version)
			if version.Commit != "" {
				fmt.Printf("  commit: %s\n", version.Commit)
			}
			if version.Date != "" {
				fmt.Printf("  built: %s\n", version.Date)
			}
		},
	}
}
