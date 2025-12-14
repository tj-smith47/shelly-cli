// Package off provides the quick off command.
package off

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds command options.
type Options struct {
	Device  string
	All     bool
	Factory *cmdutil.Factory
}

// NewCommand creates the off command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "off <device>",
		Aliases: []string{"turn-off", "disable"},
		Short:   "Turn off a device (auto-detects type)",
		Long: `Turn off a device by automatically detecting its type.

Works with switches, lights, covers, and RGB devices. For covers,
this closes them. For switches/lights/RGB, this turns them off.

Use --all to turn off all controllable components on the device.`,
		Example: `  # Turn off a switch or light
  shelly off living-room

  # Turn off all components on a device
  shelly off living-room --all

  # Close a cover
  shelly off bedroom-blinds`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.All, "all", "a", false, "Turn off all controllable components")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	var turnedOff int
	err := cmdutil.RunWithSpinner(ctx, ios, "Turning off...", func(ctx context.Context) error {
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

			// If not --all, just control the first one
			toControl := controllable
			if !opts.All && len(controllable) > 1 {
				toControl = controllable[:1]
			}

			for _, comp := range toControl {
				var opErr error
				switch comp.Type {
				case model.ComponentSwitch:
					opErr = conn.Switch(comp.ID).Off(ctx)
				case model.ComponentLight:
					opErr = conn.Light(comp.ID).Off(ctx)
				case model.ComponentRGB:
					opErr = conn.RGB(comp.ID).Off(ctx)
				case model.ComponentCover:
					opErr = conn.Cover(comp.ID).Close(ctx, nil)
				default:
					ios.Debug("skipping unsupported component type %s:%d", comp.Type, comp.ID)
					continue
				}
				if opErr != nil {
					return fmt.Errorf("failed to turn off %s:%d: %w", comp.Type, comp.ID, opErr)
				}
				turnedOff++
			}

			return nil
		})
	})
	if err != nil {
		return err
	}

	if turnedOff == 1 {
		ios.Success("Device %q turned off", opts.Device)
	} else {
		ios.Success("Turned off %d components on %q", turnedOff, opts.Device)
	}
	return nil
}
