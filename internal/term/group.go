// Package term provides composed terminal presentation for the CLI.
package term

import (
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// DisplayGroups displays a table of device groups.
func DisplayGroups(ios *iostreams.IOStreams, groups []model.GroupInfo) {
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

// DisplayGroupMembers displays members of a group in table format.
func DisplayGroupMembers(ios *iostreams.IOStreams, groupName string, devices []string) {
	ios.Title("Group: %s", groupName)
	ios.Printf("\n")

	t := output.NewTable("#", "Device")
	for i, device := range devices {
		t.AddRow(fmt.Sprintf("%d", i+1), device)
	}
	if err := t.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print group members table", err)
	}
	ios.Println()
	ios.Count("member", len(devices))
}
