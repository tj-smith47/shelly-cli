// Package list provides the link list subcommand.
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

// NewCommand creates the link list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewConfigListCommand(f, factories.ConfigListOpts[model.LinkInfo]{
		Resource: "link",
		FetchFunc: func() []model.LinkInfo {
			links := config.ListLinks()
			result := make([]model.LinkInfo, 0, len(links))
			for child, link := range links {
				result = append(result, model.LinkInfo{
					ChildDevice:  child,
					ParentDevice: link.ParentDevice,
					SwitchID:     link.SwitchID,
				})
			}
			sort.Slice(result, func(i, j int) bool {
				return result[i].ChildDevice < result[j].ChildDevice
			})
			return result
		},
		DisplayFunc: term.DisplayLinks,
		EmptyMsg:    "No links defined",
		HintMsg:     "Use 'shelly link set <child> <parent>' to create a link",
	})
}
