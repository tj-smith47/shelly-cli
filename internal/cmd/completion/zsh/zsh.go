// Package zsh provides the zsh completion command.
package zsh

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
)

// NewCommand creates the zsh completion command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewCompletionCommand(f, factories.CompletionOpts{
		Shell:   "zsh",
		Aliases: []string{"z"},
		Long: `Generate the autocompletion script for zsh.

If shell completion is not already enabled in your environment you will need
to enable it. You can execute the following once:

  echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

  source <(shelly completion zsh)

To load completions for every new session, execute once:

  shelly completion zsh > "${fpath[1]}/_shelly"`,
		Example: `  shelly completion zsh > ~/.zsh/completions/_shelly`,
		Generator: func(cmd *cobra.Command, w io.Writer) error {
			return cmd.GenZshCompletion(w)
		},
	})
}
