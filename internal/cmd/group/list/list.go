// Package list provides the group list subcommand.
package list

import (
	"sort"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
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
	return factories.NewConfigListCommand(f, factories.ConfigListOpts[GroupInfo]{
		Resource: "group",
		FetchFunc: func() []GroupInfo {
			groups := config.ListGroups()
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
			return result
		},
		DisplayFunc: displayGroups,
		EmptyMsg:    "No groups defined",
		HintMsg:     "Use 'shelly group create <name>' to create a group",
	})
}

func displayGroups(ios *iostreams.IOStreams, groups []GroupInfo) {
	table := output.NewTable("Name", "Devices")
	for _, g := range groups {
		table.AddRow(g.Name, output.FormatDeviceCount(g.DeviceCount))
	}

	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print table", err)
	}
	ios.Println()
	ios.Count("group", len(groups))
}
