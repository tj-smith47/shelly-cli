// Package term provides terminal display functions.
package term

import (
	"strings"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// DisplayDeviceAliases displays aliases for a device. Under -o json, yaml
// or template it prints {"device", "aliases"} to ios.Out, with an empty
// aliases list when there are none; otherwise it prints one line.
func DisplayDeviceAliases(ios *iostreams.IOStreams, deviceName string, aliases []string) error {
	if output.WantsStructured() {
		if aliases == nil {
			aliases = []string{}
		}
		return output.FormatOutput(ios.Out, map[string]any{
			"device":  deviceName,
			"aliases": aliases,
		})
	}
	if len(aliases) == 0 {
		ios.Info("No aliases defined for %s", deviceName)
		return nil
	}
	ios.Printf("Aliases for %s: %s\n", deviceName, strings.Join(aliases, ", "))
	return nil
}

// DisplayAliasAdded shows success message for alias addition.
func DisplayAliasAdded(ios *iostreams.IOStreams, deviceName, alias string) {
	ios.Success("Added alias %q to %s", alias, deviceName)
}

// DisplayAliasRemoved shows success message for alias removal.
func DisplayAliasRemoved(ios *iostreams.IOStreams, deviceName, alias string) {
	ios.Success("Removed alias %q from %s", alias, deviceName)
}
