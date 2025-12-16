// Package del provides the template delete subcommand.
package del

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// NewCommand creates the template delete command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return cmdutil.NewConfigDeleteCommand(f, cmdutil.ConfigDeleteOpts{
		Resource:      "template",
		ValidArgsFunc: completion.TemplateNames(),
		ExistsFunc: func(name string) (any, bool) {
			return config.GetTemplate(name)
		},
		DeleteFunc: config.DeleteTemplate,
		InfoFunc: func(resource any, name string) string {
			tpl, ok := resource.(config.Template)
			if !ok || tpl.Model == "" {
				return fmt.Sprintf("Delete template %q?", name)
			}
			return fmt.Sprintf("Delete template %q (model: %s)?", tpl.Name, tpl.Model)
		},
	})
}
