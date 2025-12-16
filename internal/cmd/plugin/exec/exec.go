// Package exec provides the extension exec command.
package exec

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
)

// NewCommand creates the extension exec command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "exec <name> [args...]",
		Aliases: []string{"run"},
		Short:   "Execute an extension",
		Long: `Execute an extension explicitly with the given arguments.

This is useful when the extension name conflicts with a built-in command
or when you want to explicitly invoke an extension.`,
		Example: `  # Run an extension
  shelly extension exec myext --some-flag

  # Run with arguments
  shelly extension exec myext arg1 arg2`,
		Args:               cobra.MinimumNArgs(1),
		ValidArgsFunction:  completion.ExtensionNames(),
		DisableFlagParsing: true, // Pass all args to the extension
		RunE: func(_ *cobra.Command, args []string) error {
			return run(f, args[0], args[1:])
		},
	}

	return cmd
}

func run(_ *cmdutil.Factory, name string, args []string) error {
	return plugins.RunPlugin(name, args)
}
