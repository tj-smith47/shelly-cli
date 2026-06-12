// Package inspect provides the provision inspect subcommand.
package inspect

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	SSID    string
	APIP    string
}

// NewCommand creates the provision inspect command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "inspect <ap-ssid>",
		Aliases: []string{"peek", "ap-read"},
		Short:   "Read a device's persisted config at its factory WiFi AP",
		Long: `Hop onto a device's factory WiFi access point, read the configuration it has
actually persisted (identity plus WiFi station settings), and return to the home
network.

Use this when a device configures but never appears on the LAN: it shows whether
the station SSID, key, and static IP took, and whether the device has associated
yet — answering "did the onboard / restore --to-ap actually write what I expected?"
without the device having to join the network first.`,
		Example: `  # Inspect a Shelly bulb sitting at its factory AP
  shelly provision inspect ShellyBulbDuo-D0DCFF

  # Use a specific host IP on the AP subnet
  shelly provision inspect ShellyBulbDuo-D0DCFF --ap-ip 192.168.33.150

  # JSON output
  shelly provision inspect ShellyBulbDuo-D0DCFF -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.SSID = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.APIP, "ap-ip", "", "Static host IP to use on the device's AP subnet (default 192.168.33.133)")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	// The AP hop plus a device read far exceed a normal request's budget.
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout*10)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	var insp *shelly.APInspection
	err := cmdutil.RunWithSpinner(ctx, ios,
		fmt.Sprintf("Inspecting device at AP %s (hopping host WiFi)...", opts.SSID),
		func(ctx context.Context) error {
			var inspErr error
			insp, inspErr = svc.InspectAtAP(ctx, opts.SSID, opts.APIP)
			return inspErr
		})
	if err != nil {
		return fmt.Errorf("failed to inspect device at AP: %w", err)
	}

	return cmdutil.PrintResult(ios, insp, term.DisplayAPInspection)
}
