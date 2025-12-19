// Package fish provides the fish completion command.
package fish

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
)

// NewCommand creates the fish completion command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewCompletionCommand(f, factories.CompletionOpts{
		Shell:   "fish",
		Aliases: []string{"f"},
		Long: `Generate the autocompletion script for fish.

To load completions in your current shell session:

  shelly completion fish | source

To load completions for every new session, execute once:

  shelly completion fish > ~/.config/fish/completions/shelly.fish`,
		Example: `  shelly completion fish > ~/.config/fish/completions/shelly.fish`,
		Generator: func(cmd *cobra.Command, w io.Writer) error {
			return cmd.GenFishCompletion(w, true)
		},
	})
}
