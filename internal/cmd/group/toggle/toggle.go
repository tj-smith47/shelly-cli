// Package toggle provides the group toggle subcommand for toggling group members.
package toggle

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// Options holds command options.
type Options struct {
	flags.QuickComponentFlags
	Factory    *cmdutil.Factory
	GroupName  string
	Concurrent int
}

// NewCommand creates the group toggle command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "toggle <group>",
		Aliases: []string{"flip"},
		Short:   "Toggle all group members",
		Long: `Toggle every device in a group.

The action is fanned out to all members concurrently and a per-member result
summary is printed. Works across mixed Gen1 and Gen2+ members. Omit --id to
control every controllable component on each member.`,
		Example: `  # Toggle every member of a group
  shelly group toggle guest-bath-bulbs

  # Toggle only component 1 on each member
  shelly group toggle living-room --id 1`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.GroupNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.GroupName = args[0]
			return run(cmd.Context(), opts)
		},
	}

	flags.AddQuickComponentFlags(cmd, &opts.QuickComponentFlags)
	cmd.Flags().IntVarP(&opts.Concurrent, "concurrent", "c", 5, "Max concurrent operations")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	f := opts.Factory
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	return cmdutil.RunGroupQuick(ctx, f, opts.GroupName, cmdutil.GroupActionToggle, opts.ComponentIDPointer(), opts.Concurrent)
}
