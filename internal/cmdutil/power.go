package cmdutil

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// PowerStatusOptions selects the power reading a `power status` or `energy
// status` command prints.
type PowerStatusOptions struct {
	// All prints every reading of every registered device.
	All bool
	// Device is the device to read when All is false.
	Device string
	// Type is the component type ("auto" for any).
	Type string
	// ID is the component index; negative selects the first matching reading.
	ID int
}

// AddPowerTypeFlag adds the --type flag that picks which component's power
// reading a command uses.
func AddPowerTypeFlag(cmd *cobra.Command, target *string) {
	cmd.Flags().StringVar(target, "type", shelly.ComponentTypeAuto,
		"Component type: auto, or one of em, em1, pm, pm1, switch, cover, light, rgb, rgbw, cct, meter (Gen1), emeter (Gen1)")
}

// ParsePowerStatusArgs reads `[device] [id]` for a power or energy status
// command. With --all no argument is allowed; otherwise the device is
// required. A missing id is returned as -1.
func ParsePowerStatusArgs(args []string, all bool) (device string, id int, err error) {
	switch {
	case all && len(args) > 0:
		return "", -1, fmt.Errorf("--all reports every registered device; do not pass a device or component ID")
	case !all && len(args) == 0:
		return "", -1, fmt.Errorf("requires a device argument, or --all")
	}
	id = -1
	if len(args) > 0 {
		device = args[0]
	}
	if len(args) == 2 {
		if id, err = strconv.Atoi(args[1]); err != nil || id < 0 {
			return "", -1, fmt.Errorf("invalid component ID %q: want a number from 0", args[1])
		}
	}
	return device, id, nil
}

// RunPowerStatus prints one power reading of a device, whichever component
// carries it, or with opts.All every reading of every registered device as
// one list. Devices skipped under --all are named on stderr.
func RunPowerStatus(ctx context.Context, f *Factory, opts PowerStatusOptions) error {
	ios := f.IOStreams()
	svc := f.ShellyService()

	if !opts.All {
		reading, err := RunWithSpinnerResult(ctx, ios, "Reading power...", func(ctx context.Context) (model.PowerReading, error) {
			return svc.ReadPowerReading(ctx, opts.Device, opts.Type, opts.ID)
		})
		if err != nil {
			return err
		}
		return PrintResult(ios, reading, term.DisplayPowerReading)
	}

	cfg, err := f.Config()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	devices := slices.Sorted(maps.Keys(cfg.Devices))
	if len(devices) == 0 {
		ios.NoResults("registered devices", "Use 'shelly device add' to register devices first")
		return nil
	}

	var readings []model.PowerReading
	var skipped []model.EnergyStatusSkip
	err = RunWithSpinner(ctx, ios, fmt.Sprintf("Reading power on %d devices...", len(devices)), func(ctx context.Context) error {
		readings, skipped = svc.CollectPowerReadings(ctx, devices)
		return nil
	})
	if err != nil {
		return err
	}
	for _, sk := range skipped {
		ios.Warning("Skipped %s: %s", sk.Device, sk.Reason)
	}
	return PrintDiscovered(ios, readings, term.DisplayPowerReadingList, "power readings")
}

// RunPowerList prints every power reading of one device, one row per
// component that meters power.
func RunPowerList(ctx context.Context, f *Factory, device string) error {
	ios := f.IOStreams()
	svc := f.ShellyService()

	readings, err := RunWithSpinnerResult(ctx, ios, "Reading power...", func(ctx context.Context) ([]model.PowerReading, error) {
		return svc.ReadPowerReadings(ctx, device)
	})
	if err != nil {
		return fmt.Errorf("failed to read power on %s: %w", device, err)
	}
	return PrintDiscovered(ios, readings, term.DisplayPowerReadingList, "components that meter power")
}
