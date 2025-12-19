// Package show provides the scene show subcommand.
package show

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// NewCommand creates the scene show command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:     "show <name>",
		Aliases: []string{"info", "get"},
		Short:   "Show scene details",
		Long:    `Display detailed information about a scene including all its actions.`,
		Example: `  # Show scene details
  shelly scene show movie-night

  # Output as JSON
  shelly scene show movie-night --output json

  # Using alias
  shelly scene info bedtime

  # Short form
  shelly sc show morning-routine`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.SceneNames(),
		RunE: func(_ *cobra.Command, args []string) error {
			return run(f, args[0], outputFormat)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format: table, json, yaml")

	return cmd
}

func run(f *cmdutil.Factory, name, outputFormat string) error {
	scene, exists := config.GetScene(name)
	if !exists {
		return fmt.Errorf("scene %q not found", name)
	}

	switch outputFormat {
	case "json":
		return output.PrintJSON(scene)
	case "yaml":
		return output.PrintYAML(scene)
	default:
		term.DisplaySceneDetails(f.IOStreams(), scene)
		return nil
	}
}
