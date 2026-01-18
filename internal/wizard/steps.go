package wizard

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tj-smith47/shelly-go/discovery"
	"github.com/tj-smith47/shelly-go/types"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly/network"
	"github.com/tj-smith47/shelly-cli/internal/telemetry"
	"github.com/tj-smith47/shelly-cli/internal/term"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/utils"
)

const (
	defaultTheme = "dracula"
	totalSteps   = 5
)

func runSetupSteps(ctx context.Context, f *cmdutil.Factory, rootCmd *cobra.Command, opts *Options) error {
	ios := f.IOStreams()

	if err := stepConfiguration(ios, opts); err != nil {
		return fmt.Errorf("configuration step failed: %w", err)
	}

	runFlagDevicesStep(ios, opts)
	discoveredDevices := runDiscoveryStepIfNeeded(ctx, ios, opts)
	runRegistrationStep(f, opts, discoveredDevices)
	runCompletionsStep(ios, rootCmd, opts)
	runCloudStep(ctx, ios, opts)
	runTelemetryStep(ios, opts)

	PrintSummary(ios)
	return nil
}

func runFlagDevicesStep(ios *iostreams.IOStreams, opts *Options) {
	if len(opts.Devices) > 0 || len(opts.DevicesJSON) > 0 {
		if err := stepFlagDevices(ios, opts); err != nil {
			ios.Warning("Device registration failed: %v", err)
		}
	}
}

func runDiscoveryStepIfNeeded(ctx context.Context, ios *iostreams.IOStreams, opts *Options) []discovery.DiscoveredDevice {
	if opts.IsNonInteractive() && !opts.Discover {
		return nil
	}
	return runDiscoveryStep(ctx, ios, opts)
}

func runRegistrationStep(f *cmdutil.Factory, opts *Options, devices []discovery.DiscoveredDevice) {
	if len(devices) > 0 {
		if err := stepRegistration(f, opts, devices); err != nil {
			f.IOStreams().Warning("Registration failed: %v", err)
		}
	}
}

func runCompletionsStep(ios *iostreams.IOStreams, rootCmd *cobra.Command, opts *Options) {
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

func runCloudStep(ctx context.Context, ios *iostreams.IOStreams, opts *Options) {
	var err error
	if opts.IsNonInteractive() {
		if opts.WantsCloudSetup() {
			err = stepCloudNonInteractive(ctx, ios, opts)
		}
	} else {
		err = stepCloud(ctx, ios)
	}
	if err != nil {
		msg, hint := formatCloudError(err)
		ios.Warning("%s", msg)
		if hint != "" {
			ios.Info("%s", hint)
		}
	}
}

func runTelemetryStep(ios *iostreams.IOStreams, opts *Options) {
	var err error
	if opts.IsNonInteractive() {
		if opts.Telemetry {
			err = stepTelemetryNonInteractive(ios)
		}
	} else {
		err = stepTelemetry(ios)
	}
	if err != nil {
		ios.Warning("Telemetry setup failed: %v", err)
	}
}

func runDiscoveryStep(ctx context.Context, ios *iostreams.IOStreams, opts *Options) []discovery.DiscoveredDevice {
	devices, err := stepDiscovery(ctx, ios, opts)
	if err != nil {
		ios.Warning("Discovery failed: %v", err)
		ios.Info("You can manually add devices later with: shelly device add <name> <address>")
		return nil
	}
	return devices
}

func stepFlagDevices(ios *iostreams.IOStreams, opts *Options) error {
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

func stepConfiguration(ios *iostreams.IOStreams, opts *Options) error {
	nonInteractive := opts.IsNonInteractive()

	if !nonInteractive {
		ios.Println(theme.Bold().Render(fmt.Sprintf("Step 1/%d: Configuration", totalSteps)))
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

func stepDiscovery(ctx context.Context, ios *iostreams.IOStreams, opts *Options) ([]discovery.DiscoveredDevice, error) {
	nonInteractive := opts.IsNonInteractive()

	if !nonInteractive {
		ios.Println(theme.Bold().Render(fmt.Sprintf("Step 2/%d: Device Discovery", totalSteps)))
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
		timeout = defaultDiscoveryTimeout
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

func stepRegistration(f *cmdutil.Factory, opts *Options, devices []discovery.DiscoveredDevice) error {
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
	skipped := 0
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
			skipped++
			continue
		}

		// Type is the raw model code, Model is the human-readable name
		err := config.RegisterDevice(name, d.Address.String(), int(d.Generation), d.Model, types.ModelDisplayName(d.Model), nil)
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
	if skipped > 0 {
		ios.Info("Skipped %d already registered device(s)", skipped)
	}
	if registered == 0 && skipped == 0 {
		ios.Info("No devices registered")
	}
	ios.Println("")
	return nil
}

func stepCompletions(ios *iostreams.IOStreams, rootCmd *cobra.Command) error {
	ios.Println(theme.Bold().Render(fmt.Sprintf("Step 3/%d: Shell Completions", totalSteps)))
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

func stepCompletionsNonInteractive(ios *iostreams.IOStreams, rootCmd *cobra.Command, opts *Options) error {
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

func stepCloud(ctx context.Context, ios *iostreams.IOStreams) error {
	ios.Println(theme.Bold().Render(fmt.Sprintf("Step 4/%d: Cloud Access (Optional)", totalSteps)))
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

	ios.Println("")

	// Ask which auth method they prefer
	const (
		optOAuth         = "Browser OAuth (recommended, most secure)"
		optAuthKey       = "Auth Key (from Shelly mobile app)"
		optEmailPassword = "Email/Password"
	)
	authMethods := []string{optOAuth, optAuthKey, optEmailPassword}

	selection, err := ios.Select("Choose authentication method:", authMethods, 0)
	if err != nil {
		return fmt.Errorf("failed to select auth method: %w", err)
	}

	ios.Println("")

	switch selection {
	case optOAuth:
		ios.Info("Browser OAuth requires opening your web browser.")
		ios.Info("Please run: shelly cloud login")
		ios.Info("This will open Shelly Cloud in your browser for secure authentication.")
		ios.Println("")
		return nil

	case optAuthKey:
		return stepCloudAuthKey(ctx, ios)

	case optEmailPassword:
		return stepCloudEmailPassword(ctx, ios)
	}

	return nil
}

func stepCloudAuthKey(ctx context.Context, ios *iostreams.IOStreams) error {
	ios.Info("Find your auth key in the Shelly mobile app:")
	ios.Info("  User Settings → Authorization cloud key")
	ios.Println("")

	authKey, err := iostreams.Password("Authorization Key:")
	if err != nil {
		return fmt.Errorf("failed to read auth key: %w", err)
	}
	if authKey == "" {
		ios.Warning("Auth key is required")
		ios.Info("You can set up cloud access later with: shelly cloud login --key")
		ios.Println("")
		return nil
	}

	serverURL, err := ios.Input("Server URL (e.g., shelly-59-eu.shelly.cloud):", "")
	if err != nil {
		return fmt.Errorf("failed to read server URL: %w", err)
	}
	if serverURL == "" {
		ios.Warning("Server URL is required")
		ios.Info("You can set up cloud access later with: shelly cloud login --key")
		ios.Println("")
		return nil
	}

	// Validate credentials
	ios.StartProgress("Validating credentials...")
	client := network.NewCloudClientWithAuthKey(authKey, serverURL)
	_, err = client.GetAllDevices(ctx)
	ios.StopProgress()

	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Save credentials to config
	cfg := config.Get()
	cfg.Cloud.Enabled = true
	cfg.Cloud.AuthKey = authKey
	cfg.Cloud.ServerURL = serverURL
	cfg.Cloud.AccessToken = ""
	cfg.Cloud.Email = ""

	if err := config.Save(); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	ios.Success("Authenticated with auth key")
	ios.Info("Server: %s", serverURL)
	ios.Println("")
	return nil
}

func stepCloudEmailPassword(ctx context.Context, ios *iostreams.IOStreams) error {
	// Prompt for email
	email, err := ios.Input("Shelly Cloud email:", "")
	if err != nil {
		return fmt.Errorf("failed to read email: %w", err)
	}
	if email == "" {
		ios.Warning("Email is required for cloud login")
		ios.Info("You can set up cloud access later with: shelly cloud login")
		ios.Println("")
		return nil
	}

	// Prompt for password
	password, err := iostreams.Password("Shelly Cloud password:")
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	if password == "" {
		ios.Warning("Password is required for cloud login")
		ios.Info("You can set up cloud access later with: shelly cloud login")
		ios.Println("")
		return nil
	}

	// Perform the actual login
	ios.StartProgress("Authenticating with Shelly Cloud...")
	_, result, err := network.NewCloudClientWithCredentials(ctx, email, password)
	ios.StopProgress()

	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Save credentials to config
	cfg := config.Get()
	cfg.Cloud.Enabled = true
	cfg.Cloud.Email = email
	cfg.Cloud.AccessToken = result.Token
	cfg.Cloud.ServerURL = result.UserAPIURL
	cfg.Cloud.AuthKey = ""

	if err := config.Save(); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	ios.Success("Logged in as %s", email)
	ios.Println("")
	return nil
}

func stepCloudNonInteractive(ctx context.Context, ios *iostreams.IOStreams, opts *Options) error {
	if opts.CloudEmail == "" || opts.CloudPassword == "" {
		ios.Warning("Cloud setup requires both --cloud-email and --cloud-password")
		ios.Info("You can set up cloud access later with: shelly cloud login")
		return nil
	}

	// Perform the actual login
	ios.StartProgress("Authenticating with Shelly Cloud...")
	_, result, err := network.NewCloudClientWithCredentials(ctx, opts.CloudEmail, opts.CloudPassword)
	ios.StopProgress()

	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Save credentials to config
	cfg := config.Get()
	cfg.Cloud.Enabled = true
	cfg.Cloud.Email = opts.CloudEmail
	cfg.Cloud.AccessToken = result.Token
	cfg.Cloud.ServerURL = result.UserAPIURL

	if err := config.Save(); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	ios.Success("Logged in as %s", opts.CloudEmail)
	ios.Println("")
	return nil
}

func generateAndInstallCompletions(ios *iostreams.IOStreams, rootCmd *cobra.Command, shell string) error {
	return completion.GenerateAndInstall(ios, rootCmd, shell)
}

func selectOutputFormat(ios *iostreams.IOStreams, opts *Options) (string, error) {
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

func selectTheme(ios *iostreams.IOStreams, opts *Options) (string, error) {
	if opts.Theme != "" {
		return opts.Theme, nil
	}
	if opts.IsNonInteractive() {
		return defaultTheme, nil
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
		return defaultTheme, nil
	}
	return strings.Split(selected, " ")[0], nil
}

func stepTelemetry(ios *iostreams.IOStreams) error {
	ios.Println(theme.Bold().Render(fmt.Sprintf("Step %d/%d: Anonymous Usage Statistics (Optional)", totalSteps, totalSteps)))
	ios.Println(theme.Dim().Render(strings.Repeat("━", 40)))
	ios.Println("")
	ios.Println("Help improve Shelly CLI by sharing anonymous usage statistics.")
	ios.Println("")
	ios.Println(theme.Dim().Render("What we collect:"))
	ios.Println(theme.Dim().Render("  • Command names (e.g., 'device info', 'switch on')"))
	ios.Println(theme.Dim().Render("  • Success/failure status"))
	ios.Println(theme.Dim().Render("  • CLI version, OS, and architecture"))
	ios.Println("")
	ios.Println(theme.Dim().Render("What we DON'T collect:"))
	ios.Println(theme.Dim().Render("  • Device names, IP addresses, or network info"))
	ios.Println(theme.Dim().Render("  • Personal data or credentials"))
	ios.Println(theme.Dim().Render("  • Command arguments or parameters"))
	ios.Println("")

	proceed, err := ios.Confirm("Enable anonymous usage statistics?", false)
	if err != nil {
		return err
	}

	if proceed {
		if err := enableTelemetry(); err != nil {
			return err
		}
		ios.Success("Telemetry enabled")
		ios.Info("You can disable it anytime with: shelly config set telemetry false")
	} else {
		ios.Info("Telemetry disabled")
		ios.Info("You can enable it later with: shelly config set telemetry true")
	}
	ios.Println("")
	return nil
}

func stepTelemetryNonInteractive(ios *iostreams.IOStreams) error {
	if err := enableTelemetry(); err != nil {
		return err
	}
	ios.Success("Telemetry enabled")
	ios.Println("")
	return nil
}

func enableTelemetry() error {
	cfg := config.Get()
	cfg.Telemetry = true
	viper.Set("telemetry", true)
	if err := config.Save(); err != nil {
		return fmt.Errorf("failed to save telemetry setting: %w", err)
	}
	telemetry.SetEnabled(true)
	return nil
}
