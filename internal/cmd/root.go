// Package cmd provides the root command and command wiring for the CLI.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/cmd/action"
	"github.com/tj-smith47/shelly-cli/internal/cmd/alert"
	"github.com/tj-smith47/shelly-cli/internal/cmd/alias"
	apicmd "github.com/tj-smith47/shelly-cli/internal/cmd/api"
	"github.com/tj-smith47/shelly-cli/internal/cmd/audit"
	"github.com/tj-smith47/shelly-cli/internal/cmd/auth"
	"github.com/tj-smith47/shelly-cli/internal/cmd/backup"
	"github.com/tj-smith47/shelly-cli/internal/cmd/batch"
	"github.com/tj-smith47/shelly-cli/internal/cmd/benchmark"
	"github.com/tj-smith47/shelly-cli/internal/cmd/bthome"
	"github.com/tj-smith47/shelly-cli/internal/cmd/cache"
	"github.com/tj-smith47/shelly-cli/internal/cmd/cert"
	"github.com/tj-smith47/shelly-cli/internal/cmd/cloud"
	completioncmd "github.com/tj-smith47/shelly-cli/internal/cmd/completion"
	configcmd "github.com/tj-smith47/shelly-cli/internal/cmd/config"
	"github.com/tj-smith47/shelly-cli/internal/cmd/cover"
	"github.com/tj-smith47/shelly-cli/internal/cmd/dash"
	"github.com/tj-smith47/shelly-cli/internal/cmd/debug"
	"github.com/tj-smith47/shelly-cli/internal/cmd/device"
	"github.com/tj-smith47/shelly-cli/internal/cmd/discover"
	"github.com/tj-smith47/shelly-cli/internal/cmd/doctor"
	"github.com/tj-smith47/shelly-cli/internal/cmd/energy"
	"github.com/tj-smith47/shelly-cli/internal/cmd/ethernet"
	exportcmd "github.com/tj-smith47/shelly-cli/internal/cmd/export"
	"github.com/tj-smith47/shelly-cli/internal/cmd/feedback"
	"github.com/tj-smith47/shelly-cli/internal/cmd/firmware"
	"github.com/tj-smith47/shelly-cli/internal/cmd/fleet"
	"github.com/tj-smith47/shelly-cli/internal/cmd/group"
	initcmd "github.com/tj-smith47/shelly-cli/internal/cmd/init"
	"github.com/tj-smith47/shelly-cli/internal/cmd/input"
	"github.com/tj-smith47/shelly-cli/internal/cmd/kvs"
	"github.com/tj-smith47/shelly-cli/internal/cmd/light"
	logcmd "github.com/tj-smith47/shelly-cli/internal/cmd/log"
	"github.com/tj-smith47/shelly-cli/internal/cmd/lora"
	"github.com/tj-smith47/shelly-cli/internal/cmd/matter"
	mcpcmd "github.com/tj-smith47/shelly-cli/internal/cmd/mcp"
	"github.com/tj-smith47/shelly-cli/internal/cmd/metrics"
	"github.com/tj-smith47/shelly-cli/internal/cmd/migrate"
	"github.com/tj-smith47/shelly-cli/internal/cmd/mock"
	"github.com/tj-smith47/shelly-cli/internal/cmd/modbus"
	"github.com/tj-smith47/shelly-cli/internal/cmd/monitor"
	"github.com/tj-smith47/shelly-cli/internal/cmd/mqtt"
	"github.com/tj-smith47/shelly-cli/internal/cmd/off"
	"github.com/tj-smith47/shelly-cli/internal/cmd/on"
	"github.com/tj-smith47/shelly-cli/internal/cmd/party"
	"github.com/tj-smith47/shelly-cli/internal/cmd/plugin"
	"github.com/tj-smith47/shelly-cli/internal/cmd/power"
	"github.com/tj-smith47/shelly-cli/internal/cmd/profile"
	"github.com/tj-smith47/shelly-cli/internal/cmd/provision"
	"github.com/tj-smith47/shelly-cli/internal/cmd/qr"
	"github.com/tj-smith47/shelly-cli/internal/cmd/repl"
	"github.com/tj-smith47/shelly-cli/internal/cmd/report"
	"github.com/tj-smith47/shelly-cli/internal/cmd/rgb"
	"github.com/tj-smith47/shelly-cli/internal/cmd/rgbw"
	"github.com/tj-smith47/shelly-cli/internal/cmd/scene"
	"github.com/tj-smith47/shelly-cli/internal/cmd/schedule"
	"github.com/tj-smith47/shelly-cli/internal/cmd/script"
	"github.com/tj-smith47/shelly-cli/internal/cmd/sensor"
	"github.com/tj-smith47/shelly-cli/internal/cmd/sensoraddon"
	"github.com/tj-smith47/shelly-cli/internal/cmd/shell"
	"github.com/tj-smith47/shelly-cli/internal/cmd/sleep"
	"github.com/tj-smith47/shelly-cli/internal/cmd/status"
	"github.com/tj-smith47/shelly-cli/internal/cmd/switchcmd"
	"github.com/tj-smith47/shelly-cli/internal/cmd/sync"
	"github.com/tj-smith47/shelly-cli/internal/cmd/template"
	themecmd "github.com/tj-smith47/shelly-cli/internal/cmd/theme"
	"github.com/tj-smith47/shelly-cli/internal/cmd/thermostat"
	"github.com/tj-smith47/shelly-cli/internal/cmd/toggle"
	updatecmd "github.com/tj-smith47/shelly-cli/internal/cmd/update"
	versioncmd "github.com/tj-smith47/shelly-cli/internal/cmd/version"
	"github.com/tj-smith47/shelly-cli/internal/cmd/virtual"
	"github.com/tj-smith47/shelly-cli/internal/cmd/wait"
	"github.com/tj-smith47/shelly-cli/internal/cmd/wake"
	"github.com/tj-smith47/shelly-cli/internal/cmd/webhook"
	"github.com/tj-smith47/shelly-cli/internal/cmd/wifi"
	"github.com/tj-smith47/shelly-cli/internal/cmd/zigbee"
	"github.com/tj-smith47/shelly-cli/internal/cmd/zwave"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	mockpkg "github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/telemetry"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/utils"
	"github.com/tj-smith47/shelly-cli/internal/version"
)

var (
	// factory is the shared factory for dependency injection.
	// It must be package-level so demo mode can inject mock config.
	factory *cmdutil.Factory

	rootCmd = &cobra.Command{
		Use:   "shelly",
		Short: "CLI for controlling Shelly smart home devices",
		Long: `Shelly CLI - Control your Shelly smart home devices from the command line.

This tool provides a comprehensive interface for discovering, monitoring,
and controlling Shelly devices on your local network.`,
		Example: `  # Initialize configuration
  shelly init

  # Discover and control devices
  shelly discover scan
  shelly switch on kitchen

  # Pipe output to jq for processing
  shelly device list -o json | jq '.[].name'

  # Pipe device names to batch commands
  echo -e "kitchen\nbedroom" | shelly batch on

  # Launch interactive dashboard
  shelly dash`,
		// Disable Cobra's auto-generated completion to use our own with install subcommand
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		// Silence Cobra's automatic usage/error printing - we handle it ourselves
		// This prevents printing help text on runtime errors (network failures, etc.)
		SilenceUsage:  true,
		SilenceErrors: true,
	}
)

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
	// Ensure telemetry client is closed gracefully
	defer telemetry.Close()

	// Create a cancellable context that responds to signals
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Expand aliases before cobra processes the command
	expandedArgs, isShell := config.ExpandAliasArgs(os.Args[1:])

	// Handle shell aliases by executing in shell
	if isShell {
		return config.ExecuteShellAlias(ctx, expandedArgs)
	}

	// Set the expanded args for cobra to process
	rootCmd.SetArgs(expandedArgs)

	// Track command execution for telemetry
	start := time.Now()
	err := rootCmd.ExecuteContext(ctx)
	duration := time.Since(start)

	// Get the executed command path and track (non-blocking, respects opt-in setting)
	cmdPath := telemetry.GetCommandPath(rootCmd, expandedArgs)
	success := err == nil && ctx.Err() == nil
	telemetry.Track(cmdPath, success, duration)

	if err != nil {
		// Check if we were cancelled by signal
		if ctx.Err() != nil {
			// Exit quietly for signal-based cancellation
			return 130 // 128 + SIGINT (2)
		}
		// Print error with themed coloring: [ERROR]: message
		errorLabel := theme.StatusError().Render("[ERROR]")
		fmt.Fprintf(os.Stderr, "%s: %s\n", errorLabel, err.Error())
		return 1
	}

	// Show update notification if available (from cache)
	version.ShowUpdateNotification()

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
	rootCmd.PersistentFlags().CountP("verbose", "v", "Increase verbosity (-v=info, -vv=debug, -vvv=trace)")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Suppress non-essential output")
	rootCmd.PersistentFlags().String("config", "", "Config file (default $HOME/.config/shelly/config.yaml)")
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable colored output")
	rootCmd.PersistentFlags().Bool("plain", false, "Disable borders and colors (machine-readable output)")
	rootCmd.PersistentFlags().Bool("no-headers", false, "Hide table headers in output")
	rootCmd.PersistentFlags().Bool("log-json", false, "Output logs in JSON format")
	rootCmd.PersistentFlags().String("log-categories", "", "Filter logs by category (comma-separated: network,api,device,config,auth,plugin)")
	rootCmd.PersistentFlags().Bool("refresh", false, "Bypass cache and fetch fresh data from device")
	rootCmd.PersistentFlags().Bool("offline", false, "Only read from cache, error on cache miss")

	// Bind to viper - errors indicate programming bugs, panic is appropriate
	utils.Must(viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output")))
	utils.Must(viper.BindPFlag("template", rootCmd.PersistentFlags().Lookup("template")))
	utils.Must(viper.BindPFlag("verbosity", rootCmd.PersistentFlags().Lookup("verbose")))
	utils.Must(viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet")))
	utils.Must(viper.BindPFlag("no-color", rootCmd.PersistentFlags().Lookup("no-color")))
	utils.Must(viper.BindPFlag("plain", rootCmd.PersistentFlags().Lookup("plain")))
	utils.Must(viper.BindPFlag("no-headers", rootCmd.PersistentFlags().Lookup("no-headers")))
	utils.Must(viper.BindPFlag("log.json", rootCmd.PersistentFlags().Lookup("log-json")))
	utils.Must(viper.BindPFlag("log.categories", rootCmd.PersistentFlags().Lookup("log-categories")))
	utils.Must(viper.BindPFlag("refresh", rootCmd.PersistentFlags().Lookup("refresh")))
	utils.Must(viper.BindPFlag("offline", rootCmd.PersistentFlags().Lookup("offline")))

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
	factory = cmdutil.NewFactory()

	// Inject mock config if demo mode is enabled
	if mockpkg.IsDemoMode() {
		demo, err := mockpkg.Start()
		if err != nil {
			fmt.Fprintf(os.Stderr, "demo mode initialization failed: %v\n", err)
		} else {
			demo.InjectIntoFactory(factory)
		}
	}

	// Quick commands - shortcuts for common operations
	cmdutil.AddCommandsToGroup(rootCmd, groupShortcuts,
		on.NewCommand(factory),
		off.NewCommand(factory),
		toggle.NewCommand(factory),
		status.NewCommand(factory),
		qr.NewCommand(factory),
		sleep.NewCommand(factory),
		wake.NewCommand(factory),
		wait.NewCommand(factory),
	)

	// Control commands - direct device control
	cmdutil.AddCommandsToGroup(rootCmd, groupControl,
		switchcmd.NewCommand(factory),
		cover.NewCommand(factory),
		light.NewCommand(factory),
		rgb.NewCommand(factory),
		rgbw.NewCommand(factory),
		party.NewCommand(factory),
		input.NewCommand(factory),
		thermostat.NewCommand(factory),
		batch.NewCommand(factory),
		scene.NewCommand(factory),
	)

	// Management commands - device and group management
	cmdutil.AddCommandsToGroup(rootCmd, groupManagement,
		device.NewCommand(factory),
		group.NewCommand(factory),
		discover.NewCommand(factory),
		script.NewCommand(factory),
		schedule.NewCommand(factory),
		backup.NewCommand(factory),
		migrate.NewCommand(factory),
		sync.NewCommand(factory),
		fleet.NewCommand(factory),
	)

	// Configuration commands - device and service configuration
	cmdutil.AddCommandsToGroup(rootCmd, groupConfig,
		action.NewCommand(factory),
		auth.NewCommand(factory),
		bthome.NewCommand(factory),
		cert.NewCommand(factory),
		cloud.NewCommand(factory),
		configcmd.NewCommand(factory),
		ethernet.NewCommand(factory),
		kvs.NewCommand(factory),
		lora.NewCommand(factory),
		matter.NewCommand(factory),
		modbus.NewCommand(factory),
		mqtt.NewCommand(factory),
		provision.NewCommand(factory),
		sensoraddon.NewCommand(factory),
		template.NewCommand(factory),
		virtual.NewCommand(factory),
		webhook.NewCommand(factory),
		wifi.NewCommand(factory),
		zigbee.NewCommand(factory),
		zwave.NewCommand(factory),
	)

	// Monitoring commands - status and metrics
	cmdutil.AddCommandsToGroup(rootCmd, groupMonitoring,
		monitor.NewCommand(factory),
		alert.NewCommand(factory),
		energy.NewCommand(factory),
		power.NewCommand(factory),
		sensor.NewCommand(factory),
		metrics.NewCommand(factory),
		dash.NewCommand(factory),
		report.NewCommand(factory),
	)

	// Troubleshooting commands - diagnostics and debugging
	cmdutil.AddCommandsToGroup(rootCmd, groupTroubleshooting,
		audit.NewCommand(factory),
		benchmark.NewCommand(factory),
		debug.NewCommand(factory),
		doctor.NewCommand(factory),
		repl.NewCommand(factory),
		mock.NewCommand(factory),
		shell.NewCommand(factory),
	)

	// Utility commands - CLI utilities
	cmdutil.AddCommandsToGroup(rootCmd, groupUtility,
		alias.NewCommand(factory),
		apicmd.NewCommand(factory),
		cache.NewCommand(factory),
		completioncmd.NewCommand(factory),
		exportcmd.NewCommand(factory),
		feedback.NewCommand(factory),
		firmware.NewCommand(factory),
		initcmd.NewCommand(factory),
		logcmd.NewCommand(factory),
		mcpcmd.NewCommand(factory),
		plugin.NewCommand(factory),
		profile.NewCommand(factory),
		themecmd.NewCommand(factory),
		updatecmd.NewCommand(factory),
		versioncmd.NewCommand(factory),
	)
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

	// Initialize theme from config (supports name, colors, semantic, and file)
	tc := config.Get().GetThemeConfig()
	if tc.File != "" {
		// Load theme from external file
		expanded := theme.ExpandPath(tc.File)
		data, err := afero.ReadFile(config.Fs(), expanded)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to read theme file: %v, using default theme\n", err)
		} else if err := theme.ApplyThemeFromData(data, tc.Semantic); err != nil {
			fmt.Fprintf(os.Stderr, "warning: %v, using default theme\n", err)
		}
	} else if err := theme.ApplyConfig(tc.Name, tc.Colors, tc.Semantic); err != nil {
		// Log theme error but don't fail - use default theme
		fmt.Fprintf(os.Stderr, "warning: %v, using default theme\n", err)
	}

	// Handle color settings
	// Priority: --no-color flag > NO_COLOR env > SHELLY_NO_COLOR env
	if iostreams.IsColorDisabled() {
		lipgloss.Writer.Profile = colorprofile.Ascii
	}

	// Configure structured logging based on verbosity and log settings
	iostreams.ConfigureLogger()

	// Enable telemetry if configured
	if config.Get().Telemetry {
		telemetry.SetEnabled(true)
	}

	return nil
}
