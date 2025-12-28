// Package template provides script template commands.
package template

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/script/template/install"
	"github.com/tj-smith47/shelly-cli/internal/cmd/script/template/list"
	"github.com/tj-smith47/shelly-cli/internal/cmd/script/template/show"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the script template command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "template",
		Aliases: []string{"tpl"},
		Short:   "Manage script templates",
		Long: `Manage JavaScript script templates for Shelly devices.

Script templates provide reusable JavaScript code that can be installed
on Gen2+ devices. Templates include configurable variables that are
substituted during installation.

The CLI includes built-in templates for common use cases:
  - motion-light: Motion-activated lighting
  - power-monitor: Power consumption alerts
  - schedule-helper: Simple on/off scheduling
  - toggle-sync: Synchronize multiple switches
  - energy-logger: Log energy usage to KVS`,
		Example: `  # List available script templates
  shelly script template list

  # Show template details and code
  shelly script template show motion-light

  # Install a template on a device
  shelly script template install living-room motion-light

  # Install with interactive configuration
  shelly script template install living-room motion-light --configure`,
	}

	cmd.AddCommand(list.NewCommand(f))
	cmd.AddCommand(show.NewCommand(f))
	cmd.AddCommand(install.NewCommand(f))

	return cmd
}
