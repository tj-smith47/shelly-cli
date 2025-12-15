// Package cmd provides the root command and command wiring for the CLI.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/cmd/action"
	"github.com/tj-smith47/shelly-cli/internal/cmd/alias"
	"github.com/tj-smith47/shelly-cli/internal/cmd/audit"
	"github.com/tj-smith47/shelly-cli/internal/cmd/auth"
	"github.com/tj-smith47/shelly-cli/internal/cmd/backup"
	"github.com/tj-smith47/shelly-cli/internal/cmd/batch"
	"github.com/tj-smith47/shelly-cli/internal/cmd/benchmark"
	"github.com/tj-smith47/shelly-cli/internal/cmd/bthome"
	"github.com/tj-smith47/shelly-cli/internal/cmd/cache"
	"github.com/tj-smith47/shelly-cli/internal/cmd/cloud"
	completioncmd "github.com/tj-smith47/shelly-cli/internal/cmd/completion"
	configcmd "github.com/tj-smith47/shelly-cli/internal/cmd/config"
	"github.com/tj-smith47/shelly-cli/internal/cmd/cover"
	"github.com/tj-smith47/shelly-cli/internal/cmd/dash"
	"github.com/tj-smith47/shelly-cli/internal/cmd/debug"
	"github.com/tj-smith47/shelly-cli/internal/cmd/device"
	"github.com/tj-smith47/shelly-cli/internal/cmd/diff"
	"github.com/tj-smith47/shelly-cli/internal/cmd/discover"
	"github.com/tj-smith47/shelly-cli/internal/cmd/doctor"
	"github.com/tj-smith47/shelly-cli/internal/cmd/energy"
	"github.com/tj-smith47/shelly-cli/internal/cmd/ethernet"
	exportcmd "github.com/tj-smith47/shelly-cli/internal/cmd/export"
	"github.com/tj-smith47/shelly-cli/internal/cmd/feedback"
	"github.com/tj-smith47/shelly-cli/internal/cmd/firmware"
	"github.com/tj-smith47/shelly-cli/internal/cmd/group"
	initcmd "github.com/tj-smith47/shelly-cli/internal/cmd/init"
	"github.com/tj-smith47/shelly-cli/internal/cmd/input"
	"github.com/tj-smith47/shelly-cli/internal/cmd/interactive"
	"github.com/tj-smith47/shelly-cli/internal/cmd/kvs"
	"github.com/tj-smith47/shelly-cli/internal/cmd/light"
	logcmd "github.com/tj-smith47/shelly-cli/internal/cmd/log"
	"github.com/tj-smith47/shelly-cli/internal/cmd/lora"
	"github.com/tj-smith47/shelly-cli/internal/cmd/matter"
	"github.com/tj-smith47/shelly-cli/internal/cmd/metrics"
	"github.com/tj-smith47/shelly-cli/internal/cmd/migrate"
	"github.com/tj-smith47/shelly-cli/internal/cmd/mock"
	"github.com/tj-smith47/shelly-cli/internal/cmd/monitor"
	"github.com/tj-smith47/shelly-cli/internal/cmd/mqtt"
	"github.com/tj-smith47/shelly-cli/internal/cmd/off"
	"github.com/tj-smith47/shelly-cli/internal/cmd/on"
	"github.com/tj-smith47/shelly-cli/internal/cmd/open"
	"github.com/tj-smith47/shelly-cli/internal/cmd/party"
	"github.com/tj-smith47/shelly-cli/internal/cmd/ping"
	"github.com/tj-smith47/shelly-cli/internal/cmd/plugin"
	"github.com/tj-smith47/shelly-cli/internal/cmd/power"
	"github.com/tj-smith47/shelly-cli/internal/cmd/provision"
	"github.com/tj-smith47/shelly-cli/internal/cmd/qr"
	"github.com/tj-smith47/shelly-cli/internal/cmd/rebootcmd"
	"github.com/tj-smith47/shelly-cli/internal/cmd/report"
	"github.com/tj-smith47/shelly-cli/internal/cmd/resetcmd"
	"github.com/tj-smith47/shelly-cli/internal/cmd/rgb"
	"github.com/tj-smith47/shelly-cli/internal/cmd/scene"
	"github.com/tj-smith47/shelly-cli/internal/cmd/schedule"
	"github.com/tj-smith47/shelly-cli/internal/cmd/script"
	"github.com/tj-smith47/shelly-cli/internal/cmd/sensor"
	"github.com/tj-smith47/shelly-cli/internal/cmd/shellcmd"
	"github.com/tj-smith47/shelly-cli/internal/cmd/sleep"
	"github.com/tj-smith47/shelly-cli/internal/cmd/statuscmd"
	"github.com/tj-smith47/shelly-cli/internal/cmd/switchcmd"
	"github.com/tj-smith47/shelly-cli/internal/cmd/sync"
	"github.com/tj-smith47/shelly-cli/internal/cmd/template"
	themecmd "github.com/tj-smith47/shelly-cli/internal/cmd/theme"
	"github.com/tj-smith47/shelly-cli/internal/cmd/thermostat"
	"github.com/tj-smith47/shelly-cli/internal/cmd/togglecmd"
	updatecmd "github.com/tj-smith47/shelly-cli/internal/cmd/update"
	versioncmd "github.com/tj-smith47/shelly-cli/internal/cmd/versioncmd"
	"github.com/tj-smith47/shelly-cli/internal/cmd/wait"
	"github.com/tj-smith47/shelly-cli/internal/cmd/wake"
	"github.com/tj-smith47/shelly-cli/internal/cmd/webhook"
	"github.com/tj-smith47/shelly-cli/internal/cmd/wifi"
	"github.com/tj-smith47/shelly-cli/internal/cmd/zigbee"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/version"
)

var rootCmd = &cobra.Command{
	Use:   "shelly",
	Short: "CLI for controlling Shelly smart home devices",
	Long: `Shelly CLI - Control your Shelly smart home devices from the command line.

This tool provides a comprehensive interface for discovering, monitoring,
and controlling Shelly devices on your local network.`,
	// Disable Cobra's auto-generated completion to use our own with install subcommand
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

// GetRootCmd returns the root command for documentation generation.
// The command tree is fully initialized with all subcommands.
func GetRootCmd() *cobra.Command {
	return rootCmd
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

	// Expand aliases before cobra processes the command
	expandedArgs, isShell := expandAlias(os.Args[1:])

	// Handle shell aliases by executing in shell
	if isShell {
		return executeShellAlias(expandedArgs)
	}

	// Set the expanded args for cobra to process
	rootCmd.SetArgs(expandedArgs)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		// Check if we were cancelled by signal
		if ctx.Err() != nil {
			// Exit quietly for signal-based cancellation
			return 130 // 128 + SIGINT (2)
		}
		return 1
	}

	// Show update notification if available (from cache)
	showCachedUpdateNotification()

	return 0
}

// showCachedUpdateNotification displays a cached update notification if available.
// This is non-blocking and only reads from the cache file.
func showCachedUpdateNotification() {
	// Skip if update check is disabled
	if os.Getenv("SHELLY_NO_UPDATE_CHECK") != "" {
		return
	}

	// Skip for certain commands (they handle their own update info)
	if len(os.Args) > 1 {
		cmd := os.Args[1]
		if cmd == "version" || cmd == "update" || cmd == "completion" || cmd == "help" {
			return
		}
	}

	// Get cache path
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	cachePath := home + "/.config/shelly/cache/latest-version"

	// Check if cache exists and is recent (within 24 hours)
	info, err := os.Stat(cachePath)
	if err != nil {
		return
	}

	// Cache expired - skip notification, will be refreshed next version check
	if info.ModTime().Add(24 * time.Hour).Before(time.Now()) {
		return
	}

	data, err := os.ReadFile(cachePath) //nolint:gosec // G304: cachePath is from known config directory
	if err != nil {
		return
	}

	cachedVersion := strings.TrimSpace(string(data))
	if cachedVersion == "" {
		return
	}

	currentVersion := strings.TrimPrefix(version.Version, "v")
	latestVersion := strings.TrimPrefix(cachedVersion, "v")

	if currentVersion == "dev" || currentVersion == "" {
		return
	}

	// Simple semver comparison - if latest is different and "newer"
	if latestVersion > currentVersion {
		iostreams.Warning("\nUpdate available: %s -> %s (run 'shelly update' to install)\n", version.Version, cachedVersion)
	}
}

// expandAlias checks if the first argument is an alias and expands it.
// Returns the expanded args and whether it's a shell alias.
func expandAlias(args []string) (expandedArgs []string, isShell bool) {
	if len(args) == 0 {
		return args, false
	}

	// Load config to check for aliases (config may not be loaded yet)
	cfg := config.Get()
	if cfg == nil {
		return args, false
	}

	// Check if first arg is an alias
	aliasObj := cfg.GetAlias(args[0])
	if aliasObj == nil {
		return args, false
	}

	// Expand the alias with remaining arguments
	expanded := config.ExpandAlias(*aliasObj, args[1:])

	if aliasObj.Shell {
		return []string{expanded}, true
	}

	// Split expanded command into args
	expandedArgs = strings.Fields(expanded)
	return expandedArgs, false
}

// executeShellAlias runs a shell alias command.
func executeShellAlias(args []string) int {
	if len(args) == 0 {
		return 0
	}

	// Execute via shell
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	//nolint:gosec // G204: args are from user-defined aliases in their own config
	cmd := exec.CommandContext(context.Background(), shell, "-c", args[0])
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "Error executing shell alias: %v\n", err)
		return 1
	}

	return 0
}

// Command group IDs for organized help output.
const (
	groupShortcuts       = "shortcuts"
	groupControl         = "control"
	groupManagement      = "management"
	groupConfig          = "config"
	groupMonitoring      = "monitoring"
	groupTroubleshooting = "troubleshooting"
	groupUtility         = "utility"
)

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

	// Define command groups for organized help output
	rootCmd.AddGroup(
		&cobra.Group{ID: groupShortcuts, Title: "Quick Commands:"},
		&cobra.Group{ID: groupControl, Title: "Device Control:"},
		&cobra.Group{ID: groupManagement, Title: "Device Management:"},
		&cobra.Group{ID: groupConfig, Title: "Configuration:"},
		&cobra.Group{ID: groupMonitoring, Title: "Monitoring:"},
		&cobra.Group{ID: groupTroubleshooting, Title: "Troubleshooting:"},
		&cobra.Group{ID: groupUtility, Title: "Utility:"},
	)

	// Create factory for dependency injection
	f := cmdutil.NewFactory()

	// Quick commands - shortcuts for common operations
	addCommandsToGroup(rootCmd, groupShortcuts,
		on.NewCommand(f),
		off.NewCommand(f),
		togglecmd.NewCommand(f),
		statuscmd.NewCommand(f),
		rebootcmd.NewCommand(f),
		resetcmd.NewCommand(f),
		open.NewCommand(f),
		ping.NewCommand(f),
		qr.NewCommand(f),
		sleep.NewCommand(f),
		wake.NewCommand(f),
		wait.NewCommand(f),
	)

	// Control commands - direct device control
	addCommandsToGroup(rootCmd, groupControl,
		switchcmd.NewCommand(f),
		cover.NewCommand(f),
		light.NewCommand(f),
		rgb.NewCommand(f),
		party.NewCommand(f),
		input.NewCommand(f),
		thermostat.NewCommand(f),
		batch.NewCommand(f),
		scene.NewCommand(f),
	)

	// Management commands - device and group management
	addCommandsToGroup(rootCmd, groupManagement,
		device.NewCommand(f),
		group.NewCommand(f),
		discover.NewCommand(f),
		script.NewCommand(f),
		schedule.NewCommand(f),
		backup.NewCommand(f),
		migrate.NewCommand(f),
		sync.NewCommand(f),
	)

	// Configuration commands - device and service configuration
	addCommandsToGroup(rootCmd, groupConfig,
		configcmd.NewCommand(f),
		wifi.NewCommand(f),
		ethernet.NewCommand(f),
		cloud.NewCommand(f),
		auth.NewCommand(f),
		mqtt.NewCommand(f),
		webhook.NewCommand(f),
		kvs.NewCommand(f),
		template.NewCommand(f),
		provision.NewCommand(f),
		action.NewCommand(f),
		bthome.NewCommand(f),
		zigbee.NewCommand(f),
		lora.NewCommand(f),
		matter.NewCommand(f),
	)

	// Monitoring commands - status and metrics
	addCommandsToGroup(rootCmd, groupMonitoring,
		monitor.NewCommand(f),
		energy.NewCommand(f),
		power.NewCommand(f),
		sensor.NewCommand(f),
		metrics.NewCommand(f),
		dash.NewCommand(f),
		report.NewCommand(f),
	)

	// Troubleshooting commands - diagnostics and debugging
	addCommandsToGroup(rootCmd, groupTroubleshooting,
		interactive.NewCommand(f),
		shellcmd.NewCommand(f),
		debug.NewCommand(f),
		doctor.NewCommand(f),
		audit.NewCommand(f),
		diff.NewCommand(f),
		benchmark.NewCommand(f),
		mock.NewCommand(f),
	)

	// Utility commands - CLI utilities
	addCommandsToGroup(rootCmd, groupUtility,
		initcmd.NewCommand(f),
		firmware.NewCommand(f),
		exportcmd.NewCommand(f),
		alias.NewCommand(f),
		plugin.NewCommand(f),
		themecmd.NewCommand(f),
		updatecmd.NewCommand(f),
		completioncmd.NewCommand(f),
		versioncmd.NewCommand(f),
		cache.NewCommand(f),
		logcmd.NewCommand(f),
		feedback.NewCommand(f),
	)
}

// addCommandsToGroup adds multiple commands to the root and assigns them to a group.
func addCommandsToGroup(root *cobra.Command, groupID string, cmds ...*cobra.Command) {
	for _, cmd := range cmds {
		cmd.GroupID = groupID
		root.AddCommand(cmd)
	}
}

func initializeConfig(_ *cobra.Command, _ []string) error {
	// Load config: flag > env var > default
	configFile, err := rootCmd.Flags().GetString("config")
	if err != nil {
		return fmt.Errorf("failed to get config flag: %w", err)
	}

	// Check SHELLY_CONFIG env var if flag not set
	if configFile == "" {
		configFile = os.Getenv("SHELLY_CONFIG")
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
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
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

	// Initialize theme from config (supports name, colors, and file)
	tc := config.Get().GetThemeConfig()
	if err := theme.ApplyConfig(tc.Name, tc.Colors, tc.File); err != nil {
		// Log theme error but don't fail - use default theme
		fmt.Fprintf(os.Stderr, "warning: %v, using default theme\n", err)
	}

	// Handle color settings
	// Priority: --no-color flag > NO_COLOR env > SHELLY_NO_COLOR env
	if shouldDisableColor() {
		lipgloss.Writer.Profile = colorprofile.Ascii
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
