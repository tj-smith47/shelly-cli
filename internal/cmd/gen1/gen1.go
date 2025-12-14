// Package gen1 provides Gen1 device management commands.
package gen1

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/gen1/actions"
	"github.com/tj-smith47/shelly-cli/internal/cmd/gen1/coiot"
	"github.com/tj-smith47/shelly-cli/internal/cmd/gen1/color"
	"github.com/tj-smith47/shelly-cli/internal/cmd/gen1/light"
	"github.com/tj-smith47/shelly-cli/internal/cmd/gen1/ota"
	"github.com/tj-smith47/shelly-cli/internal/cmd/gen1/relay"
	"github.com/tj-smith47/shelly-cli/internal/cmd/gen1/roller"
	"github.com/tj-smith47/shelly-cli/internal/cmd/gen1/settings"
	"github.com/tj-smith47/shelly-cli/internal/cmd/gen1/status"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the gen1 command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gen1",
		Aliases: []string{"g1", "legacy"},
		Short:   "Manage Gen1 Shelly devices",
		Long: `Manage Gen1 Shelly devices using their HTTP REST API.

Gen1 devices (released before 2021) include:
- Shelly 1, 1PM, 1L, 2, 2.5, 4Pro
- Shelly Plug, Plug S, Plug US
- Shelly Bulb, Vintage, Duo
- Shelly RGBW, RGBW2
- Shelly Dimmer, Dimmer 2
- Shelly EM, 3EM
- Shelly i3, Button1
- Shelly Gas, Smoke, Flood
- Shelly Door/Window, H&T, Motion
- Shelly UNI

Gen1 devices use a different API than Gen2+ devices:
- HTTP REST endpoints instead of RPC
- Action URLs instead of webhooks/scripts
- CoIoT (CoAP) for real-time updates

For Gen2+ devices, use the standard commands directly.`,
		Example: `  # Show Gen1 device status
  shelly gen1 status living-room

  # View Gen1 device settings
  shelly gen1 settings living-room

  # List Gen1 action URLs
  shelly gen1 actions living-room`,
	}

	cmd.AddCommand(status.NewCommand(f))
	cmd.AddCommand(settings.NewCommand(f))
	cmd.AddCommand(actions.NewCommand(f))
	cmd.AddCommand(relay.NewCommand(f))
	cmd.AddCommand(roller.NewCommand(f))
	cmd.AddCommand(light.NewCommand(f))
	cmd.AddCommand(color.NewCommand(f))
	cmd.AddCommand(ota.NewCommand(f))
	cmd.AddCommand(coiot.NewCommand(f))

	return cmd
}
