// Package completion provides shell completion generation.
package completion

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// NewCommand creates the completion command and its subcommands.
func NewCommand(rootCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for shelly.

To load completions:

Bash:
  # Linux:
  $ shelly completion bash > /etc/bash_completion.d/shelly
  # macOS:
  $ shelly completion bash > $(brew --prefix)/etc/bash_completion.d/shelly

Zsh:
  # If shell completion is not already enabled:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # Generate completion:
  $ shelly completion zsh > "${fpath[1]}/_shelly"
  # Or for Oh My Zsh:
  $ shelly completion zsh > ~/.oh-my-zsh/completions/_shelly

Fish:
  $ shelly completion fish > ~/.config/fish/completions/shelly.fish

PowerShell:
  PS> shelly completion powershell > shelly.ps1
  # Then source this file from your PowerShell profile.

After installation, restart your shell or source the completion file.`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				return rootCmd.GenZshCompletion(os.Stdout)
			case "fish":
				return rootCmd.GenFishCompletion(os.Stdout, true)
			case "powershell":
				return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
			default:
				return fmt.Errorf("unknown shell: %s", args[0])
			}
		},
	}

	// Add individual shell commands for convenience
	cmd.AddCommand(newBashCmd(rootCmd))
	cmd.AddCommand(newZshCmd(rootCmd))
	cmd.AddCommand(newFishCmd(rootCmd))
	cmd.AddCommand(newPowerShellCmd(rootCmd))
	cmd.AddCommand(newInstallCmd(rootCmd))

	return cmd
}

func newBashCmd(rootCmd *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   "bash",
		Short: "Generate bash completion script",
		Long: `Generate bash completion script for shelly.

To load completions:
  # Linux:
  $ shelly completion bash > /etc/bash_completion.d/shelly

  # macOS:
  $ shelly completion bash > $(brew --prefix)/etc/bash_completion.d/shelly`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return rootCmd.GenBashCompletion(os.Stdout)
		},
	}
}

func newZshCmd(rootCmd *cobra.Command) *cobra.Command {
	var noDescriptions bool

	cmd := &cobra.Command{
		Use:   "zsh",
		Short: "Generate zsh completion script",
		Long: `Generate zsh completion script for shelly.

To load completions:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
  $ shelly completion zsh > "${fpath[1]}/_shelly"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if noDescriptions {
				return rootCmd.GenZshCompletionNoDesc(os.Stdout)
			}
			return rootCmd.GenZshCompletion(os.Stdout)
		},
	}

	cmd.Flags().BoolVar(&noDescriptions, "no-descriptions", false, "Disable completion descriptions")

	return cmd
}

func newFishCmd(rootCmd *cobra.Command) *cobra.Command {
	var noDescriptions bool

	cmd := &cobra.Command{
		Use:   "fish",
		Short: "Generate fish completion script",
		Long: `Generate fish completion script for shelly.

To load completions:
  $ shelly completion fish > ~/.config/fish/completions/shelly.fish`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return rootCmd.GenFishCompletion(os.Stdout, !noDescriptions)
		},
	}

	cmd.Flags().BoolVar(&noDescriptions, "no-descriptions", false, "Disable completion descriptions")

	return cmd
}

func newPowerShellCmd(rootCmd *cobra.Command) *cobra.Command {
	var noDescriptions bool

	cmd := &cobra.Command{
		Use:   "powershell",
		Short: "Generate PowerShell completion script",
		Long: `Generate PowerShell completion script for shelly.

To load completions:
  PS> shelly completion powershell > shelly.ps1
  PS> . .\shelly.ps1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if noDescriptions {
				return rootCmd.GenPowerShellCompletion(os.Stdout)
			}
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		},
	}

	cmd.Flags().BoolVar(&noDescriptions, "no-descriptions", false, "Disable completion descriptions")

	return cmd
}

func newInstallCmd(rootCmd *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install completions for the current shell",
		Long: `Automatically install shell completions for the current shell.

This command detects your shell and installs completions appropriately.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			shell := detectShell()
			if shell == "" {
				return fmt.Errorf("could not detect shell; please use 'shelly completion <shell>' manually")
			}

			fmt.Printf("Detected shell: %s\n", shell)

			switch shell {
			case "bash":
				return installBashCompletion(rootCmd)
			case "zsh":
				return installZshCompletion(rootCmd)
			case "fish":
				return installFishCompletion(rootCmd)
			default:
				return fmt.Errorf("automatic installation not supported for %s; please use 'shelly completion %s' manually", shell, shell)
			}
		},
	}
}

func detectShell() string {
	// Check SHELL environment variable
	shell := os.Getenv("SHELL")
	if shell != "" {
		switch {
		case contains(shell, "bash"):
			return "bash"
		case contains(shell, "zsh"):
			return "zsh"
		case contains(shell, "fish"):
			return "fish"
		}
	}

	// Check PSModulePath for PowerShell
	if os.Getenv("PSModulePath") != "" {
		return "powershell"
	}

	return ""
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func installBashCompletion(rootCmd *cobra.Command) error {
	// Try common completion directories
	paths := []string{
		"/etc/bash_completion.d/shelly",
		os.Getenv("HOME") + "/.local/share/bash-completion/completions/shelly",
	}

	for _, path := range paths {
		f, err := os.Create(path)
		if err != nil {
			continue
		}
		defer f.Close()

		if err := rootCmd.GenBashCompletion(f); err != nil {
			return err
		}

		fmt.Printf("Installed bash completions to: %s\n", path)
		fmt.Println("Restart your shell or run: source " + path)
		return nil
	}

	return fmt.Errorf("could not write to completion directories; try running with sudo or use 'shelly completion bash > FILE' manually")
}

func installZshCompletion(rootCmd *cobra.Command) error {
	// Try Oh My Zsh first, then standard fpath locations
	home := os.Getenv("HOME")
	paths := []string{
		home + "/.oh-my-zsh/completions/_shelly",
		home + "/.zsh/completions/_shelly",
		"/usr/local/share/zsh/site-functions/_shelly",
	}

	for _, path := range paths {
		// Create directory if needed
		dir := path[:len(path)-len("/_shelly")]
		if err := os.MkdirAll(dir, 0o755); err != nil {
			continue
		}

		f, err := os.Create(path)
		if err != nil {
			continue
		}
		defer f.Close()

		if err := rootCmd.GenZshCompletion(f); err != nil {
			return err
		}

		fmt.Printf("Installed zsh completions to: %s\n", path)
		fmt.Println("Restart your shell or run: compinit")
		return nil
	}

	return fmt.Errorf("could not write to completion directories; use 'shelly completion zsh > FILE' manually")
}

func installFishCompletion(rootCmd *cobra.Command) error {
	home := os.Getenv("HOME")
	path := home + "/.config/fish/completions/shelly.fish"

	// Create directory if needed
	dir := home + "/.config/fish/completions"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("could not create completions directory: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("could not create completion file: %w", err)
	}
	defer f.Close()

	if err := rootCmd.GenFishCompletion(f, true); err != nil {
		return err
	}

	fmt.Printf("Installed fish completions to: %s\n", path)
	fmt.Println("Restart your shell to enable completions.")
	return nil
}
