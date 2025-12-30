// Package del provides the template delete subcommand.
package deletecmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// NewCommand creates the template delete command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewConfigDeleteCommand(f, factories.ConfigDeleteOpts{
		Resource:      "template",
		ValidArgsFunc: completion.TemplateNames(),
		ExistsFunc: func(name string) (any, bool) {
			return config.GetDeviceTemplate(name)
		},
		DeleteFunc: config.DeleteDeviceTemplate,
		InfoFunc: func(resource any, name string) string {
			tpl, ok := resource.(config.DeviceTemplate)
			if !ok || tpl.Model == "" {
				return fmt.Sprintf("Delete template %q?", name)
			}
			return fmt.Sprintf("Delete template %q (model: %s)?", tpl.Name, tpl.Model)
		},
	})
}
