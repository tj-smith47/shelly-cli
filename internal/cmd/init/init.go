// Package init provides the init command for first-run setup.
package init

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/cmd/completion/install"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/helpers"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds the init command options.
type Options struct {
	NonInteractive  bool
	Force           bool
	Check           bool
	SkipDiscovery   bool
	SkipCompletions bool
	SkipCloud       bool
	Theme           string
	OutputFormat    string
	Network         string
}

// NewCommand creates the init command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:     "init",
		Aliases: []string{"setup", "configure"},
		Short:   "Initialize shelly CLI for first-time use",
		Long: `Initialize the shelly CLI with a guided setup wizard.

This command helps new users get started by:
  - Creating a configuration file with sensible defaults
  - Discovering Shelly devices on your network
  - Registering discovered devices with friendly names
  - Installing shell completions for tab completion
  - Optionally setting up Shelly Cloud access

The wizard can be skipped with --non-interactive (-y) flag.
Use --check to verify your current setup without making changes.`,
		Example: `  # Interactive setup wizard
  shelly init

  # Skip prompts, use defaults
  shelly init -y

  # Check current setup without changes
  shelly init --check

  # Skip device discovery
  shelly init --skip-discovery

  # Set theme during init
  shelly init --theme nord

  # Specify network for discovery
  shelly init --network 192.168.1.0/24`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Check {
				return runCheck(cmd.Context(), f)
			}
			return run(cmd.Context(), f, cmd.Root(), opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.NonInteractive, "non-interactive", "y", false, "Skip prompts, use defaults")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "Overwrite existing configuration")
	cmd.Flags().BoolVar(&opts.Check, "check", false, "Verify current setup without making changes")
	cmd.Flags().BoolVar(&opts.SkipDiscovery, "skip-discovery", false, "Skip device discovery")
	cmd.Flags().BoolVar(&opts.SkipCompletions, "skip-completions", false, "Skip shell completion setup")
	cmd.Flags().BoolVar(&opts.SkipCloud, "skip-cloud", false, "Skip cloud auth setup")
	cmd.Flags().StringVar(&opts.Theme, "theme", "", "Set theme (default: dracula)")
	cmd.Flags().StringVar(&opts.OutputFormat, "output-format", "", "Set output format (default: table)")
	cmd.Flags().StringVar(&opts.Network, "network", "", "Network subnet for discovery (e.g., 192.168.1.0/24)")

	return cmd
}

// runCheck verifies the current setup without making changes.
func runCheck(_ context.Context, f *cmdutil.Factory) error {
	ios := f.IOStreams()

	ios.Println("")
	ios.Println(theme.Title().Render("Shelly CLI Setup Check"))
	ios.Println(theme.Dim().Render(strings.Repeat("━", 40)))
	ios.Println("")

	issues := 0

	// Check config file
	configExists, configPath := checkExistingConfig()
	if configExists {
		ios.Success("Config file: %s", configPath)

		// Validate config
		cfg := config.Get()
		if err := validateConfig(cfg); err != nil {
			ios.Warning("Config validation: %v", err)
			issues++
		} else {
			ios.Success("Config validation: OK")
		}

		// Show config summary
		ios.Info("  Theme: %s", cfg.Theme)
		ios.Info("  Output: %s", cfg.Output)
		ios.Info("  API Mode: %s", cfg.APIMode)
	} else {
		ios.Warning("Config file: not found")
		ios.Info("  Run 'shelly init' to create configuration")
		issues++
	}

	ios.Println("")

	// Check devices
	devices := config.ListDevices()
	if len(devices) > 0 {
		ios.Success("Registered devices: %d", len(devices))
		for name, d := range devices {
			ios.Info("  %s @ %s (%s)", name, d.Address, d.Model)
		}
	} else {
		ios.Warning("Registered devices: none")
		ios.Info("  Run 'shelly discover mdns --register' to find devices")
		issues++
	}

	ios.Println("")

	// Check shell completions
	shell, err := install.DetectShell()
	if err != nil {
		ios.Warning("Shell detection: %v", err)
	} else {
		completionInstalled := checkCompletionInstalled(shell)
		if completionInstalled {
			ios.Success("Shell completions: installed for %s", shell)
		} else {
			ios.Warning("Shell completions: not installed for %s", shell)
			ios.Info("  Run 'shelly completion install' to set up tab completion")
			issues++
		}
	}

	ios.Println("")

	// Check cloud auth
	cfg := config.Get()
	if cfg.Cloud.AccessToken != "" {
		ios.Success("Cloud auth: configured")
	} else {
		ios.Info("Cloud auth: not configured (optional)")
		ios.Info("  Run 'shelly cloud login' to enable remote control")
	}

	ios.Println("")
	ios.Println(theme.Dim().Render(strings.Repeat("━", 40)))

	if issues == 0 {
		ios.Success("All checks passed!")
	} else {
		ios.Warning("%d issue(s) found", issues)
	}

	ios.Println("")

	return nil
}

func run(ctx context.Context, f *cmdutil.Factory, rootCmd *cobra.Command, opts *Options) error {
	ios := f.IOStreams()

	// Print welcome banner
	printWelcome(ios)

	// Check for existing config
	shouldContinue, err := checkAndConfirmConfig(ios, opts)
	if err != nil {
		return err
	}
	if !shouldContinue {
		return nil
	}

	// Run all setup steps
	return runSetupSteps(ctx, ios, rootCmd, opts)
}

func checkAndConfirmConfig(ios *iostreams.IOStreams, opts *Options) (bool, error) {
	configExists, configPath := checkExistingConfig()
	if !configExists || opts.Force {
		return true, nil
	}

	if opts.NonInteractive {
		ios.Warning("Configuration already exists at %s", configPath)
		ios.Info("Use --force to overwrite")
		return false, nil
	}

	overwrite, err := ios.Confirm("Configuration already exists. Overwrite?", false)
	if err != nil {
		return false, err
	}
	if !overwrite {
		ios.Info("Setup cancelled. Your existing configuration was preserved.")
		return false, nil
	}
	return true, nil
}

func runSetupSteps(ctx context.Context, ios *iostreams.IOStreams, rootCmd *cobra.Command, opts *Options) error {
	// Step 1: Configuration
	if err := stepConfiguration(ctx, ios, opts); err != nil {
		return fmt.Errorf("configuration step failed: %w", err)
	}

	// Step 2: Device Discovery
	discoveredDevices := runDiscoveryStep(ctx, ios, opts)

	// Step 3: Device Registration
	if len(discoveredDevices) > 0 {
		if err := stepRegistration(ctx, ios, opts, discoveredDevices); err != nil {
			ios.Warning("Registration failed: %v", err)
		}
	}

	// Step 4: Shell Completions
	if !opts.SkipCompletions {
		if err := stepCompletions(ctx, ios, rootCmd, opts); err != nil {
			ios.Warning("Completion setup failed: %v", err)
			ios.Info("You can install completions later with: shelly completion install")
		}
	}

	// Step 5: Cloud Setup (optional)
	if !opts.SkipCloud && !opts.NonInteractive {
		if err := stepCloud(ctx, ios, opts); err != nil {
			ios.Warning("Cloud setup skipped: %v", err)
		}
	}

	// Print completion summary
	printSummary(ios)

	return nil
}

func runDiscoveryStep(ctx context.Context, ios *iostreams.IOStreams, opts *Options) []discovery.DiscoveredDevice {
	if opts.SkipDiscovery {
		return nil
	}

	devices, err := stepDiscovery(ctx, ios, opts)
	if err != nil {
		ios.Warning("Discovery failed: %v", err)
		ios.Info("You can manually add devices later with: shelly device add <name> <address>")
		return nil
	}
	return devices
}

func printWelcome(ios *iostreams.IOStreams) {
	ios.Println("")
	ios.Println(theme.Title().Render("Welcome to Shelly CLI!"))
	ios.Println("")
	ios.Println("This wizard will help you set up shelly for the first time.")
	ios.Println("")
}

func checkExistingConfig() (exists bool, path string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return false, ""
	}

	path = filepath.Join(home, ".config", "shelly", "config.yaml")
	if _, err := os.Stat(path); err == nil {
		return true, path
	}
	return false, path
}

func stepConfiguration(_ context.Context, ios *iostreams.IOStreams, opts *Options) error {
	ios.Println(theme.Bold().Render("Step 1/4: Configuration"))
	ios.Println(theme.Dim().Render(strings.Repeat("━", 40)))
	ios.Println("")

	// Select output format
	outputFormat, err := selectOutputFormat(ios, opts)
	if err != nil {
		return err
	}

	// Select theme
	themeName, err := selectTheme(ios, opts)
	if err != nil {
		return err
	}

	// Apply theme immediately
	theme.SetTheme(themeName)

	// Create config
	cfg := config.Get()
	cfg.Output = outputFormat
	cfg.Theme = themeName
	cfg.Color = true
	cfg.APIMode = "local"

	// Save config
	viper.Set("output", outputFormat)
	viper.Set("theme", themeName)
	viper.Set("color", true)
	viper.Set("api_mode", "local")

	if err := config.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	configDir, err := config.Dir()
	if err != nil {
		return err
	}

	ios.Success("Created %s/config.yaml", configDir)
	ios.Println("")

	return nil
}

func stepDiscovery(_ context.Context, ios *iostreams.IOStreams, opts *Options) ([]discovery.DiscoveredDevice, error) {
	ios.Println(theme.Bold().Render("Step 2/4: Device Discovery"))
	ios.Println(theme.Dim().Render(strings.Repeat("━", 40)))
	ios.Println("")

	if !opts.NonInteractive {
		proceed, err := ios.Confirm("Discover devices on your network?", true)
		if err != nil {
			return nil, err
		}
		if !proceed {
			ios.Info("Skipping device discovery")
			ios.Println("")
			return nil, nil
		}
	}

	ios.Println("")
	ios.StartProgress("Discovering devices via mDNS (this may take 10-15 seconds)...")

	// Use mDNS discovery - fastest and most reliable for Gen2+
	mdnsDiscoverer := discovery.NewMDNSDiscoverer()
	defer func() {
		if err := mdnsDiscoverer.Stop(); err != nil {
			ios.DebugErr("stopping mDNS discoverer", err)
		}
	}()

	devices, err := mdnsDiscoverer.Discover(15 * time.Second)
	ios.StopProgress()

	if err != nil {
		return nil, fmt.Errorf("mDNS discovery failed: %w", err)
	}

	ios.Println("")

	if len(devices) == 0 {
		ios.Warning("No devices found")
		ios.Info("Ensure devices are powered on and on the same network")
		ios.Info("You can try again later with: shelly discover mdns")
		ios.Println("")
		return nil, nil
	}

	ios.Success("Found %d device(s)", len(devices))
	ios.Println("")

	// Display discovered devices
	helpers.DisplayDiscoveredDevices(devices)
	ios.Println("")

	return devices, nil
}

func stepRegistration(_ context.Context, ios *iostreams.IOStreams, opts *Options, devices []discovery.DiscoveredDevice) error {
	if !opts.NonInteractive {
		proceed, err := ios.Confirm("Register these devices with friendly names?", true)
		if err != nil {
			return err
		}
		if !proceed {
			ios.Info("Skipping device registration")
			ios.Info("You can register devices later with: shelly device add <name> <address>")
			ios.Println("")
			return nil
		}
	}

	ios.Println("")

	registered := 0
	for _, d := range devices {
		// Default name is the device ID
		defaultName := sanitizeDeviceName(d.ID)
		if d.Name != "" {
			defaultName = sanitizeDeviceName(d.Name)
		}

		var name string
		if opts.NonInteractive {
			name = defaultName
		} else {
			promptMsg := fmt.Sprintf("Name for %s (%s):", d.ID, d.Address)
			var err error
			name, err = ios.Input(promptMsg, defaultName)
			if err != nil {
				return err
			}
		}

		// Skip empty names
		if name == "" {
			continue
		}

		// Check if already registered
		if _, exists := config.GetDevice(name); exists {
			ios.Info("Device %q already registered, skipping", name)
			continue
		}

		err := config.RegisterDevice(
			name,
			d.Address.String(),
			int(d.Generation),
			d.Model,
			d.Model,
			nil,
		)
		if err != nil {
			ios.Warning("Failed to register %q: %v", name, err)
			continue
		}

		registered++
	}

	ios.Println("")
	if registered > 0 {
		ios.Success("Registered %d device(s)", registered)
	}
	ios.Println("")

	return nil
}

func stepCompletions(_ context.Context, ios *iostreams.IOStreams, rootCmd *cobra.Command, opts *Options) error {
	ios.Println(theme.Bold().Render("Step 3/4: Shell Completions"))
	ios.Println(theme.Dim().Render(strings.Repeat("━", 40)))
	ios.Println("")

	// Detect shell
	shell, err := install.DetectShell()
	if err != nil {
		return handleShellDetectionError(ios, opts, err)
	}

	if !opts.NonInteractive {
		proceed, err := ios.Confirm(fmt.Sprintf("Install shell completions for %s?", shell), true)
		if err != nil {
			return err
		}
		if !proceed {
			ios.Info("Skipping shell completion setup")
			ios.Println("")
			return nil
		}
	}

	// Generate and install completions
	if err := generateAndInstallCompletions(ios, rootCmd, shell); err != nil {
		return err
	}

	ios.Println("")
	return nil
}

func stepCloud(_ context.Context, ios *iostreams.IOStreams, _ *Options) error {
	ios.Println(theme.Bold().Render("Step 4/4: Cloud Access (Optional)"))
	ios.Println(theme.Dim().Render(strings.Repeat("━", 40)))
	ios.Println("")

	proceed, err := ios.Confirm("Set up Shelly Cloud access for remote control?", false)
	if err != nil {
		return err
	}
	if !proceed {
		ios.Info("Skipping cloud setup")
		ios.Info("You can set up cloud access later with: shelly cloud login")
		ios.Println("")
		return nil
	}

	ios.Info("Run 'shelly cloud login' to authenticate with Shelly Cloud")
	ios.Println("")

	return nil
}

func printSummary(ios *iostreams.IOStreams) {
	ios.Println(theme.Dim().Render(strings.Repeat("━", 50)))
	ios.Success("Setup complete!")
	ios.Println("")
	ios.Println(theme.Bold().Render("Quick start commands:"))
	ios.Println(theme.Code().Render("  shelly device list") + "          " + theme.Dim().Render("# List your devices"))
	ios.Println(theme.Code().Render("  shelly switch status <name>") + " " + theme.Dim().Render("# Check switch status"))
	ios.Println(theme.Code().Render("  shelly dash") + "                 " + theme.Dim().Render("# Open TUI dashboard"))
	ios.Println("")
	ios.Println("Run " + theme.Code().Render("shelly --help") + " for all commands.")
	ios.Println("")
}

// Helper functions

func handleShellDetectionError(ios *iostreams.IOStreams, opts *Options, err error) error {
	if opts.NonInteractive {
		ios.Info("Could not detect shell, skipping completions")
		ios.Println("")
		return nil
	}
	ios.Warning("Could not detect shell: %v", err)
	ios.Info("You can install completions manually with: shelly completion install --shell <bash|zsh|fish|powershell>")
	ios.Println("")
	return nil
}

func generateAndInstallCompletions(ios *iostreams.IOStreams, rootCmd *cobra.Command, shell string) error {
	var buf bytes.Buffer
	var err error

	switch shell {
	case install.ShellBash:
		err = rootCmd.GenBashCompletion(&buf)
	case install.ShellZsh:
		err = rootCmd.GenZshCompletion(&buf)
	case install.ShellFish:
		err = rootCmd.GenFishCompletion(&buf, true)
	case install.ShellPowerShell:
		err = rootCmd.GenPowerShellCompletionWithDesc(&buf)
	default:
		ios.Warning("Unsupported shell: %s", shell)
		ios.Println("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to generate completions: %w", err)
	}

	return installCompletionsForShell(ios, shell, buf.Bytes())
}

func installCompletionsForShell(ios *iostreams.IOStreams, shell string, script []byte) error {
	switch shell {
	case install.ShellBash:
		return install.Bash(ios, script)
	case install.ShellZsh:
		return install.Zsh(ios, script)
	case install.ShellFish:
		return install.Fish(ios, script)
	case install.ShellPowerShell:
		return install.PowerShell(ios, script)
	}
	return nil
}

func selectOutputFormat(ios *iostreams.IOStreams, opts *Options) (string, error) {
	if opts.OutputFormat != "" {
		return opts.OutputFormat, nil
	}
	if opts.NonInteractive {
		return "table", nil
	}

	formatOptions := []string{
		"table (default, human-readable)",
		"json (machine-readable)",
		"yaml (config files)",
	}
	selected, err := ios.Select("Output format:", formatOptions, 0)
	if err != nil {
		return "", err
	}
	return strings.Split(selected, " ")[0], nil
}

func selectTheme(ios *iostreams.IOStreams, opts *Options) (string, error) {
	if opts.Theme != "" {
		return opts.Theme, nil
	}
	if opts.NonInteractive {
		return "dracula", nil
	}

	themeOptions := []string{
		"dracula (default)",
		"nord",
		"tokyo-night",
		"gruvbox",
		"catppuccin",
		"one-dark",
		"solarized-dark",
		"[browse 280+ themes later with: shelly theme list]",
	}
	selected, err := ios.Select("Theme:", themeOptions, 0)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(selected, "[") {
		return "dracula", nil
	}
	return strings.Split(selected, " ")[0], nil
}

func sanitizeDeviceName(name string) string {
	// Replace spaces and special chars with hyphens
	result := strings.ToLower(name)
	result = strings.ReplaceAll(result, " ", "-")
	result = strings.ReplaceAll(result, "_", "-")
	// Remove any characters that aren't alphanumeric or hyphen
	var cleaned strings.Builder
	for _, r := range result {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			cleaned.WriteRune(r)
		}
	}
	return cleaned.String()
}

func checkCompletionInstalled(shell string) bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	var completionPath string
	switch shell {
	case install.ShellBash:
		completionPath = filepath.Join(home, ".local", "share", "bash-completion", "completions", "shelly")
	case install.ShellZsh:
		completionPath = filepath.Join(home, ".zsh", "completions", "_shelly")
	case install.ShellFish:
		completionPath = filepath.Join(home, ".config", "fish", "completions", "shelly.fish")
	case install.ShellPowerShell:
		completionPath = filepath.Join(home, ".config", "powershell", "shelly.ps1")
	default:
		return false
	}

	_, err = os.Stat(completionPath)
	return err == nil
}

// validateConfig validates the configuration and returns any errors.
func validateConfig(cfg *config.Config) error {
	var errs []string

	// Validate output format
	validOutputs := map[string]bool{"table": true, "json": true, "yaml": true, "text": true, "template": true}
	if cfg.Output != "" && !validOutputs[cfg.Output] {
		errs = append(errs, fmt.Sprintf("invalid output format: %s", cfg.Output))
	}

	// Validate API mode
	validAPIModes := map[string]bool{"local": true, "cloud": true, "auto": true}
	if cfg.APIMode != "" && !validAPIModes[cfg.APIMode] {
		errs = append(errs, fmt.Sprintf("invalid api_mode: %s", cfg.APIMode))
	}

	// Validate theme exists
	if cfg.Theme != "" {
		if _, exists := theme.GetTheme(cfg.Theme); !exists {
			errs = append(errs, fmt.Sprintf("unknown theme: %s", cfg.Theme))
		}
	}

	// Validate devices
	for name, device := range cfg.Devices {
		if device.Address == "" {
			errs = append(errs, fmt.Sprintf("device %q has no address", name))
		}
	}

	// Validate groups reference existing devices
	for groupName, group := range cfg.Groups {
		for _, deviceName := range group.Devices {
			if _, exists := cfg.Devices[deviceName]; !exists {
				// Allow non-registered devices (might be addresses)
				// but warn if it looks like a name
				if !strings.Contains(deviceName, ".") {
					errs = append(errs, fmt.Sprintf("group %q references unknown device: %s", groupName, deviceName))
				}
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}

	return nil
}
