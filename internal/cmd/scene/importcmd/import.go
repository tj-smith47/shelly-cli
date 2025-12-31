// Package importcmd provides the scene import subcommand.
package importcmd

import (
	"fmt"

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
Use --name to override the scene name from the file.`,
		Example: `  # Import from YAML file
  shelly scene import scene.yaml

  # Import from JSON file
  shelly scene import scene.json

  # Import with different name
  shelly scene import scene.yaml --name my-scene

  # Overwrite existing scene
  shelly scene import scene.yaml --overwrite`,
		SupportsNameArg: false,
		NameFlagEnabled: true,
		ForceFlagName:   "overwrite",
		Importer:        importScene,
	})
}

func importScene(file, nameOverride string, overwrite bool) (string, error) {
	scene, err := config.ParseSceneFile(file)
	if err != nil {
		return "", err
	}

	// Override name if specified
	if nameOverride != "" {
		scene.Name = nameOverride
	}

	if scene.Name == "" {
		return "", fmt.Errorf("scene name is required (use --name to specify)")
	}

	if err := config.ImportScene(scene, overwrite); err != nil {
		return "", err
	}

	return fmt.Sprintf("Imported scene %q with %d action(s)", scene.Name, len(scene.Actions)), nil
}
