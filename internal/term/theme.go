package term

import (
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// ApplyImportedTheme applies an imported theme and displays success message.
func ApplyImportedTheme(ios *iostreams.IOStreams, themeName string, colors map[string]string) error {
	if err := theme.ApplyConfig(themeName, colors, ""); err != nil {
		return fmt.Errorf("failed to apply theme: %w", err)
	}

	if len(colors) > 0 {
		ios.Success("Theme '%s' with %d color overrides imported and applied", themeName, len(colors))
	} else {
		ios.Success("Theme '%s' imported and applied", themeName)
	}
	return nil
}

// DisplayValidationResult displays theme validation results.
func DisplayValidationResult(ios *iostreams.IOStreams, themeName string, colors map[string]string) {
	ios.Success("Theme file validated successfully")
	if themeName != "" {
		ios.Info("Base theme: %s", themeName)
	}
	if len(colors) > 0 {
		ios.Info("Color overrides: %d", len(colors))
	}
	ios.Info("Use --apply to apply the theme")
}
