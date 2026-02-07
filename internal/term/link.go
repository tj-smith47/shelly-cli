// Package term provides composed terminal presentation for the CLI.
package term

import (
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output/table"
)

// DisplayLinks displays a table of device links.
func DisplayLinks(ios *iostreams.IOStreams, links []model.LinkInfo) {
	builder := table.NewBuilder("Child Device", "Parent Device", "Switch")
	for _, l := range links {
		builder.AddRow(l.ChildDevice, l.ParentDevice, fmt.Sprintf("%d", l.SwitchID))
	}

	tbl := builder.WithModeStyle(ios).Build()
	if err := tbl.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print links table", err)
	}
	ios.Println()
	ios.Count("link", len(links))
}

// DisplayLinkStatuses displays a table of link statuses with resolved state.
func DisplayLinkStatuses(ios *iostreams.IOStreams, statuses []model.LinkStatus) {
	builder := table.NewBuilder("Child Device", "Parent Device", "Switch", "Parent", "State")
	for _, s := range statuses {
		parentStatus := "Online"
		if !s.ParentOnline {
			parentStatus = "Offline"
		}
		builder.AddRow(s.ChildDevice, s.ParentDevice, fmt.Sprintf("%d", s.SwitchID), parentStatus, s.State)
	}

	tbl := builder.WithModeStyle(ios).Build()
	if err := tbl.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print link statuses table", err)
	}
	ios.Println()
	ios.Count("link", len(statuses))
}
