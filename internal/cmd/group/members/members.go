// Package members provides the group members subcommand.
package members

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// NewCommand creates the group members command.
func NewCommand() *cobra.Command {
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
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd, args[0])
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")

	return cmd
}

func run(cmd *cobra.Command, groupName string) error {
	group, ok := config.GetGroup(groupName)
	if !ok {
		return fmt.Errorf("group %q not found", groupName)
	}

	if len(group.Devices) == 0 {
		iostreams.NoResults("members in group %q", groupName)
		return nil
	}

	return outputMembers(cmd, groupName, group.Devices)
}

func outputMembers(cmd *cobra.Command, groupName string, devices []string) error {
	switch viper.GetString("output") {
	case string(output.FormatJSON):
		data := map[string]any{
			"group":   groupName,
			"members": devices,
			"count":   len(devices),
		}
		return output.JSON(cmd.OutOrStdout(), data)
	case string(output.FormatYAML):
		data := map[string]any{
			"group":   groupName,
			"members": devices,
			"count":   len(devices),
		}
		return output.YAML(cmd.OutOrStdout(), data)
	default:
		printTable(groupName, devices)
		return nil
	}
}

func printTable(groupName string, devices []string) {
	iostreams.Title("Group: %s", groupName)
	fmt.Println()

	t := output.NewTable("#", "Device")
	for i, device := range devices {
		t.AddRow(fmt.Sprintf("%d", i+1), device)
	}
	t.Print()

	iostreams.Count("member", len(devices))
}
