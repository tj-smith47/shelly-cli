package roller

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

// OpenOptions holds open command options.
type OpenOptions struct {
	Factory  *cmdutil.Factory
	Device   string
	ID       int
	Duration float64
}

func newOpenCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &OpenOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "open <device>",
		Short: "Open roller/cover",
		Long: `Start opening a Gen1 roller/cover.

Optionally specify a duration to open for a specific time.`,
		Example: `  # Open roller fully
  shelly gen1 roller open living-room

  # Open for 5 seconds
  shelly gen1 roller open living-room --duration 5

  # Open roller 1 (second roller)
  shelly gen1 roller open living-room --id 1`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return runOpen(cmd.Context(), opts)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &opts.ID, "Roller")
	cmd.Flags().Float64Var(&opts.Duration, "duration", 0, "Open for specified seconds (0 = fully open)")

	return cmd
}

func runOpen(ctx context.Context, opts *OpenOptions) error {
	ios := opts.Factory.IOStreams()

	gen1Client, err := connectGen1(ctx, ios, opts.Device)
	if err != nil {
		return err
	}
	defer iostreams.CloseWithDebug("closing gen1 client", gen1Client)

	ios.StartProgress("Opening roller...")

	roller := gen1Client.Roller(opts.ID)
	if opts.Duration > 0 {
		err = roller.OpenForDuration(ctx, opts.Duration)
	} else {
		err = roller.Open(ctx)
	}

	ios.StopProgress()

	if err != nil {
		return err
	}

	if opts.Duration > 0 {
		ios.Success("Roller %d opening for %.1f seconds", opts.ID, opts.Duration)
	} else {
		ios.Success("Roller %d opening", opts.ID)
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
