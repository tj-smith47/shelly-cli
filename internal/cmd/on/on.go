// Package on provides the quick on command.
package on

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds command options.
type Options struct {
	Device  string
	All     bool
	Factory *cmdutil.Factory
}

// NewCommand creates the on command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "on <device>",
		Aliases: []string{"turn-on", "enable"},
		Short:   "Turn on a device (auto-detects type)",
		Long: `Turn on a device by automatically detecting its type.

Works with switches, lights, covers, and RGB devices. For covers,
this opens them. For switches/lights/RGB, this turns them on.

Use --all to turn on all controllable components on the device.`,
		Example: `  # Turn on a switch or light
  shelly on living-room

  # Turn on all components on a device
  shelly on living-room --all

  # Open a cover
  shelly on bedroom-blinds`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.All, "all", "a", false, "Turn on all controllable components")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	var turnedOn int
	err := cmdutil.RunWithSpinner(ctx, ios, "Turning on...", func(ctx context.Context) error {
		return svc.WithConnection(ctx, opts.Device, func(conn *client.Client) error {
			components, err := conn.ListComponents(ctx)
			if err != nil {
				return fmt.Errorf("failed to list components: %w", err)
			}

			// Find controllable components
			var controllable []model.Component
			for _, comp := range components {
				switch comp.Type {
				case model.ComponentSwitch, model.ComponentLight, model.ComponentRGB, model.ComponentCover:
					controllable = append(controllable, comp)
				default:
					ios.Debug("skipping non-controllable component %s:%d", comp.Type, comp.ID)
				}
			}

			if len(controllable) == 0 {
				return fmt.Errorf("no controllable components found on device")
			}

			// If not --all, just control the first one (or first of each type)
			toControl := controllable
			if !opts.All && len(controllable) > 1 {
				// Just use the first controllable component
				toControl = controllable[:1]
			}

			for _, comp := range toControl {
				var opErr error
				switch comp.Type {
				case model.ComponentSwitch:
					opErr = conn.Switch(comp.ID).On(ctx)
				case model.ComponentLight:
					opErr = conn.Light(comp.ID).On(ctx)
				case model.ComponentRGB:
					opErr = conn.RGB(comp.ID).On(ctx)
				case model.ComponentCover:
					opErr = conn.Cover(comp.ID).Open(ctx, nil)
				default:
					ios.Debug("skipping unsupported component type %s:%d", comp.Type, comp.ID)
					continue
				}
				if opErr != nil {
					return fmt.Errorf("failed to turn on %s:%d: %w", comp.Type, comp.ID, opErr)
				}
				turnedOn++
			}

			return nil
		})
	})
	if err != nil {
		return err
	}

	if turnedOn == 1 {
		ios.Success("Device %q turned on", opts.Device)
	} else {
		ios.Success("Turned on %d components on %q", turnedOn, opts.Device)
	}
	return nil
}
