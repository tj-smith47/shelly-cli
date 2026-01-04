// Package list provides the virtual list command.
package list

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cache"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	flags.OutputFlags
	Device  string
	Factory *cmdutil.Factory
}

// NewCommand creates the virtual list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "list <device>",
		Aliases: []string{"ls", "l"},
		Short:   "List virtual components on a device",
		Long: `List all virtual components on a Shelly Gen2+ device.

Virtual components include boolean, number, text, enum, button, and group types.
Component IDs in the range 200-299 are reserved for virtual components.`,
		Example: `  # List all virtual components
  shelly virtual list kitchen

  # JSON output
  shelly virtual list kitchen -o json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	flags.AddOutputFlags(cmd, &opts.OutputFlags)

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	return cmdutil.RunCachedList(ctx, opts.Factory, opts.Device,
		cache.TypeVirtuals, cache.TTLAutomation,
		"Fetching virtual components...",
		"virtual components",
		func(ctx context.Context, svc *shelly.Service, device string) ([]shelly.VirtualComponent, error) {
			return svc.ListVirtualComponents(ctx, device)
		},
		func(ios *iostreams.IOStreams, items []shelly.VirtualComponent) {
			term.DisplayVirtualComponents(ios, items)
		})
}
