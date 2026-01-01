package wizard

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// CheckExistingConfig checks if a config file exists.
func CheckExistingConfig() (exists bool, path string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return false, ""
	}

	path = filepath.Join(home, ".config", "shelly", "config.yaml")
	exists, _ = afero.Exists(config.Fs(), path)
	return exists, path
}

// CheckAndConfirmConfig checks for existing config and confirms overwrite.
func CheckAndConfirmConfig(ios *iostreams.IOStreams, opts *Options) (bool, error) {
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

	exists, _ := afero.Exists(config.Fs(), completionPath)
	return exists
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
