// Package completion provides shell completion commands.
package completion

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/completion/bash"
	"github.com/tj-smith47/shelly-cli/internal/cmd/completion/fish"
	"github.com/tj-smith47/shelly-cli/internal/cmd/completion/install"
	"github.com/tj-smith47/shelly-cli/internal/cmd/completion/powershell"
	"github.com/tj-smith47/shelly-cli/internal/cmd/completion/zsh"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the completion command with subcommands.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "completion [command]",
		Aliases: []string{"comp"},
		Short:   "Generate shell completion scripts",
		Long: `Generate shell completion scripts for shelly.

To load completions:

Bash:

  $ source <(shelly completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ shelly completion bash > /etc/bash_completion.d/shelly
  # macOS:
  $ shelly completion bash > $(brew --prefix)/etc/bash_completion.d/shelly

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ shelly completion zsh > "${fpath[1]}/_shelly"

  # You will need to start a new shell for this setup to take effect.

Fish:

  $ shelly completion fish | source

  # To load completions for each session, execute once:
  $ shelly completion fish > ~/.config/fish/completions/shelly.fish

PowerShell:

  PS> shelly completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> shelly completion powershell > shelly.ps1
  # and source this file from your PowerShell profile.

Use 'shelly completion install' to automatically install completions.`,
		Example: `  # Generate bash completions
  shelly completion bash

  # Generate and install completions automatically
  shelly completion install

  # Install for specific shell
  shelly completion install --shell zsh`,
	}

	cmd.AddCommand(bash.NewCommand(f))
	cmd.AddCommand(zsh.NewCommand(f))
	cmd.AddCommand(fish.NewCommand(f))
	cmd.AddCommand(powershell.NewCommand(f))
	cmd.AddCommand(install.NewCommand(f))

	return cmd
}
