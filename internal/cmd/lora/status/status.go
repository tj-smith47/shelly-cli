// Package status provides the lora status command.
package status

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	flags.ComponentFlags
	Factory *cmdutil.Factory
	Device  string
	JSON    bool
}

// NewCommand creates the lora status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "status <device>",
		Aliases: []string{"st", "info"},
		Short:   "Show LoRa add-on status",
		Long: `Show LoRa add-on status for a Shelly device.

Displays the current LoRa configuration and signal quality
information from the last received packet.`,
		Example: `  # Show LoRa status
  shelly lora status living-room

  # Specify component ID (default: 100)
  shelly lora status living-room --id 100

  # Output as JSON
  shelly lora status living-room --json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	flags.AddComponentFlags(cmd, &opts.ComponentFlags, "LoRa")
	opts.ID = 100 // LoRa components start at ID 100
	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output as JSON")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	full, err := svc.Wireless().FetchLoRaFullStatus(ctx, opts.Device, opts.ID, ios)
	if err != nil {
		return err
	}

	if opts.JSON {
		return term.OutputLoRaStatusJSON(ios, full)
	}

	term.DisplayLoRaStatus(ios, full)
	return nil
}
