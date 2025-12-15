// Package config provides the lora config command.
package config

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// Options holds command options.
type Options struct {
	Factory   *cmdutil.Factory
	Device    string
	ID        int
	Freq      int64
	Bandwidth int
	DataRate  int
	TxPower   int
}

// NewCommand creates the lora config command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "config <device>",
		Aliases: []string{"set", "configure"},
		Short:   "Configure LoRa settings",
		Long: `Configure LoRa add-on settings on a Shelly device.

Allows setting radio parameters:
- Frequency: RF frequency in Hz (e.g., 868000000 for 868 MHz)
- Bandwidth: Channel bandwidth setting
- Data Rate: Spreading factor (higher = longer range, lower throughput)
- TX Power: Transmit power in dBm

Common frequencies:
- EU: 868 MHz (868000000)
- US: 915 MHz (915000000)
- Asia: 433 MHz (433000000)`,
		Example: `  # Set frequency to 868 MHz (EU)
  shelly lora config living-room --freq 868000000

  # Set transmit power to 14 dBm
  shelly lora config living-room --power 14

  # Configure multiple settings
  shelly lora config living-room --freq 915000000 --power 20 --dr 7`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVar(&opts.ID, "id", 100, "LoRa component ID")
	cmd.Flags().Int64Var(&opts.Freq, "freq", 0, "RF frequency in Hz")
	cmd.Flags().IntVar(&opts.Bandwidth, "bw", 0, "Bandwidth setting")
	cmd.Flags().IntVar(&opts.DataRate, "dr", 0, "Data rate / spreading factor")
	cmd.Flags().IntVar(&opts.TxPower, "power", 0, "Transmit power in dBm")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Build config params
	config := make(map[string]any)
	if opts.Freq > 0 {
		config["freq"] = opts.Freq
	}
	if opts.Bandwidth > 0 {
		config["bw"] = opts.Bandwidth
	}
	if opts.DataRate > 0 {
		config["dr"] = opts.DataRate
	}
	if opts.TxPower > 0 {
		config["txp"] = opts.TxPower
	}

	if len(config) == 0 {
		ios.Warning("No configuration options specified.")
		ios.Info("Use --freq, --bw, --dr, or --power to set values.")
		ios.Info("See 'shelly lora config --help' for details.")
		return nil
	}

	if err := svc.LoRaSetConfig(ctx, opts.Device, opts.ID, config); err != nil {
		return err
	}

	ios.Success("LoRa configuration updated.")

	if opts.Freq > 0 {
		ios.Info("  Frequency: %d Hz (%.3f MHz)", opts.Freq, float64(opts.Freq)/1e6)
	}
	if opts.Bandwidth > 0 {
		ios.Info("  Bandwidth: %d", opts.Bandwidth)
	}
	if opts.DataRate > 0 {
		ios.Info("  Data Rate: %d", opts.DataRate)
	}
	if opts.TxPower > 0 {
		ios.Info("  TX Power: %d dBm", opts.TxPower)
	}

	return nil
}
