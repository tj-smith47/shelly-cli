// Package export provides the scene export subcommand.
package export

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// NewCommand creates the scene export command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewConfigExportCommand(f, factories.ConfigExportOpts[config.Scene]{
		Component: "scene",
		Aliases:   []string{"save", "backup"},
		Short:     "Export a scene to file",
		Long: `Export a scene definition to a file.

If no file is specified, outputs to stdout.
Format is auto-detected from file extension (.json, .yaml, .yml).`,
		Example: `  # Export to YAML file
  shelly scene export movie-night scene.yaml

  # Export to JSON file
  shelly scene export movie-night scene.json

  # Export to stdout as YAML
  shelly scene export movie-night

  # Export to stdout as JSON
  shelly scene export movie-night --format json`,
		ValidArgsFunc: completion.SceneNames(),
		Fetcher:       config.GetScene,
		DefaultFormat: output.FormatYAML,
	})
}
