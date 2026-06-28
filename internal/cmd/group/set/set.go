// Package set provides the group set subcommand for controlling group members.
package set

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// Options holds command options.
type Options struct {
	flags.ComponentFlags
	Factory    *cmdutil.Factory
	GroupName  string
	Brightness int
	Temp       int
	On         bool
	Concurrent int
}

// NewCommand creates the group set command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Factory:    f,
		Brightness: -1,
		Temp:       -1,
	}

	cmd := &cobra.Command{
		Use:     "set <group>",
		Aliases: []string{"brightness", "br"},
		Short:   "Set light parameters on all group members",
		Long: `Set light parameters on every device in a group.

You can set brightness, white color temperature (Gen1 white-temp bulbs such as
the Duo), and on/off state. Values not specified are left unchanged. The change
is fanned out to all members concurrently and a per-member result summary is
printed.

Unlike on/off/toggle, --id targets a single light component (default 0) on each
member rather than all components.`,
		Example: `  # Set every member to 100% and turn on
  shelly group set guest-bath-bulbs -b 100 --on

  # Set white color temperature to 4200K on all members (Gen1 Duo)
  shelly group set master-bath -t 4200

  # Target component 1 on every member
  shelly group set living-room -b 50 --id 1`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.GroupNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.GroupName = args[0]
			return run(cmd.Context(), opts)
		},
	}

	flags.AddComponentFlags(cmd, &opts.ComponentFlags, "Light")
	cmd.Flags().IntVarP(&opts.Brightness, "brightness", "b", -1, "Brightness (0-100)")
	cmd.Flags().IntVarP(&opts.Temp, "temp", "t", -1, "White color temperature in Kelvin (Gen1 Duo: 2700-6500)")
	cmd.Flags().BoolVar(&opts.On, "on", false, "Turn on")
	cmd.Flags().IntVarP(&opts.Concurrent, "concurrent", "c", 5, "Max concurrent operations")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	f := opts.Factory
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	var brightnessPtr *int
	if opts.Brightness >= 0 && opts.Brightness <= 100 {
		brightnessPtr = &opts.Brightness
	}

	var tempPtr *int
	if opts.Temp > 0 {
		tempPtr = &opts.Temp
	}

	var onPtr *bool
	if opts.On {
		onPtr = &opts.On
	}

	return cmdutil.RunGroupLightSet(ctx, f, opts.GroupName, opts.Concurrent, opts.ID, brightnessPtr, tempPtr, onPtr)
}
