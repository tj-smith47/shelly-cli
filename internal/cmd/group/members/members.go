// Package members provides the group members subcommand.
package members

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// NewCommand creates the group members command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var outputFormat string

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
			return run(f, cmd, args[0])
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")

	return cmd
}

func run(f *cmdutil.Factory, cmd *cobra.Command, groupName string) error {
	ios := f.IOStreams()

	group := f.GetGroup(groupName)
	if group == nil {
		return fmt.Errorf("group %q not found", groupName)
	}

	if len(group.Devices) == 0 {
		ios.NoResults("members in group %q", groupName)
		return nil
	}

	if output.WantsJSON() {
		data := map[string]any{
			"group":   groupName,
			"members": group.Devices,
			"count":   len(group.Devices),
		}
		return output.JSON(cmd.OutOrStdout(), data)
	}
	if output.WantsYAML() {
		data := map[string]any{
			"group":   groupName,
			"members": group.Devices,
			"count":   len(group.Devices),
		}
		return output.YAML(cmd.OutOrStdout(), data)
	}

	term.DisplayGroupMembers(ios, groupName, group.Devices)
	return nil
}
