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
	return factories.NewConfigListCommand(f, factories.ConfigListOpts[config.Template]{
		Resource: "template",
		FetchFunc: func() []config.Template {
			templates := config.ListTemplates()
			result := make([]config.Template, 0, len(templates))
			for _, t := range templates {
				result = append(result, t)
			}
			sort.Slice(result, func(i, j int) bool {
				return result[i].Name < result[j].Name
			})
			return result
		},
		DisplayFunc: term.DisplayTemplateList,
		EmptyMsg:    "No templates configured",
	})
}
