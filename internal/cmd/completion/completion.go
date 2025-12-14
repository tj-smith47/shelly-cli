// Package completion provides shell completion commands.
package completion

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/completion/install"
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

	// Add shell-specific completion subcommands
	cmd.AddCommand(newBashCmd())
	cmd.AddCommand(newZshCmd())
	cmd.AddCommand(newFishCmd())
	cmd.AddCommand(newPowerShellCmd())

	// Add install subcommand
	cmd.AddCommand(install.NewCommand(f))

	return cmd
}

func newBashCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "bash",
		Aliases: []string{"b"},
		Short:   "Generate bash completion script",
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
		Args:                  cobra.NoArgs,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenBashCompletion(os.Stdout)
		},
	}
}

func newZshCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "zsh",
		Aliases: []string{"z"},
		Short:   "Generate zsh completion script",
		Long: `Generate the autocompletion script for zsh.

If shell completion is not already enabled in your environment you will need
to enable it. You can execute the following once:

  echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

  source <(shelly completion zsh)

To load completions for every new session, execute once:

  shelly completion zsh > "${fpath[1]}/_shelly"`,
		Example:               `  shelly completion zsh > ~/.zsh/completions/_shelly`,
		Args:                  cobra.NoArgs,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenZshCompletion(os.Stdout)
		},
	}
}

func newFishCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "fish",
		Aliases: []string{"f"},
		Short:   "Generate fish completion script",
		Long: `Generate the autocompletion script for fish.

To load completions in your current shell session:

  shelly completion fish | source

To load completions for every new session, execute once:

  shelly completion fish > ~/.config/fish/completions/shelly.fish`,
		Example:               `  shelly completion fish > ~/.config/fish/completions/shelly.fish`,
		Args:                  cobra.NoArgs,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenFishCompletion(os.Stdout, true)
		},
	}
}

func newPowerShellCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "powershell",
		Aliases: []string{"ps", "pwsh"},
		Short:   "Generate PowerShell completion script",
		Long: `Generate the autocompletion script for PowerShell.

To load completions in your current shell session:

  shelly completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your PowerShell profile.`,
		Example: `  shelly completion powershell > shelly.ps1
  . ./shelly.ps1`,
		Args:                  cobra.NoArgs,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		},
	}
}
