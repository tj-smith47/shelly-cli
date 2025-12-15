// Package doctor provides the doctor command for system diagnostics.
package doctor

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/version"
)

// Options holds the command options.
type Options struct {
	Network bool
	Devices bool
	Full    bool
}

// NewCommand creates the doctor command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:     "doctor",
		Aliases: []string{"diagnose", "check", "health"},
		Short:   "Check system health and diagnose issues",
		Long: `Run comprehensive diagnostics on the Shelly CLI setup.

Checks include:
  - CLI version and update availability
  - Configuration file validity
  - Registered devices and their reachability
  - Network connectivity
  - Cloud authentication status
  - Firmware update availability

Use --full for all checks including device connectivity tests.`,
		Example: `  # Run basic diagnostics
  shelly doctor

  # Check network connectivity
  shelly doctor --network

  # Test all registered devices
  shelly doctor --devices

  # Full diagnostic suite
  shelly doctor --full`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Full {
				opts.Network = true
				opts.Devices = true
			}
			return run(cmd.Context(), f, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Network, "network", false, "Check network connectivity")
	cmd.Flags().BoolVar(&opts.Devices, "devices", false, "Test device reachability")
	cmd.Flags().BoolVar(&opts.Full, "full", false, "Run all diagnostic checks")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, opts *Options) error {
	ios := f.IOStreams()

	printHeader(ios)

	issues := 0
	warnings := 0

	// Check CLI version
	verIssues, verWarnings := checkCLIVersion(ios)
	issues += verIssues
	warnings += verWarnings

	// Check config
	issues += checkConfig(ios)

	// Check devices
	devIssues, devWarnings := checkDevices(ios)
	issues += devIssues
	warnings += devWarnings

	// Check cloud auth
	warnings += checkCloudAuth(ios)

	// Optional: Network checks
	if opts.Network {
		issues += checkNetwork(ctx, ios)
	}

	// Optional: Device connectivity
	if opts.Devices {
		issues += checkDeviceConnectivity(ctx, f)
	}

	printSummary(ios, issues, warnings)

	return nil
}

func printHeader(ios *iostreams.IOStreams) {
	ios.Println()
	ios.Println(theme.Title().Render("Shelly CLI Doctor"))
	ios.Println(theme.Dim().Render(strings.Repeat("━", 50)))
	ios.Println()
}

// warnStdout prints a warning message to stdout (not stderr) for consistent diagnostic output ordering.
func warnStdout(ios *iostreams.IOStreams, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	ios.Println(theme.StatusWarn().Render("⚠") + " " + msg)
}

//nolint:unparam // issues always 0 by design - version checks produce warnings, not issues
func checkCLIVersion(ios *iostreams.IOStreams) (issues, warnings int) {
	ios.Println(theme.Bold().Render("CLI Version"))

	currentVersion := version.Version
	if currentVersion == "" || currentVersion == "dev" {
		ios.Info("  Version: development build")
		ios.Println()
		return 0, 0
	}

	ios.Success("  Version: %s", currentVersion)
	ios.Info("  Go: %s", runtime.Version())
	ios.Info("  OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)

	// Check for cached update info
	home, err := os.UserHomeDir()
	if err == nil {
		cachePath := filepath.Join(home, ".config", "shelly", "cache", "latest-version")
		data, readErr := os.ReadFile(cachePath) //nolint:gosec // Path is constructed from known constants
		if readErr == nil {
			latestVersion := strings.TrimSpace(string(data))
			current := strings.TrimPrefix(currentVersion, "v")
			latest := strings.TrimPrefix(latestVersion, "v")
			if latest > current {
				warnStdout(ios, "  Update available: %s -> %s", currentVersion, latestVersion)
				ios.Info("    Run 'shelly update' to upgrade")
				ios.Println()
				return 0, 1 // Warning, not issue
			}
		}
	}

	ios.Success("  Up to date")
	ios.Println()
	return 0, 0
}

func checkConfig(ios *iostreams.IOStreams) int {
	ios.Println(theme.Bold().Render("Configuration"))
	issues := 0

	home, err := os.UserHomeDir()
	if err != nil {
		ios.Error("  Cannot determine home directory: %v", err)
		ios.Println()
		return 1
	}

	configPath := filepath.Join(home, ".config", "shelly", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		warnStdout(ios, "  Config file: not found")
		ios.Info("    Run 'shelly init' to create configuration")
		issues++
	} else {
		ios.Success("  Config file: %s", configPath)

		cfg := config.Get()
		if cfg.Output == "" {
			cfg.Output = "table"
		}
		ios.Info("    Output format: %s", cfg.Output)
		ios.Info("    Theme: %s", cfg.GetThemeConfig().Name)
		ios.Info("    API mode: %s", cfg.APIMode)
	}

	ios.Println()
	return issues
}

func checkDevices(ios *iostreams.IOStreams) (issues, warnings int) {
	ios.Println(theme.Bold().Render("Registered Devices"))

	devices := config.ListDevices()
	if len(devices) == 0 {
		warnStdout(ios, "  No devices registered")
		ios.Info("    Run 'shelly discover mdns --register' to find devices")
		ios.Println()
		return 0, 1
	}

	ios.Success("  %d device(s) registered", len(devices))

	// Check for potential issues
	for name, d := range devices {
		if d.Address == "" {
			warnStdout(ios, "    %s: missing address", name)
			issues++
		} else {
			ios.Info("    %s @ %s", name, d.Address)
		}
	}

	ios.Println()
	return issues, 0
}

func checkCloudAuth(ios *iostreams.IOStreams) int {
	ios.Println(theme.Bold().Render("Cloud Authentication"))

	cfg := config.Get()
	if cfg.Cloud.AccessToken != "" {
		ios.Success("  Cloud auth: configured")
		if cfg.Cloud.Email != "" {
			ios.Info("    Account: %s", cfg.Cloud.Email)
		}
	} else {
		ios.Info("  Cloud auth: not configured (optional)")
		ios.Info("    Run 'shelly cloud login' to enable remote control")
	}

	ios.Println()
	return 0
}

func checkNetwork(ctx context.Context, ios *iostreams.IOStreams) int {
	ios.Println(theme.Bold().Render("Network Connectivity"))
	issues := 0

	// Test internet connectivity
	client := &http.Client{Timeout: 5 * time.Second}

	endpoints := []struct {
		name string
		url  string
	}{
		{"Internet", "https://www.google.com"},
		{"Shelly Cloud", "https://shelly-api-eu.shelly.cloud"},
		{"GitHub (updates)", "https://api.github.com"},
	}

	for _, ep := range endpoints {
		req, err := http.NewRequestWithContext(ctx, http.MethodHead, ep.url, http.NoBody)
		if err != nil {
			ios.Error("  %s: request error", ep.name)
			issues++
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			warnStdout(ios, "  %s: unreachable", ep.name)
			issues++
			continue
		}
		if err := resp.Body.Close(); err != nil {
			ios.DebugErr("closing response body", err)
		}

		ios.Success("  %s: reachable", ep.name)
	}

	ios.Println()
	return issues
}

func checkDeviceConnectivity(ctx context.Context, f *cmdutil.Factory) int {
	ios := f.IOStreams()
	ios.Println(theme.Bold().Render("Device Connectivity"))

	devices := config.ListDevices()
	if len(devices) == 0 {
		ios.Info("  No devices to test")
		ios.Println()
		return 0
	}

	svc := f.ShellyService()
	issues := 0
	online := 0

	ios.StartProgress("Testing device connectivity...")

	for name, d := range devices {
		testCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		_, err := svc.DevicePing(testCtx, d.Address)
		cancel()

		if err != nil {
			warnStdout(ios, "  %s (%s): offline or unreachable", name, d.Address)
			issues++
		} else {
			online++
		}
	}

	ios.StopProgress()

	if online > 0 {
		ios.Success("  %d/%d device(s) online", online, len(devices))
	}
	if issues > 0 {
		warnStdout(ios, "  %d device(s) unreachable", issues)
	}

	ios.Println()
	return issues
}

func printSummary(ios *iostreams.IOStreams, issues, warnings int) {
	ios.Println(theme.Dim().Render(strings.Repeat("━", 50)))

	switch {
	case issues == 0 && warnings == 0:
		ios.Success("No issues found. Your Shelly CLI setup looks healthy!")
	case issues == 0:
		ios.Success("No critical issues found.")
		if warnings > 0 {
			ios.Info("%d warning(s) - see above for details", warnings)
		}
	default:
		warnStdout(ios, "%d issue(s) found - see above for details", issues)
		if warnings > 0 {
			ios.Info("%d additional warning(s)", warnings)
		}
	}

	ios.Println()
	ios.Println(fmt.Sprintf("Run %s for all diagnostics including device tests.",
		theme.Code().Render("shelly doctor --full")))
	ios.Println()
}
