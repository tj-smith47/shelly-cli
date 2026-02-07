// Package set provides the link set subcommand.
package set

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Options holds the options for the set command.
type Options struct {
	Factory      *cmdutil.Factory
	ChildDevice  string
	ParentDevice string
	SwitchID     int
}

// NewCommand creates the link set command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "set <child-device> <parent-device>",
		Aliases: []string{"create", "add"},
		Short:   "Set a device power link",
		Long: `Set a parent-child power link between devices.

The child device is powered by a switch on the parent device. When the
child is offline, its state can be derived from the parent switch state.`,
		Example: `  # Link bulb to switch:0 on bedroom-2pm
  shelly link set bulb-duo bedroom-2pm

  # Link to a specific switch ID
  shelly link set garage-light garage-switch --switch-id 1

  # Update an existing link
  shelly link set bulb-duo new-switch`,
		Args: cobra.ExactArgs(2),
		ValidArgsFunction: func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completion.DeviceNames()(nil, args, toComplete)
		},
		RunE: func(_ *cobra.Command, args []string) error {
			opts.ChildDevice = args[0]
			opts.ParentDevice = args[1]
			return run(opts)
		},
	}

	cmd.Flags().IntVar(&opts.SwitchID, "switch-id", 0, "Switch component ID on the parent device")

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	if err := config.SetLink(opts.ChildDevice, opts.ParentDevice, opts.SwitchID); err != nil {
		return fmt.Errorf("failed to set link: %w", err)
	}

	ios.Success("Link set: %q -> %q switch:%d", opts.ChildDevice, opts.ParentDevice, opts.SwitchID)
	return nil
}
