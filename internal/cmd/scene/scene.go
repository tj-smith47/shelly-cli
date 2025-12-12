// Package scene provides the scene command group for managing device scenes.
package scene

import (
	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"

	"github.com/tj-smith47/shelly-cli/internal/cmd/scene/activate"
	"github.com/tj-smith47/shelly-cli/internal/cmd/scene/create"
	"github.com/tj-smith47/shelly-cli/internal/cmd/scene/deletecmd"
	"github.com/tj-smith47/shelly-cli/internal/cmd/scene/export"
	"github.com/tj-smith47/shelly-cli/internal/cmd/scene/importcmd"
	"github.com/tj-smith47/shelly-cli/internal/cmd/scene/list"
	"github.com/tj-smith47/shelly-cli/internal/cmd/scene/show"
)

// NewCommand creates the scene command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "scene",
		Aliases: []string{"sc"},
		Short:   "Manage device scenes",
		Long: `Manage saved device state configurations (scenes).

Scenes allow you to save and recall specific device states with a single command.
Each scene contains one or more actions that are executed when the scene is activated.`,
		Example: `  # List all scenes
  shelly scene list

  # Create a new scene
  shelly scene create movie-night

  # Show scene details
  shelly scene show movie-night

  # Activate a scene
  shelly scene activate movie-night

  # Export a scene to file
  shelly scene export movie-night scene.yaml

  # Import a scene from file
  shelly scene import scene.yaml

  # Delete a scene
  shelly scene delete movie-night`,
	}

	cmd.AddCommand(list.NewCommand(f))
	cmd.AddCommand(create.NewCommand(f))
	cmd.AddCommand(deletecmd.NewCommand(f))
	cmd.AddCommand(activate.NewCommand(f))
	cmd.AddCommand(show.NewCommand(f))
	cmd.AddCommand(export.NewCommand(f))
	cmd.AddCommand(importcmd.NewCommand(f))

	return cmd
}
