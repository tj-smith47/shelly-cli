// Package zwave provides the zwave command group.
package zwave

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/zwave/config"
	"github.com/tj-smith47/shelly-cli/internal/cmd/zwave/exclusion"
	"github.com/tj-smith47/shelly-cli/internal/cmd/zwave/inclusion"
	"github.com/tj-smith47/shelly-cli/internal/cmd/zwave/info"
	"github.com/tj-smith47/shelly-cli/internal/cmd/zwave/reset"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the zwave command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "zwave",
		Aliases: []string{"zw", "wave"},
		Short:   "Z-Wave device utilities",
		Long: `Utilities for working with Shelly Wave Z-Wave devices.

Shelly Wave devices are Z-Wave end devices that require a third-party
gateway/hub for full operation. They support both standard Z-Wave mesh
networks and Z-Wave Long Range (ZWLR) star topology.

Supported gateways: Home Assistant (Z-Wave JS), Hubitat, HomeSeer,
SmartThings, Vera/ezlo, OpenHAB, and other Z-Wave certified controllers.

Note: Many Wave devices also include WiFi or Ethernet connectivity,
allowing direct control via the standard Gen2 RPC API.`,
		Example: `  # Show Z-Wave device info
  shelly zwave info SNSW-001P16ZW

  # Show inclusion instructions
  shelly zwave inclusion SNSW-001P16ZW --mode button

  # Show exclusion instructions
  shelly zwave exclusion SNSW-001P16ZW

  # Show factory reset instructions
  shelly zwave reset SNSW-001P16ZW

  # Show common configuration parameters
  shelly zwave config`,
	}

	cmd.AddCommand(info.NewCommand(f))
	cmd.AddCommand(inclusion.NewCommand(f))
	cmd.AddCommand(exclusion.NewCommand(f))
	cmd.AddCommand(reset.NewCommand(f))
	cmd.AddCommand(config.NewCommand(f))

	return cmd
}
