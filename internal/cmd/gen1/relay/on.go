package relay

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// OnOptions holds on command options.
type OnOptions struct {
	Factory  *cmdutil.Factory
	Device   string
	ID       int
	Duration int
}

func newOnCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &OnOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "on <device>",
		Short: "Turn relay on",
		Long: `Turn on a Gen1 relay switch.

Optionally specify a duration to auto-turn-off after the specified seconds.`,
		Example: `  # Turn relay on
  shelly gen1 relay on living-room

  # Turn on for 60 seconds
  shelly gen1 relay on living-room --duration 60

  # Turn on relay 1 (second relay)
  shelly gen1 relay on living-room --id 1`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return runOn(cmd.Context(), opts)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &opts.ID, "Relay")
	cmd.Flags().IntVar(&opts.Duration, "duration", 0, "Auto-off after seconds (0 = disabled)")

	return cmd
}

func runOn(ctx context.Context, opts *OnOptions) error {
	ios := opts.Factory.IOStreams()

	gen1Client, err := connectGen1(ctx, ios, opts.Device)
	if err != nil {
		return err
	}
	defer iostreams.CloseWithDebug("closing gen1 client", gen1Client)

	ios.StartProgress("Turning relay on...")

	relay := gen1Client.Relay(opts.ID)
	if opts.Duration > 0 {
		err = relay.TurnOnForDuration(ctx, opts.Duration)
	} else {
		err = relay.TurnOn(ctx)
	}

	ios.StopProgress()

	if err != nil {
		return err
	}

	if opts.Duration > 0 {
		ios.Success("Relay %d turned on for %d seconds", opts.ID, opts.Duration)
	} else {
		ios.Success("Relay %d turned on", opts.ID)
	}

	return nil
}

// connectGen1 resolves device config and connects to a Gen1 device.
func connectGen1(ctx context.Context, ios *iostreams.IOStreams, deviceName string) (*client.Gen1Client, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	devCfg, err := config.ResolveDevice(deviceName)
	if err != nil {
		return nil, err
	}

	device := model.Device{
		Name:    devCfg.Name,
		Address: devCfg.Address,
	}
	if devCfg.Auth != nil {
		device.Auth = &model.Auth{
			Username: devCfg.Auth.Username,
			Password: devCfg.Auth.Password,
		}
	}

	ios.StartProgress("Connecting to device...")
	gen1Client, err := client.ConnectGen1(ctx, device)
	ios.StopProgress()

	if err != nil {
		return nil, err
	}

	return gen1Client, nil
}
