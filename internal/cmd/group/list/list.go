// Package list provides the group list subcommand.
package list

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// NewCommand creates the group list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all device groups",
		Long:    `List all device groups and their member counts.`,
		Example: `  # List all groups
  shelly group list

  # Output as JSON
  shelly group list --output json

  # Short form
  shelly grp ls`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return run(outputFormat)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format: table, json, yaml")

	return cmd
}

// GroupInfo represents a group for JSON/YAML output.
type GroupInfo struct {
	Name        string   `json:"name" yaml:"name"`
	DeviceCount int      `json:"device_count" yaml:"device_count"`
	Devices     []string `json:"devices" yaml:"devices"`
}

func run(outputFormat string) error {
	groups := config.ListGroups()

	if len(groups) == 0 {
		iostreams.Info("No groups defined")
		iostreams.Info("Use 'shelly group create <name>' to create a group")
		return nil
	}

	switch outputFormat {
	case "json":
		return outputJSON(groups)
	case "yaml":
		return outputYAML(groups)
	default:
		return outputTable(groups)
	}
}

func outputJSON(groups map[string]config.Group) error {
	result := make([]GroupInfo, 0, len(groups))
	for name, group := range groups {
		result = append(result, GroupInfo{
			Name:        name,
			DeviceCount: len(group.Devices),
			Devices:     group.Devices,
		})
	}
	return output.PrintJSON(result)
}

func outputYAML(groups map[string]config.Group) error {
	result := make([]GroupInfo, 0, len(groups))
	for name, group := range groups {
		result = append(result, GroupInfo{
			Name:        name,
			DeviceCount: len(group.Devices),
			Devices:     group.Devices,
		})
	}
	return output.PrintYAML(result)
}

func outputTable(groups map[string]config.Group) error {
	table := output.NewTable("Name", "Devices")

	for name, group := range groups {
		count := len(group.Devices)
		table.AddRow(name, formatDeviceCount(count))
	}

	table.Print()
	iostreams.Count("group", len(groups))

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
