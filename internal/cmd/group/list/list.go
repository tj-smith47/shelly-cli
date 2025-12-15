// Package list provides the group list subcommand.
package list

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// GroupInfo represents a group for JSON/YAML output.
type GroupInfo struct {
	Name        string   `json:"name" yaml:"name"`
	DeviceCount int      `json:"device_count" yaml:"device_count"`
	Devices     []string `json:"devices" yaml:"devices"`
}

// NewCommand creates the group list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all device groups",
		Long: `List all device groups and their member counts.

Groups allow organizing devices for batch operations. Each group can
contain multiple devices, and devices can belong to multiple groups.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting.

Columns: Name, Devices (count)`,
		Example: `  # List all groups
  shelly group list

  # Output as JSON
  shelly group list -o json

  # Get names of groups containing devices
  shelly group list -o json | jq -r '.[] | select(.device_count > 0) | .name'

  # List devices in all groups
  shelly group list -o json | jq -r '.[] | "\(.name): \(.devices | join(", "))"'

  # Count total groups
  shelly group list -o json | jq length

  # Short form
  shelly grp ls`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return run(f)
		},
	}

	return cmd
}

func run(f *cmdutil.Factory) error {
	ios := f.IOStreams()
	groups := config.ListGroups()

	if len(groups) == 0 {
		ios.Info("No groups defined")
		ios.Info("Use 'shelly group create <name>' to create a group")
		return nil
	}

	// Build sorted list for consistent output
	result := make([]GroupInfo, 0, len(groups))
	for name, group := range groups {
		result = append(result, GroupInfo{
			Name:        name,
			DeviceCount: len(group.Devices),
			Devices:     group.Devices,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	// Handle structured output (JSON/YAML) via global -o flag
	if output.WantsStructured() {
		return output.FormatOutput(ios.Out, result)
	}

	// Table output
	table := output.NewTable("Name", "Devices")
	for _, g := range result {
		table.AddRow(g.Name, formatDeviceCount(g.DeviceCount))
	}

	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print table", err)
	}
	ios.Count("group", len(result))

	return nil
}

func formatDeviceCount(count int) string {
	if count == 0 {
		return "0 (empty)"
	}
	if count == 1 {
		return "1 device"
	}
	return fmt.Sprintf("%d devices", count)
}
