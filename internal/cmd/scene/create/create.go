// Package create provides the scene create subcommand.
package create

import (
	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"

	"github.com/tj-smith47/shelly-cli/internal/config"
)

// NewCommand creates the scene create command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var description string

	cmd := &cobra.Command{
		Use:     "create <name>",
		Aliases: []string{"new"},
		Short:   "Create a new scene",
		Long: `Create a new empty scene.

After creating a scene, use 'shelly scene add-action' to add device actions,
or import an existing scene definition from a file.`,
		Example: `  # Create a new scene
  shelly scene create movie-night

  # Create with description
  shelly scene create movie-night --description "Dim lights for movies"

  # Using alias
  shelly scene new bedtime

  # Short form
  shelly sc create morning-routine`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return run(args[0], description)
		},
	}

	cmd.Flags().StringVarP(&description, "description", "d", "", "Scene description")

	return cmd
}

func run(name, description string) error {
	err := config.CreateScene(name, description)
	if err != nil {
		return err
	}

	iostreams.Success("Scene %q created", name)
	iostreams.Info("Add actions with: shelly scene add-action %s <device> <method> [params]", name)

	return nil
}
