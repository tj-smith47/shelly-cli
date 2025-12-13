// Package deletecmd provides the scene delete subcommand.
package deletecmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// NewCommand creates the scene delete command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:     "delete <name>",
		Aliases: []string{"rm", "del"},
		Short:   "Delete a scene",
		Long:    `Delete a saved scene permanently.`,
		Example: `  # Delete a scene (with confirmation)
  shelly scene delete movie-night

  # Delete without confirmation
  shelly scene delete movie-night --yes

  # Using alias
  shelly scene rm bedtime

  # Short form
  shelly sc del old-scene`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return run(args[0], yes)
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func run(name string, skipConfirm bool) error {
	// Check if scene exists
	scene, exists := config.GetScene(name)
	if !exists {
		return fmt.Errorf("scene %q not found", name)
	}

	// Confirm unless --yes is specified
	if !skipConfirm {
		msg := fmt.Sprintf("Delete scene %q?", name)
		if len(scene.Actions) > 0 {
			msg = fmt.Sprintf("Delete scene %q with %d action(s)?", name, len(scene.Actions))
		}
		confirmed, err := iostreams.Confirm(msg, false)
		if err != nil {
			return fmt.Errorf("confirmation failed: %w", err)
		}
		if !confirmed {
			iostreams.Info("Deletion cancelled")
			return nil
		}
	}

	err := config.DeleteScene(name)
	if err != nil {
		return fmt.Errorf("failed to delete scene: %w", err)
	}

	iostreams.Success("Scene %q deleted", name)

	return nil
}
