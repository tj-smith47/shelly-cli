// Package remove provides the extension remove command.
package remove

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
)

// Options holds the options for the remove command.
type Options struct {
	Factory *cmdutil.Factory
	Name    string
}

// NewCommand creates the extension remove command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "remove <name>",
		Aliases: []string{"rm", "uninstall", "delete"},
		Short:   "Remove an installed extension",
		Long: `Remove an installed extension.

Only extensions installed in the user plugins directory can be removed.
Extensions found in PATH cannot be removed with this command.`,
		Example: `  # Remove an extension
  shelly extension remove myext`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ExtensionNames(),
		RunE: func(_ *cobra.Command, args []string) error {
			opts.Name = args[0]
			return run(opts)
		},
	}

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	registry, err := plugins.NewRegistry()
	if err != nil {
		return err
	}

	if err := registry.Remove(opts.Name); err != nil {
		return err
	}

	ios.Success("Removed extension '%s'", opts.Name)
	return nil
}
