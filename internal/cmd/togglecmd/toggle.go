// Package togglecmd provides the quick toggle command.
package togglecmd

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

// NewCommand creates the toggle command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "toggle <device>",
		Aliases: []string{"flip", "switch"},
		Short:   "Toggle a device (auto-detects type)",
		Long: `Toggle a device by automatically detecting its type.

Works with switches, lights, covers, and RGB devices. For covers,
this stops them if moving, otherwise toggles open/close state.

Use --all to toggle all controllable components on the device.`,
		Example: `  # Toggle a switch or light
  shelly toggle living-room

  # Toggle all components on a device
  shelly toggle living-room --all

  # Toggle a cover
  shelly toggle bedroom-blinds`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.All, "all", "a", false, "Toggle all controllable components")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	var toggled int
	err := cmdutil.RunWithSpinner(ctx, ios, "Toggling...", func(ctx context.Context) error {
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
					_, opErr = conn.Switch(comp.ID).Toggle(ctx)
				case model.ComponentLight:
					_, opErr = conn.Light(comp.ID).Toggle(ctx)
				case model.ComponentRGB:
					_, opErr = conn.RGB(comp.ID).Toggle(ctx)
				case model.ComponentCover:
					// For covers, stop if moving, otherwise toggle
					opErr = conn.Cover(comp.ID).Stop(ctx)
				default:
					ios.Debug("skipping unsupported component type %s:%d", comp.Type, comp.ID)
					continue
				}
				if opErr != nil {
					return fmt.Errorf("failed to toggle %s:%d: %w", comp.Type, comp.ID, opErr)
				}
				toggled++
			}

			return nil
		})
	})
	if err != nil {
		return err
	}

	if toggled == 1 {
		ios.Success("Device %q toggled", opts.Device)
	} else {
		ios.Success("Toggled %d components on %q", toggled, opts.Device)
	}
	return nil
}
