// Package members provides the group members subcommand.
package members

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	flags.OutputFlags
	Factory   *cmdutil.Factory
	GroupName string
}

// NewCommand creates the group members command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "members <group>",
		Aliases: []string{"show", "ls"},
		Short:   "List group members",
		Long:    `List all devices that are members of the specified group.`,
		Example: `  # List members of a group
  shelly group members living-room

  # Output as JSON
  shelly group members living-room -o json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.GroupNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.GroupName = args[0]
			return run(cmd, opts)
		},
	}

	flags.AddOutputFlags(cmd, &opts.OutputFlags)

	return cmd
}

func run(cmd *cobra.Command, opts *Options) error {
	ios := opts.Factory.IOStreams()

	group := opts.Factory.GetGroup(opts.GroupName)
	if group == nil {
		return fmt.Errorf("group %q not found", opts.GroupName)
	}

	if len(group.Devices) == 0 {
		ios.NoResults("members in group %q", opts.GroupName)
		return nil
	}

	if output.WantsJSON() {
		data := map[string]any{
			"group":   opts.GroupName,
			"members": group.Devices,
			"count":   len(group.Devices),
		}
		return output.JSON(cmd.OutOrStdout(), data)
	}
	if output.WantsYAML() {
		data := map[string]any{
			"group":   opts.GroupName,
			"members": group.Devices,
			"count":   len(group.Devices),
		}
		return output.YAML(cmd.OutOrStdout(), data)
	}

	term.DisplayGroupMembers(ios, opts.GroupName, group.Devices)
	return nil
}
