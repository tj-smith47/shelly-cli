package term

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/version"
)

// DisplayDoctorHeader displays the doctor command header.
func DisplayDoctorHeader(ios *iostreams.IOStreams) {
	ios.Println()
	ios.Println(theme.Title().Render("Shelly CLI Doctor"))
	ios.Println(theme.Dim().Render(strings.Repeat("━", 50)))
	ios.Println()
}

// DisplayDoctorSummary displays the doctor command summary.
func DisplayDoctorSummary(ios *iostreams.IOStreams, issues, warnings int) {
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
		WarnStdout(ios, "%d issue(s) found - see above for details", issues)
		if warnings > 0 {
			ios.Info("%d additional warning(s)", warnings)
		}
	}

	ios.Println()
	ios.Println(fmt.Sprintf("Run %s for all diagnostics including device tests.",
		theme.Code().Render("shelly doctor --full")))
	ios.Println()
}

// WarnStdout prints a warning message to stdout (not stderr) for consistent diagnostic output ordering.
func WarnStdout(ios *iostreams.IOStreams, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	ios.Println(theme.StatusWarn().Render("⚠") + " " + msg)
}

// CheckCLIVersion displays CLI version diagnostics.
// The issues return value is always 0 by design - version checks produce warnings, not issues.
func CheckCLIVersion(ios *iostreams.IOStreams) (issues, warnings int) {
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
		data, readErr := afero.ReadFile(config.Fs(), cachePath)
		if readErr == nil {
			latestVersion := strings.TrimSpace(string(data))
			current := strings.TrimPrefix(currentVersion, "v")
			latest := strings.TrimPrefix(latestVersion, "v")
			if latest > current {
				WarnStdout(ios, "  Update available: %s -> %s", currentVersion, latestVersion)
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

// CheckConfig displays configuration diagnostics.
func CheckConfig(ios *iostreams.IOStreams) int {
	ios.Println(theme.Bold().Render("Configuration"))
	issues := 0

	home, err := os.UserHomeDir()
	if err != nil {
		ios.Error("  Cannot determine home directory: %v", err)
		ios.Println()
		return 1
	}

	configPath := filepath.Join(home, ".config", "shelly", "config.yaml")
	exists, existsErr := afero.Exists(config.Fs(), configPath)
	if existsErr != nil || !exists {
		WarnStdout(ios, "  Config file: not found")
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

// CheckDevices displays registered devices diagnostics.
func CheckDevices(ios *iostreams.IOStreams) (issues, warnings int) {
	ios.Println(theme.Bold().Render("Registered Devices"))

	devices := config.ListDevices()
	if len(devices) == 0 {
		WarnStdout(ios, "  No devices registered")
		ios.Info("    Run 'shelly discover mdns --register' to find devices")
		ios.Println()
		return 0, 1
	}

	ios.Success("  %d device(s) registered", len(devices))

	// Check for potential issues
	for name, d := range devices {
		if d.Address == "" {
			WarnStdout(ios, "    %s: missing address", name)
			issues++
		} else {
			ios.Info("    %s @ %s", name, d.Address)
		}
	}

	ios.Println()
	return issues, 0
}

// CheckCloudAuth displays cloud authentication diagnostics.
func CheckCloudAuth(ios *iostreams.IOStreams) int {
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

// CheckNetwork displays network connectivity diagnostics.
func CheckNetwork(ctx context.Context, ios *iostreams.IOStreams) int {
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
			WarnStdout(ios, "  %s: unreachable", ep.name)
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

// CheckDeviceConnectivity displays device connectivity diagnostics.
func CheckDeviceConnectivity(ctx context.Context, ios *iostreams.IOStreams, svc *shelly.Service) int {
	ios.Println(theme.Bold().Render("Device Connectivity"))

	devices := config.ListDevices()
	if len(devices) == 0 {
		ios.Info("  No devices to test")
		ios.Println()
		return 0
	}

	issues := 0
	online := 0

	ios.StartProgress("Testing device connectivity...")

	for name, d := range devices {
		testCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		_, err := svc.DevicePing(testCtx, d.Address)
		cancel()

		if err != nil {
			WarnStdout(ios, "  %s (%s): offline or unreachable", name, d.Address)
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
		WarnStdout(ios, "  %d device(s) unreachable", issues)
	}

	ios.Println()
	return issues
}
