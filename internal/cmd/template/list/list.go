// Package list provides the template list subcommand.
package list

import (
	"sort"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// NewCommand creates the template list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewConfigListCommand(f, factories.ConfigListOpts[config.DeviceTemplate]{
		Resource: "template",
		FetchFunc: func() []config.DeviceTemplate {
			templates := config.ListDeviceTemplates()
			result := make([]config.DeviceTemplate, 0, len(templates))
			for _, t := range templates {
				result = append(result, t)
			}
			sort.Slice(result, func(i, j int) bool {
				return result[i].Name < result[j].Name
			})
			return result
		},
		DisplayFunc: term.DisplayDeviceTemplateList,
		EmptyMsg:    "No templates configured",
	})
}
