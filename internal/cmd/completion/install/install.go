// Package install provides the completion install command.
package install

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// NewCommand creates the completion install command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var shell string

	cmd := &cobra.Command{
		Use:     "install",
		Aliases: []string{"setup", "i"},
		Short:   "Install shell completions",
		Long: `Automatically install shell completions for shelly.

This command detects your current shell and installs completions
to the appropriate location. It also updates your shell configuration
to source the completions on startup.

Supported shells: bash, zsh, fish, powershell`,
		Example: `  # Auto-detect shell and install
  shelly completion install

  # Install for specific shell
  shelly completion install --shell bash
  shelly completion install --shell zsh`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(f, cmd.Root(), shell)
		},
	}

	cmd.Flags().StringVar(&shell, "shell", "", "Shell to install completions for (auto-detected if not specified)")

	return cmd
}

func run(f *cmdutil.Factory, rootCmd *cobra.Command, shell string) error {
	ios := f.IOStreams()

	// Auto-detect shell if not specified
	if shell == "" {
		var err error
		shell, err = completion.DetectShell()
		if err != nil {
			return fmt.Errorf("could not detect shell: %w\nSpecify shell with --shell flag", err)
		}
	}

	// Validate shell
	switch shell {
	case completion.ShellBash, completion.ShellZsh, completion.ShellFish, completion.ShellPowerShell:
		// Valid
	default:
		return fmt.Errorf("unsupported shell: %s\nSupported: bash, zsh, fish, powershell", shell)
	}

	ios.Info("Detected shell: %s", shell)

	return completion.GenerateAndInstall(ios, rootCmd, shell)
}
