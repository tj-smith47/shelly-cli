// Package powershell provides the powershell completion command.
package powershell

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
)

// NewCommand creates the powershell completion command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewCompletionCommand(f, factories.CompletionOpts{
		Shell:   "powershell",
		Aliases: []string{"ps", "pwsh"},
		Long: `Generate the autocompletion script for PowerShell.

To load completions in your current shell session:

  shelly completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your PowerShell profile.`,
		Example: `  shelly completion powershell > shelly.ps1
  . ./shelly.ps1`,
		Generator: func(cmd *cobra.Command, w io.Writer) error {
			return cmd.GenPowerShellCompletionWithDesc(w)
		},
	})
}
