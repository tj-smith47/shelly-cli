// Package term provides composed terminal presentation for the CLI.
package term

import (
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/output/table"
)

// DisplayGroups displays a table of device groups.
func DisplayGroups(ios *iostreams.IOStreams, groups []model.GroupInfo) {
	builder := table.NewBuilder("Name", "Devices")
	for _, g := range groups {
		builder.AddRow(g.Name, output.FormatDeviceCount(g.DeviceCount))
	}

	table := builder.WithModeStyle(ios).Build()
	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print groups table", err)
	}
	ios.Println()
	ios.Count("group", len(groups))
}

// DisplayGroupMembers displays members of a group in table format.
func DisplayGroupMembers(ios *iostreams.IOStreams, groupName string, devices []string) {
	ios.Title("Group: %s", groupName)
	ios.Printf("\n")

	builder := table.NewBuilder("#", "Device")
	for i, device := range devices {
		builder.AddRow(fmt.Sprintf("%d", i+1), device)
	}
	table := builder.WithModeStyle(ios).Build()
	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print group members table", err)
	}
	ios.Println()
	ios.Count("member", len(devices))
}
