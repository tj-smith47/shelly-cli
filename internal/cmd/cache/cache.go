// Package cache provides the cache command for managing CLI cache.
package cache

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/cache/clearcmd"
	"github.com/tj-smith47/shelly-cli/internal/cmd/cache/show"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the cache command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage CLI cache",
		Long: `Manage the Shelly CLI file cache.

The cache stores device data for faster access:
  - Device information and configuration
  - Firmware update status
  - Automation settings (schedules, webhooks, scripts)
  - Protocol configurations (MQTT, Modbus)

Cache data has type-specific TTLs and is shared between CLI and TUI.`,
		Aliases: []string{"ca"},
		Example: `  # Show cache statistics
  shelly cache show

  # Clear all cache
  shelly cache clear

  # Clear cache for specific device
  shelly cache clear --device kitchen

  # Clear only expired entries
  shelly cache clear --expired`,
	}

	cmd.AddCommand(clearcmd.NewCommand(f))
	cmd.AddCommand(show.NewCommand(f))

	return cmd
}
