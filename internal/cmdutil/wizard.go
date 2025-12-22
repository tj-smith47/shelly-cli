// Package cmdutil provides command utilities.
package cmdutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/branding"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/term"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/utils"
)

// InitOptions holds options for the init wizard.
type InitOptions struct {
	// Device flags
	Devices     []string
	DevicesJSON []string

	// Discovery flags
	Discover        bool
	DiscoverTimeout time.Duration
	DiscoverModes   string
	Network         string

	// Completion flags
	Completions string
	Aliases     bool

	// Config flags
	Theme        string
	OutputFormat string
	NoColor      bool
	APIMode      string

	// Cloud flags
	CloudEmail    string
	CloudPassword string

	// Control flags
	Force bool
}

// IsNonInteractive returns true if non-interactive mode should be used.
func (o *InitOptions) IsNonInteractive() bool {
	return len(o.Devices) > 0 ||
		len(o.DevicesJSON) > 0 ||
		o.Theme != "" ||
		o.OutputFormat != "" ||
		o.APIMode != "" ||
		o.NoColor ||
		o.CloudEmail != "" ||
		o.CloudPassword != "" ||
		o.Completions != "" ||
		o.Aliases ||
		o.Discover ||
		o.Force
}

// WantsCloudSetup returns true if cloud setup should be performed.
func (o *InitOptions) WantsCloudSetup() bool {
	return o.CloudEmail != "" || o.CloudPassword != ""
}

// RunInitCheck verifies the current setup without making changes.
func RunInitCheck(f *Factory) error {
	ios := f.IOStreams()

	ios.Println("")
	ios.Println(theme.Title().Render("Shelly CLI Setup Check"))
	ios.Println(theme.Dim().Render(strings.Repeat("━", 40)))
	ios.Println("")

	issues := 0

	// Check config file
	configExists, configPath := CheckExistingConfig()
	if configExists {
		ios.Success("Config file: %s", configPath)

		cfg := config.Get()
		if err := ValidateConfig(cfg); err != nil {
			ios.Warning("Config validation: %v", err)
			issues++
		} else {
			ios.Success("Config validation: OK")
		}

		ios.Info("  Theme: %s", cfg.GetThemeConfig().Name)
		ios.Info("  Output: %s", cfg.Output)
		ios.Info("  API Mode: %s", cfg.APIMode)
	} else {
		ios.Warning("Config file: not found")
		ios.Info("  Run 'shelly init' to create configuration")
		issues++
	}

	ios.Println("")

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

	shell, err := completion.DetectShell()
	if err != nil {
		ios.Warning("Shell detection: %v", err)
	} else {
		if CheckCompletionInstalled(shell) {
			ios.Success("Shell completions: installed for %s", shell)
		} else {
			ios.Warning("Shell completions: not installed for %s", shell)
			ios.Info("  Run 'shelly completion install' to set up tab completion")
			issues++
		}
	}

	ios.Println("")

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

// RunInitWizard runs the init wizard.
func RunInitWizard(ctx context.Context, f *Factory, rootCmd *cobra.Command, opts *InitOptions) error {
	ios := f.IOStreams()

	PrintWelcome(ios)

	shouldContinue, err := CheckAndConfirmConfig(ios, opts)
	if err != nil {
		return err
	}
	if !shouldContinue {
		return nil
	}

	return runSetupSteps(ctx, f, rootCmd, opts)
}

func runSetupSteps(ctx context.Context, f *Factory, rootCmd *cobra.Command, opts *InitOptions) error {
	ios := f.IOStreams()

	if err := stepConfiguration(ios, opts); err != nil {
		return fmt.Errorf("configuration step failed: %w", err)
	}

	runFlagDevicesStep(ios, opts)
	discoveredDevices := runDiscoveryStepIfNeeded(ctx, ios, opts)
	runRegistrationStep(f, opts, discoveredDevices)
	runCompletionsStep(ios, rootCmd, opts)
	runCloudStep(ios, opts)

	PrintSummary(ios)
	return nil
}

func runFlagDevicesStep(ios *iostreams.IOStreams, opts *InitOptions) {
	if len(opts.Devices) > 0 || len(opts.DevicesJSON) > 0 {
		if err := stepFlagDevices(ios, opts); err != nil {
			ios.Warning("Device registration failed: %v", err)
		}
	}
}

func runDiscoveryStepIfNeeded(ctx context.Context, ios *iostreams.IOStreams, opts *InitOptions) []discovery.DiscoveredDevice {
	if opts.IsNonInteractive() && !opts.Discover {
		return nil
	}
	return runDiscoveryStep(ctx, ios, opts)
}

func runRegistrationStep(f *Factory, opts *InitOptions, devices []discovery.DiscoveredDevice) {
	if len(devices) > 0 {
		if err := stepRegistration(f, opts, devices); err != nil {
			f.IOStreams().Warning("Registration failed: %v", err)
		}
	}
}

func runCompletionsStep(ios *iostreams.IOStreams, rootCmd *cobra.Command, opts *InitOptions) {
	var err error
	if opts.IsNonInteractive() {
		if opts.Completions != "" {
			err = stepCompletionsNonInteractive(ios, rootCmd, opts)
		}
	} else {
		err = stepCompletions(ios, rootCmd)
	}
	if err != nil {
		ios.Warning("Completion setup failed: %v", err)
		ios.Info("You can install completions later with: shelly completion install")
	}
}

func runCloudStep(ios *iostreams.IOStreams, opts *InitOptions) {
	var err error
	if opts.IsNonInteractive() {
		if opts.WantsCloudSetup() {
			err = stepCloudNonInteractive(ios, opts)
		}
	} else {
		err = stepCloud(ios)
	}
	if err != nil {
		ios.Warning("Cloud setup failed: %v", err)
	}
}

func runDiscoveryStep(ctx context.Context, ios *iostreams.IOStreams, opts *InitOptions) []discovery.DiscoveredDevice {
	devices, err := stepDiscovery(ctx, ios, opts)
	if err != nil {
		ios.Warning("Discovery failed: %v", err)
		ios.Info("You can manually add devices later with: shelly device add <name> <address>")
		return nil
	}
	return devices
}

func stepFlagDevices(ios *iostreams.IOStreams, opts *InitOptions) error {
	ios.Println(theme.Bold().Render("Registering devices from flags..."))
	ios.Println("")

	registered, err := utils.RegisterDevicesFromFlags(opts.Devices, opts.DevicesJSON)
	if err != nil {
		if registered > 0 {
			ios.Success("Registered %d device(s)", registered)
			ios.Warning("Some devices failed: %v", err)
			ios.Println("")
			return nil
		}
		return err
	}

	if registered > 0 {
		ios.Success("Registered %d device(s)", registered)
	} else {
		ios.Info("No devices to register")
	}
	ios.Println("")
	return nil
}

// PrintWelcome prints the welcome banner.
func PrintWelcome(ios *iostreams.IOStreams) {
	ios.Println("")
	ios.Println(branding.StyledBanner())
	ios.Println("")
	ios.Println(theme.Title().Render("Welcome to Shelly CLI!"))
	ios.Println("")
	ios.Println("This wizard will help you set up shelly for the first time.")
	ios.Println("")
}

// CheckExistingConfig checks if a config file exists.
func CheckExistingConfig() (exists bool, path string) {
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

// CheckAndConfirmConfig checks for existing config and confirms overwrite.
func CheckAndConfirmConfig(ios *iostreams.IOStreams, opts *InitOptions) (bool, error) {
	configExists, configPath := CheckExistingConfig()
	if !configExists || opts.Force {
		return true, nil
	}

	if opts.IsNonInteractive() {
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

func stepConfiguration(ios *iostreams.IOStreams, opts *InitOptions) error {
	nonInteractive := opts.IsNonInteractive()

	if !nonInteractive {
		ios.Println(theme.Bold().Render("Step 1/4: Configuration"))
		ios.Println(theme.Dim().Render(strings.Repeat("━", 40)))
		ios.Println("")
	}

	outputFormat, err := selectOutputFormat(ios, opts)
	if err != nil {
		return err
	}

	themeName, err := selectTheme(ios, opts)
	if err != nil {
		return err
	}

	colorEnabled := !opts.NoColor
	apiMode := opts.APIMode
	if apiMode == "" {
		apiMode = "local"
	}

	theme.SetTheme(themeName)

	cfg := config.Get()
	cfg.Output = outputFormat
	cfg.Theme = config.ThemeConfig{Name: themeName}
	cfg.Color = colorEnabled
	cfg.APIMode = apiMode

	if cfg.Aliases == nil {
		cfg.Aliases = make(map[string]config.Alias)
	}
	if !nonInteractive || opts.Aliases {
		for name, alias := range config.DefaultAliases {
			cfg.Aliases[name] = alias
		}
	}

	viper.Set("output", outputFormat)
	viper.Set("theme", map[string]any{"name": themeName})
	viper.Set("color", colorEnabled)
	viper.Set("api_mode", apiMode)
	if len(cfg.Aliases) > 0 {
		viper.Set("aliases", cfg.Aliases)
	}

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

func stepDiscovery(ctx context.Context, ios *iostreams.IOStreams, opts *InitOptions) ([]discovery.DiscoveredDevice, error) {
	nonInteractive := opts.IsNonInteractive()

	if !nonInteractive {
		ios.Println(theme.Bold().Render("Step 2/4: Device Discovery"))
		ios.Println(theme.Dim().Render(strings.Repeat("━", 40)))
		ios.Println("")

		proceed, err := ios.Confirm("Discover devices on your network?", true)
		if err != nil {
			return nil, err
		}
		if !proceed {
			ios.Info("Skipping device discovery")
			ios.Println("")
			return nil, nil
		}
		ios.Println("")
	}

	// Select discovery methods
	methods := selectDiscoveryMethods(ios, opts)
	if len(methods) == 0 {
		ios.Info("No discovery methods selected")
		ios.Println("")
		return nil, nil
	}

	timeout := opts.DiscoverTimeout
	if timeout == 0 {
		timeout = 15 * time.Second
	}

	// Run discovery for each method and combine results
	var allDevices []discovery.DiscoveredDevice
	seenAddresses := make(map[string]bool)

	for _, method := range methods {
		devices, err := runDiscoveryMethod(ctx, ios, method, timeout, opts.Network)
		if err != nil {
			ios.Warning("%s discovery failed: %v", method, err)
			continue
		}

		// Deduplicate by address
		for _, d := range devices {
			addr := d.Address.String()
			if !seenAddresses[addr] {
				seenAddresses[addr] = true
				allDevices = append(allDevices, d)
			}
		}
	}

	ios.Println("")

	if len(allDevices) == 0 {
		ios.Warning("No devices found")
		ios.Info("Ensure devices are powered on and on the same network")
		ios.Info("You can try again later with: shelly discover")
		ios.Println("")
		return nil, nil
	}

	ios.Success("Found %d device(s)", len(allDevices))
	ios.Println("")

	term.DisplayDiscoveredDevices(ios, allDevices)
	ios.Println("")

	return allDevices, nil
}

// selectDiscoveryMethods selects which discovery methods to use.
func selectDiscoveryMethods(ios *iostreams.IOStreams, opts *InitOptions) []string {
	// Non-interactive: parse --discover-modes flag
	if opts.IsNonInteractive() {
		return parseDiscoverModes(opts.DiscoverModes)
	}

	// Interactive: multi-select
	options := []string{
		"HTTP (recommended, works everywhere)",
		"mDNS (fast, Gen2+ devices)",
		"CoIoT (Gen1 devices)",
		"BLE (Bluetooth, provisioning mode)",
	}
	defaults := []string{options[0]} // Default to HTTP

	selected, err := ios.MultiSelect("Select discovery methods:", options, defaults)
	if err != nil {
		ios.DebugErr("multi-select", err)
		return []string{"http"} // Fallback to http
	}

	var methods []string
	for _, s := range selected {
		switch {
		case strings.HasPrefix(s, "HTTP"):
			methods = append(methods, "http")
		case strings.HasPrefix(s, "mDNS"):
			methods = append(methods, "mdns")
		case strings.HasPrefix(s, "CoIoT"):
			methods = append(methods, "coiot")
		case strings.HasPrefix(s, "BLE"):
			methods = append(methods, "ble")
		}
	}

	if len(methods) == 0 {
		methods = []string{"http"} // Fallback
	}
	return methods
}

// parseDiscoverModes parses the --discover-modes flag value.
func parseDiscoverModes(modes string) []string {
	if modes == "" || modes == "all" {
		return []string{"http", "mdns", "coiot"}
	}

	var result []string
	for _, m := range strings.Split(modes, ",") {
		m = strings.TrimSpace(strings.ToLower(m))
		switch m {
		case "http", "scan":
			result = append(result, "http")
		case "mdns", "zeroconf", "bonjour":
			result = append(result, "mdns")
		case "coiot", "coap":
			result = append(result, "coiot")
		case "ble", "bluetooth":
			result = append(result, "ble")
		}
	}

	if len(result) == 0 {
		result = []string{"http"} // Default fallback
	}
	return result
}

// runDiscoveryMethod runs a single discovery method.
func runDiscoveryMethod(ctx context.Context, ios *iostreams.IOStreams, method string, timeout time.Duration, subnet string) ([]discovery.DiscoveredDevice, error) {
	switch method {
	case "http":
		return runHTTPDiscovery(ctx, ios, timeout, subnet)
	case "mdns":
		return runMDNSDiscovery(ios, timeout)
	case "coiot":
		return runCoIoTDiscovery(ios, timeout)
	case "ble":
		return runBLEDiscovery(ctx, ios, timeout)
	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}

func runHTTPDiscovery(ctx context.Context, ios *iostreams.IOStreams, timeout time.Duration, subnet string) ([]discovery.DiscoveredDevice, error) {
	if subnet == "" {
		var err error
		subnet, err = utils.DetectSubnet()
		if err != nil {
			return nil, fmt.Errorf("failed to detect subnet: %w", err)
		}
	}

	ios.StartProgress(fmt.Sprintf("Scanning %s (timeout: %s)...", subnet, timeout))
	defer ios.StopProgress()

	addresses := discovery.GenerateSubnetAddresses(subnet)
	if len(addresses) == 0 {
		return nil, fmt.Errorf("no addresses in subnet %s", subnet)
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	devices := discovery.ProbeAddressesWithProgress(ctx, addresses, func(_ discovery.ProbeProgress) bool {
		return ctx.Err() == nil
	})

	return devices, nil
}

func runMDNSDiscovery(ios *iostreams.IOStreams, timeout time.Duration) ([]discovery.DiscoveredDevice, error) {
	ios.StartProgress(fmt.Sprintf("Discovering via mDNS (timeout: %s)...", timeout))
	defer ios.StopProgress()

	mdnsDiscoverer := discovery.NewMDNSDiscoverer()
	defer func() {
		if err := mdnsDiscoverer.Stop(); err != nil {
			ios.DebugErr("stopping mDNS discoverer", err)
		}
	}()

	return mdnsDiscoverer.Discover(timeout)
}

func runCoIoTDiscovery(ios *iostreams.IOStreams, timeout time.Duration) ([]discovery.DiscoveredDevice, error) {
	ios.StartProgress(fmt.Sprintf("Discovering via CoIoT (timeout: %s)...", timeout))
	defer ios.StopProgress()

	coiotDiscoverer := discovery.NewCoIoTDiscoverer()
	defer func() {
		if err := coiotDiscoverer.Stop(); err != nil {
			ios.DebugErr("stopping CoIoT discoverer", err)
		}
	}()

	return coiotDiscoverer.Discover(timeout)
}

func runBLEDiscovery(ctx context.Context, ios *iostreams.IOStreams, timeout time.Duration) ([]discovery.DiscoveredDevice, error) {
	ios.StartProgress(fmt.Sprintf("Discovering via BLE (timeout: %s)...", timeout))
	defer ios.StopProgress()

	bleDiscoverer, err := discovery.NewBLEDiscoverer()
	if err != nil {
		return nil, fmt.Errorf("BLE not available: %w", err)
	}
	defer func() {
		if stopErr := bleDiscoverer.Stop(); stopErr != nil {
			ios.DebugErr("stopping BLE discoverer", stopErr)
		}
	}()

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return bleDiscoverer.DiscoverWithContext(ctx)
}

func stepRegistration(f *Factory, opts *InitOptions, devices []discovery.DiscoveredDevice) error {
	ios := f.IOStreams()
	nonInteractive := opts.IsNonInteractive()

	if !nonInteractive {
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
		defaultName := SanitizeDeviceName(d.ID)
		if d.Name != "" {
			defaultName = SanitizeDeviceName(d.Name)
		}

		var name string
		if nonInteractive {
			name = defaultName
		} else {
			promptMsg := fmt.Sprintf("Name for %s (%s):", d.ID, d.Address)
			var err error
			name, err = ios.Input(promptMsg, defaultName)
			if err != nil {
				return err
			}
		}

		if name == "" {
			continue
		}

		if f.GetDevice(name) != nil {
			ios.Info("Device %q already registered, skipping", name)
			continue
		}

		err := config.RegisterDevice(name, d.Address.String(), int(d.Generation), d.Model, d.Model, nil)
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

func stepCompletions(ios *iostreams.IOStreams, rootCmd *cobra.Command) error {
	ios.Println(theme.Bold().Render("Step 3/4: Shell Completions"))
	ios.Println(theme.Dim().Render(strings.Repeat("━", 40)))
	ios.Println("")

	shell, err := completion.DetectShell()
	if err != nil {
		ios.Warning("Could not detect shell: %v", err)
		ios.Info("You can install completions manually with: shelly completion install --shell <bash|zsh|fish|powershell>")
		ios.Println("")
		return nil
	}

	proceed, err := ios.Confirm(fmt.Sprintf("Install shell completions for %s?", shell), true)
	if err != nil {
		return err
	}
	if !proceed {
		ios.Info("Skipping shell completion setup")
		ios.Println("")
		return nil
	}

	if err := generateAndInstallCompletions(ios, rootCmd, shell); err != nil {
		return err
	}

	ios.Println("")
	return nil
}

func stepCompletionsNonInteractive(ios *iostreams.IOStreams, rootCmd *cobra.Command, opts *InitOptions) error {
	shells := strings.Split(opts.Completions, ",")
	installed := 0
	var errs []string

	for _, shell := range shells {
		shell = strings.TrimSpace(strings.ToLower(shell))
		if shell == "" {
			continue
		}

		switch shell {
		case completion.ShellBash, completion.ShellZsh, completion.ShellFish, completion.ShellPowerShell:
			// Valid
		default:
			ios.Warning("Unknown shell %q, skipping (valid: bash,zsh,fish,powershell)", shell)
			errs = append(errs, fmt.Sprintf("unknown shell %q", shell))
			continue
		}

		if err := generateAndInstallCompletions(ios, rootCmd, shell); err != nil {
			ios.Warning("Failed to install completions for %s: %v", shell, err)
			errs = append(errs, fmt.Sprintf("%s: %v", shell, err))
			continue
		}
		installed++
	}

	if installed > 0 {
		ios.Success("Installed completions for %d shell(s)", installed)
	}
	ios.Println("")

	if len(errs) > 0 && installed == 0 {
		return fmt.Errorf("all completion installs failed: %s", strings.Join(errs, "; "))
	}
	return nil
}

func stepCloud(ios *iostreams.IOStreams) error {
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

func stepCloudNonInteractive(ios *iostreams.IOStreams, opts *InitOptions) error {
	if opts.CloudEmail == "" || opts.CloudPassword == "" {
		ios.Warning("Cloud setup requires both --cloud-email and --cloud-password")
		ios.Info("You can set up cloud access later with: shelly cloud login")
		return nil
	}

	cfg := config.Get()
	cfg.Cloud.Email = opts.CloudEmail

	viper.Set("cloud.email", opts.CloudEmail)

	if err := config.Save(); err != nil {
		return fmt.Errorf("failed to save cloud config: %w", err)
	}

	ios.Success("Cloud credentials configured for %s", opts.CloudEmail)
	ios.Info("Run 'shelly cloud login' to complete authentication")
	ios.Println("")
	return nil
}

// PrintSummary prints the setup completion summary.
func PrintSummary(ios *iostreams.IOStreams) {
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

func generateAndInstallCompletions(ios *iostreams.IOStreams, rootCmd *cobra.Command, shell string) error {
	return completion.GenerateAndInstall(ios, rootCmd, shell)
}

func selectOutputFormat(ios *iostreams.IOStreams, opts *InitOptions) (string, error) {
	if opts.OutputFormat != "" {
		return opts.OutputFormat, nil
	}
	if opts.IsNonInteractive() {
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

func selectTheme(ios *iostreams.IOStreams, opts *InitOptions) (string, error) {
	if opts.Theme != "" {
		return opts.Theme, nil
	}
	if opts.IsNonInteractive() {
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

// SanitizeDeviceName sanitizes a device name for registration.
func SanitizeDeviceName(name string) string {
	result := strings.ToLower(name)
	result = strings.ReplaceAll(result, " ", "-")
	result = strings.ReplaceAll(result, "_", "-")
	var cleaned strings.Builder
	for _, r := range result {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			cleaned.WriteRune(r)
		}
	}
	return cleaned.String()
}

// CheckCompletionInstalled checks if shell completions are installed.
func CheckCompletionInstalled(shell string) bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	var completionPath string
	switch shell {
	case completion.ShellBash:
		completionPath = filepath.Join(home, ".local", "share", "bash-completion", "completions", "shelly")
	case completion.ShellZsh:
		completionPath = filepath.Join(home, ".zsh", "completions", "_shelly")
	case completion.ShellFish:
		completionPath = filepath.Join(home, ".config", "fish", "completions", "shelly.fish")
	case completion.ShellPowerShell:
		completionPath = filepath.Join(home, ".config", "powershell", "shelly.ps1")
	default:
		return false
	}

	_, err = os.Stat(completionPath)
	return err == nil
}

// ValidateConfig validates the configuration and returns any errors.
func ValidateConfig(cfg *config.Config) error {
	var errs []string

	validOutputs := map[string]bool{"table": true, "json": true, "yaml": true, "text": true, "template": true}
	if cfg.Output != "" && !validOutputs[cfg.Output] {
		errs = append(errs, fmt.Sprintf("invalid output format: %s", cfg.Output))
	}

	validAPIModes := map[string]bool{"local": true, "cloud": true, "auto": true}
	if cfg.APIMode != "" && !validAPIModes[cfg.APIMode] {
		errs = append(errs, fmt.Sprintf("invalid api_mode: %s", cfg.APIMode))
	}

	tc := cfg.GetThemeConfig()
	if tc.Name != "" {
		if _, exists := theme.GetTheme(tc.Name); !exists {
			errs = append(errs, fmt.Sprintf("unknown theme: %s", tc.Name))
		}
	}

	for name, device := range cfg.Devices {
		if device.Address == "" {
			errs = append(errs, fmt.Sprintf("device %q has no address", name))
		}
	}

	for groupName, group := range cfg.Groups {
		for _, deviceName := range group.Devices {
			if _, exists := cfg.Devices[deviceName]; !exists {
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
