// Package remove provides the extension remove command.
package remove

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
)

// Options holds the options for the remove command.
type Options struct {
	Factory *cmdutil.Factory
	Name    string
	Yes     bool
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
  shelly extension remove myext

  # Skip the confirmation prompt (scripts)
  shelly extension remove myext --yes`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ExtensionNames(),
		RunE: func(_ *cobra.Command, args []string) error {
			opts.Name = args[0]
			return run(opts)
		},
	}

	flags.AddYesFlag(cmd, &opts.Yes)

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	registry, err := plugins.NewRegistry()
	if err != nil {
		return err
	}

	if !registry.IsInstalled(opts.Name) {
		// Remove explains why: unknown name, or a plugin found on PATH outside the plugins directory.
		return registry.Remove(opts.Name)
	}

	confirmed, err := opts.Factory.ConfirmAction(fmt.Sprintf("Remove extension %q?", opts.Name), opts.Yes)
	if err != nil {
		return err
	}
	if !confirmed {
		ios.Info("Cancelled")
		return nil
	}

	if err := registry.Remove(opts.Name); err != nil {
		return err
	}

	ios.Success("Removed extension '%s'", opts.Name)
	return nil
}
