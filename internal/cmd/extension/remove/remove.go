// Package remove provides the extension remove command.
package remove

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
)

// NewCommand creates the extension remove command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
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
		ValidArgsFunction: cmdutil.CompleteExtensionNames(),
		RunE: func(_ *cobra.Command, args []string) error {
			return run(f, args[0])
		},
	}

	return cmd
}

func run(f *cmdutil.Factory, name string) error {
	ios := f.IOStreams()

	registry, err := plugins.NewRegistry()
	if err != nil {
		return err
	}

	if err := registry.Remove(name); err != nil {
		return err
	}

	ios.Success("Removed extension '%s'", name)
	return nil
}
