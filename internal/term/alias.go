// Package term provides terminal display functions.
package term

import (
	"strings"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// DisplayDeviceAliases displays aliases for a device.
// If aliases is empty, shows an info message.
func DisplayDeviceAliases(ios *iostreams.IOStreams, deviceName string, aliases []string) {
	if len(aliases) == 0 {
		ios.Info("No aliases defined for %s", deviceName)
		return
	}

	if output.WantsJSON() {
		if err := output.PrintJSON(map[string]any{
			"device":  deviceName,
			"aliases": aliases,
		}); err != nil {
			ios.DebugErr("print JSON", err)
		}
		return
	}
	if output.WantsYAML() {
		if err := output.PrintYAML(map[string]any{
			"device":  deviceName,
			"aliases": aliases,
		}); err != nil {
			ios.DebugErr("print YAML", err)
		}
		return
	}

	ios.Printf("Aliases for %s: %s\n", deviceName, strings.Join(aliases, ", "))
}

// DisplayAliasAdded shows success message for alias addition.
func DisplayAliasAdded(ios *iostreams.IOStreams, deviceName, alias string) {
	ios.Success("Added alias %q to %s", alias, deviceName)
}

// DisplayAliasRemoved shows success message for alias removal.
func DisplayAliasRemoved(ios *iostreams.IOStreams, deviceName, alias string) {
	ios.Success("Removed alias %q from %s", alias, deviceName)
}
