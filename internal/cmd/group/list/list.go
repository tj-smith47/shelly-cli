// Package list provides the group list subcommand.
package list

import (
	"sort"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// NewCommand creates the group list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewConfigListCommand(f, factories.ConfigListOpts[model.GroupInfo]{
		Resource: "group",
		FetchFunc: func() []model.GroupInfo {
			groups := config.ListGroups()
			result := make([]model.GroupInfo, 0, len(groups))
			for name, group := range groups {
				result = append(result, model.GroupInfo{
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
		DisplayFunc: term.DisplayGroups,
		EmptyMsg:    "No groups defined",
		HintMsg:     "Use 'shelly group create <name>' to create a group",
	})
}
