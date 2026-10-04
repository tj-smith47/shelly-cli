// Package members provides the group members subcommand.
package members

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// keyGroup is the output field name for the group identifier.
const keyGroup = "group"

// Options holds command options.
type Options struct {
	Factory   *cmdutil.Factory
	GroupName string
}

// NewCommand creates the group members command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "members <group>",
		Aliases: []string{"show"},
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
			return run(opts)
		},
	}

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	group := opts.Factory.GetGroup(opts.GroupName)
	if group == nil {
		return fmt.Errorf("group %q not found", opts.GroupName)
	}

	members := group.Devices
	if members == nil {
		members = []string{}
	}
	data := map[string]any{
		keyGroup:  opts.GroupName,
		"members": members,
		"count":   len(members),
	}
	return cmdutil.PrintResult(ios, data, func(ios *iostreams.IOStreams, _ map[string]any) {
		if len(members) == 0 {
			ios.NoResults(fmt.Sprintf("members in group %q", opts.GroupName))
			return
		}
		term.DisplayGroupMembers(ios, opts.GroupName, members)
	})
}
