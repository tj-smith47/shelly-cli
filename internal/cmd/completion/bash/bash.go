// Package bash provides the bash completion command.
package bash

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
)

// NewCommand creates the bash completion command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewCompletionCommand(f, factories.CompletionOpts{
		Shell:   "bash",
		Aliases: []string{"b"},
		Long: `Generate the autocompletion script for bash.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

  source <(shelly completion bash)

To load completions for every new session, execute once:

  # Linux:
  shelly completion bash > /etc/bash_completion.d/shelly

  # macOS:
  shelly completion bash > $(brew --prefix)/etc/bash_completion.d/shelly`,
		Example: `  shelly completion bash > /tmp/shelly.bash
  source /tmp/shelly.bash`,
		Generator: func(cmd *cobra.Command, w io.Writer) error {
			return cmd.GenBashCompletion(w)
		},
	})
}
