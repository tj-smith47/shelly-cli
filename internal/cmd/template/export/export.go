// Package export provides the template export subcommand.
package export

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// NewCommand creates the template export command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewConfigExportCommand(f, factories.ConfigExportOpts[config.DeviceTemplate]{
		Component: "template",
		Aliases:   []string{"save", "dump"},
		Short:     "Export a template to a file",
		Long: `Export a configuration template to a JSON or YAML file.

If no file is specified, outputs to stdout.
Format is auto-detected from file extension, or can be specified with --format.`,
		Example: `  # Export to YAML file
  shelly template export my-config template.yaml

  # Export to JSON file
  shelly template export my-config template.json

  # Export to stdout as JSON
  shelly template export my-config --format json

  # Export to stdout as YAML
  shelly template export my-config`,
		ValidArgsFunc: completion.TemplateThenFile(),
		Fetcher:       config.GetDeviceTemplate,
		DefaultFormat: output.FormatYAML,
	})
}
