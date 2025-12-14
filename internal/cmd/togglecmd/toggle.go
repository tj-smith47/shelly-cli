// Package togglecmd provides the quick toggle command.
package togglecmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
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
this toggles between open and close based on current state.

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
			controllable, err := findControllable(ctx, conn, ios)
			if err != nil {
				return err
			}

			toControl := selectComponents(controllable, opts.All)

			for _, comp := range toControl {
				if err := toggleComponent(ctx, conn, comp, ios); err != nil {
					return err
				}
				toggled++
			}

			return nil
		})
	})
	if err != nil {
		return err
	}

	displayResult(ios, toggled, opts.Device)
	return nil
}

func findControllable(ctx context.Context, conn *client.Client, ios *iostreams.IOStreams) ([]model.Component, error) {
	components, err := conn.ListComponents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list components: %w", err)
	}

	var controllable []model.Component
	for _, comp := range components {
		if isControllable(comp.Type) {
			controllable = append(controllable, comp)
		} else {
			ios.Debug("skipping non-controllable component %s:%d", comp.Type, comp.ID)
		}
	}

	if len(controllable) == 0 {
		return nil, fmt.Errorf("no controllable components found on device")
	}

	return controllable, nil
}

func isControllable(t model.ComponentType) bool {
	switch t {
	case model.ComponentSwitch, model.ComponentLight, model.ComponentRGB, model.ComponentCover:
		return true
	case model.ComponentInput:
		return false
	default:
		return false
	}
}

func selectComponents(controllable []model.Component, all bool) []model.Component {
	if !all && len(controllable) > 1 {
		return controllable[:1]
	}
	return controllable
}

func toggleComponent(ctx context.Context, conn *client.Client, comp model.Component, ios *iostreams.IOStreams) error {
	var err error
	switch comp.Type {
	case model.ComponentSwitch:
		_, err = conn.Switch(comp.ID).Toggle(ctx)
	case model.ComponentLight:
		_, err = conn.Light(comp.ID).Toggle(ctx)
	case model.ComponentRGB:
		_, err = conn.RGB(comp.ID).Toggle(ctx)
	case model.ComponentCover:
		err = toggleCover(ctx, conn.Cover(comp.ID))
	default:
		ios.Debug("skipping unsupported component type %s:%d", comp.Type, comp.ID)
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to toggle %s:%d: %w", comp.Type, comp.ID, err)
	}
	return nil
}

func toggleCover(ctx context.Context, cover *client.CoverComponent) error {
	status, err := cover.GetStatus(ctx)
	if err != nil {
		return err
	}

	switch status.State {
	case "open", "opening":
		return cover.Close(ctx, nil)
	case "closed", "closing":
		return cover.Open(ctx, nil)
	default:
		// If stopped mid-way or unknown, open
		return cover.Open(ctx, nil)
	}
}

func displayResult(ios *iostreams.IOStreams, toggled int, device string) {
	if toggled == 1 {
		ios.Success("Device %q toggled", device)
	} else {
		ios.Success("Toggled %d components on %q", toggled, device)
	}
}
