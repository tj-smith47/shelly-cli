// Package importcmd provides the template import subcommand.
package importcmd

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// NewCommand creates the template import command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewConfigImportCommand(f, factories.ConfigImportOpts{
		Component: "template",
		Aliases:   []string{"load"},
		Short:     "Import a template from a file",
		Long: `Import a configuration template from a JSON or YAML file.

If no name is specified, the template name from the file is used.
Use --force to overwrite an existing template with the same name.`,
		Example: `  # Import a template
  shelly template import template.yaml

  # Import with a different name
  shelly template import template.yaml my-new-config

  # Overwrite existing template
  shelly template import template.yaml --force`,
		SupportsNameArg: true,
		NameFlagEnabled: false,
		ForceFlagName:   "force",
		ValidArgsFunc:   completion.FileThenNoComplete(),
		Importer:        config.ImportTemplateFromFile,
	})
}
