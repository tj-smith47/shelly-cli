package wizard

import (
	"strings"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// RunCheck verifies the current setup without making changes.
func RunCheck(f *cmdutil.Factory) error {
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
