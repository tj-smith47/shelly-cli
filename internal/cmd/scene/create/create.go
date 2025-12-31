// Package create provides the scene create subcommand.
package create

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Options holds the options for the create command.
type Options struct {
	Factory     *cmdutil.Factory
	Description string
	Name        string
}

// NewCommand creates the scene create command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

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
			opts.Name = args[0]
			return run(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Description, "description", "d", "", "Scene description")

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	if err := config.CreateScene(opts.Name, opts.Description); err != nil {
		return err
	}

	ios.Success("Scene %q created", opts.Name)
	ios.Info("Add actions with: shelly scene add-action %s <device> <method> [params]", opts.Name)

	return nil
}
