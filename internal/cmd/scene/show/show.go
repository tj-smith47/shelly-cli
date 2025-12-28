// Package show provides the scene show subcommand.
package show

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	flags.OutputFlags
	Factory *cmdutil.Factory
	Name    string
}

// NewCommand creates the scene show command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

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
			opts.Name = args[0]
			return run(opts)
		},
	}

	flags.AddOutputFlags(cmd, &opts.OutputFlags)

	return cmd
}

func run(opts *Options) error {
	scene, exists := config.GetScene(opts.Name)
	if !exists {
		return fmt.Errorf("scene %q not found", opts.Name)
	}

	switch opts.Format {
	case "json":
		return output.PrintJSON(scene)
	case "yaml":
		return output.PrintYAML(scene)
	default:
		term.DisplaySceneDetails(opts.Factory.IOStreams(), scene)
		return nil
	}
}
