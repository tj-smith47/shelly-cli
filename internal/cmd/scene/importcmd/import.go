// Package importcmd provides the scene import subcommand.
package importcmd

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// NewCommand creates the scene import command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewConfigImportCommand(f, factories.ConfigImportOpts{
		Component: "scene",
		Aliases:   []string{"load", "restore"},
		Short:     "Import a scene from file",
		Long: `Import a scene definition from a file.

Format is auto-detected from file extension (.json, .yaml, .yml).
Use --name to override the scene name from the file, and --force to replace
a scene that already has that name.`,
		Example: `  # Import from YAML file
  shelly scene import scene.yaml

  # Import from JSON file
  shelly scene import scene.json

  # Import with different name
  shelly scene import scene.yaml --name my-scene

  # Replace an existing scene with the same name
  shelly scene import scene.yaml --force`,
		SupportsNameArg: false,
		NameFlagEnabled: true,
		ForceFlagAlias:  "overwrite",
		Importer:        config.ImportSceneFromFile,
	})
}
