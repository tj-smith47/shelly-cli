// Package profile provides device profile commands.
package profile

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/profile/info"
	"github.com/tj-smith47/shelly-cli/internal/cmd/profile/list"
	"github.com/tj-smith47/shelly-cli/internal/cmd/profile/search"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the profile command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "profile",
		Aliases: []string{"profiles", "model", "models"},
		Short:   "Device profile information",
		Long: `Query device profiles and capabilities.

Device profiles contain static information about Shelly device models,
including hardware capabilities, supported protocols, and resource limits.

Use these commands to:
- List all known device models
- Look up capabilities by model number
- Search for devices with specific features`,
		Example: `  # List all device profiles
  shelly profile list

  # Show info for a specific model
  shelly profile info SNSW-001P16EU

  # Search for dimming-capable devices
  shelly profile search dimmer`,
	}

	cmd.AddCommand(
		list.NewCommand(f),
		info.NewCommand(f),
		search.NewCommand(f),
	)

	return cmd
}
