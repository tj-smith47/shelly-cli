// Package exclusion provides the zwave exclusion command.
package exclusion

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/profiles"
	"github.com/tj-smith47/shelly-go/zwave"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds command options.
type Options struct {
	Model   string
	Mode    string
	Factory *cmdutil.Factory
}

// NewCommand creates the zwave exclusion command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "exclusion <model>",
		Aliases: []string{"exclude", "unpair", "remove"},
		Short:   "Show exclusion instructions",
		Long: `Show Z-Wave exclusion (unpairing) instructions for a device.

Exclusion modes:
  button - Manual exclusion using the S button (default)
  switch - Manual exclusion using the connected switch`,
		Example: `  # Show button-based exclusion (default)
  shelly zwave exclusion SNSW-001P16ZW

  # Show switch-based exclusion
  shelly zwave exclusion SNSW-001P16ZW --mode switch`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Model = args[0]
			return run(opts)
		},
	}

	cmd.Flags().StringVar(&opts.Mode, "mode", "button", "Exclusion mode (button, switch)")

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	profile, ok := profiles.Get(opts.Model)
	if !ok {
		return fmt.Errorf("unknown device model: %s", opts.Model)
	}

	device := zwave.NewDevice(profile)
	if !device.IsZWave() {
		return fmt.Errorf("%s is not a Z-Wave device", opts.Model)
	}

	var mode zwave.InclusionMode
	switch strings.ToLower(opts.Mode) {
	case "button", "s":
		mode = zwave.InclusionButton
	case "switch":
		mode = zwave.InclusionSwitch
	default:
		return fmt.Errorf("invalid mode %q, must be one of: button, switch", opts.Mode)
	}

	info := zwave.GetExclusionInfo(device, mode)

	ios.Title("Z-Wave Exclusion Instructions")
	ios.Println()

	ios.Printf("  %s: %s\n", theme.Dim().Render("Device"), theme.Highlight().Render(device.Name()))
	ios.Printf("  %s: %s\n", theme.Dim().Render("Mode"), string(info.Mode))

	ios.Println()
	ios.Printf("  %s:\n", theme.Dim().Render("Steps"))

	for _, step := range info.Instructions {
		ios.Printf("    %s\n", step)
	}

	ios.Println()
	return nil
}
