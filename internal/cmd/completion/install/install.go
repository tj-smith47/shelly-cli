// Package install provides the completion install command.
package install

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// Shell type constants.
const (
	ShellBash       = "bash"
	ShellZsh        = "zsh"
	ShellFish       = "fish"
	ShellPowerShell = "powershell"
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
		RunE: func(cmd *cobra.Command, args []string) error {
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
		shell, err = DetectShell()
		if err != nil {
			return fmt.Errorf("could not detect shell: %w\nSpecify shell with --shell flag", err)
		}
	}

	// Validate shell
	switch shell {
	case ShellBash, ShellZsh, ShellFish, ShellPowerShell:
		// Valid
	default:
		return fmt.Errorf("unsupported shell: %s\nSupported: bash, zsh, fish, powershell", shell)
	}

	ios.Info("Detected shell: %s", shell)

	// Generate completion script
	var buf bytes.Buffer
	var err error

	switch shell {
	case ShellBash:
		err = rootCmd.GenBashCompletion(&buf)
	case ShellZsh:
		err = rootCmd.GenZshCompletion(&buf)
	case ShellFish:
		err = rootCmd.GenFishCompletion(&buf, true)
	case ShellPowerShell:
		err = rootCmd.GenPowerShellCompletionWithDesc(&buf)
	}
	if err != nil {
		return fmt.Errorf("failed to generate completion script: %w", err)
	}

	// Install based on shell
	switch shell {
	case ShellBash:
		return Bash(ios, buf.Bytes())
	case ShellZsh:
		return Zsh(ios, buf.Bytes())
	case ShellFish:
		return Fish(ios, buf.Bytes())
	case ShellPowerShell:
		return PowerShell(ios, buf.Bytes())
	}

	return nil
}

// DetectShell attempts to detect the user's shell.
func DetectShell() (string, error) {
	// Check SHELL environment variable
	shell := os.Getenv("SHELL")
	if shell != "" {
		base := filepath.Base(shell)
		switch base {
		case "bash":
			return ShellBash, nil
		case "zsh":
			return ShellZsh, nil
		case "fish":
			return ShellFish, nil
		case "pwsh", "powershell":
			return ShellPowerShell, nil
		}
	}

	// On Windows, check for PowerShell
	if runtime.GOOS == "windows" {
		return ShellPowerShell, nil
	}

	// Try to get parent process name
	if ppid := os.Getppid(); ppid > 0 {
		if name, err := getProcessName(ppid); err == nil {
			switch filepath.Base(name) {
			case "bash":
				return ShellBash, nil
			case "zsh":
				return ShellZsh, nil
			case "fish":
				return ShellFish, nil
			case "pwsh", "powershell":
				return ShellPowerShell, nil
			}
		}
	}

	return "", fmt.Errorf("could not detect shell from environment")
}

// getProcessName gets the name of a process by PID.
func getProcessName(pid int) (string, error) {
	// Try /proc on Linux
	if runtime.GOOS == "linux" {
		cmdline, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
		if err == nil && len(cmdline) > 0 {
			// cmdline is null-separated, get first part
			parts := strings.SplitN(string(cmdline), "\x00", 2)
			if len(parts) > 0 {
				return parts[0], nil
			}
		}
	}

	// Try ps command as fallback using a short timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	//nolint:gosec // G204: pid is from os.Getppid() which is safe
	out, err := exec.CommandContext(ctx, "ps", "-p", fmt.Sprintf("%d", pid), "-o", "comm=").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// Bash installs bash completions.
func Bash(ios *iostreams.IOStreams, script []byte) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	completionDir, completionFile, err := writeBashCompletion(home, script)
	if err != nil {
		return err
	}

	ios.Success("Completion script installed to: %s", completionFile)

	// Check if we need to add source line to .bashrc
	userCompletionDir := filepath.Join(home, ".local", "share", "bash-completion", "completions")
	if completionDir == userCompletionDir {
		rcFile := filepath.Join(home, ".bashrc")
		sourceLine := fmt.Sprintf("source %s", completionFile)
		if needsSourceLine(rcFile, completionFile) {
			ios.Info("Add the following to your %s to enable completions:", rcFile)
			ios.Printf("\n  %s\n\n", sourceLine)
		}
	} else {
		ios.Info("Completions will be loaded automatically from %s", completionDir)
	}

	ios.Success("Restart your shell or run: source %s", completionFile)
	return nil
}

// writeBashCompletion writes the bash completion script to the appropriate location.
// It tries the system directory first, then falls back to user directory.
func writeBashCompletion(home string, script []byte) (completionDir, completionFile string, err error) {
	userDir := filepath.Join(home, ".local", "share", "bash-completion", "completions")

	// Try system-wide completion directory on Linux
	if runtime.GOOS == "linux" {
		if _, statErr := os.Stat("/etc/bash_completion.d"); statErr == nil {
			completionFile = "/etc/bash_completion.d/shelly"
			//nolint:gosec // G306: 0644 is required for shell to source completion files
			if writeErr := os.WriteFile(completionFile, script, 0o644); writeErr == nil {
				return "/etc/bash_completion.d", completionFile, nil
			}
			// Fall through to user directory on write failure
		}
	}

	// Use user directory
	//nolint:gosec // G301: 0755 is required for directories to be traversable
	if err = os.MkdirAll(userDir, 0o755); err != nil {
		return "", "", fmt.Errorf("failed to create completion directory: %w", err)
	}

	completionFile = filepath.Join(userDir, "shelly")
	//nolint:gosec // G306: 0644 is required for shell to source completion files
	if err = os.WriteFile(completionFile, script, 0o644); err != nil {
		return "", "", fmt.Errorf("failed to write completion script: %w", err)
	}

	return userDir, completionFile, nil
}

// Zsh installs zsh completions.
func Zsh(ios *iostreams.IOStreams, script []byte) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Use user's fpath-compatible directory
	completionDir := filepath.Join(home, ".zsh", "completions")
	//nolint:gosec // G301: 0755 is required for directories to be traversable
	if err := os.MkdirAll(completionDir, 0o755); err != nil {
		return fmt.Errorf("failed to create completion directory: %w", err)
	}

	completionFile := filepath.Join(completionDir, "_shelly")

	// Write completion script
	//nolint:gosec // G306: 0644 is required for shell to source completion files
	if err := os.WriteFile(completionFile, script, 0o644); err != nil {
		return fmt.Errorf("failed to write completion script: %w", err)
	}

	ios.Success("Completion script installed to: %s", completionFile)

	// Check .zshrc for fpath setup
	rcFile := filepath.Join(home, ".zshrc")

	if needsSourceLine(rcFile, completionDir) {
		if err := updateZshRC(ios, rcFile, completionDir); err != nil {
			ios.Warning("Could not update %s: %v", rcFile, err)
			printZshInstructions(ios, rcFile, completionDir)
		}
	} else {
		ios.Info("fpath already configured in %s", rcFile)
	}

	ios.Success("Restart your shell to enable completions")
	return nil
}

// updateZshRC adds fpath and compinit to .zshrc.
func updateZshRC(ios *iostreams.IOStreams, rcFile, completionDir string) error {
	fpathLine := fmt.Sprintf("fpath=(%s $fpath)", completionDir)
	additions := []string{
		"",
		"# Shelly CLI completions",
		fpathLine,
		"autoload -Uz compinit && compinit",
	}

	//nolint:gosec // G302: 0644 is standard for shell rc files
	f, err := os.OpenFile(rcFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer iostreams.CloseWithDebug("close zshrc", f)

	for _, line := range additions {
		if _, err := fmt.Fprintln(f, line); err != nil {
			return err
		}
	}

	ios.Success("Updated %s with completion setup", rcFile)
	return nil
}

// printZshInstructions prints manual zsh setup instructions.
func printZshInstructions(ios *iostreams.IOStreams, rcFile, completionDir string) {
	fpathLine := fmt.Sprintf("fpath=(%s $fpath)", completionDir)
	ios.Info("Add the following to your %s:", rcFile)
	ios.Println("")
	ios.Println("  # Shelly CLI completions")
	ios.Printf("  %s\n", fpathLine)
	ios.Println("  autoload -Uz compinit && compinit")
}

// Fish installs fish completions.
func Fish(ios *iostreams.IOStreams, script []byte) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Fish completions go in ~/.config/fish/completions
	completionDir := filepath.Join(home, ".config", "fish", "completions")
	//nolint:gosec // G301: 0755 is required for directories to be traversable
	if err := os.MkdirAll(completionDir, 0o755); err != nil {
		return fmt.Errorf("failed to create completion directory: %w", err)
	}

	completionFile := filepath.Join(completionDir, "shelly.fish")

	// Write completion script
	//nolint:gosec // G306: 0644 is required for shell to source completion files
	if err := os.WriteFile(completionFile, script, 0o644); err != nil {
		return fmt.Errorf("failed to write completion script: %w", err)
	}

	ios.Success("Completion script installed to: %s", completionFile)
	ios.Info("Fish will automatically load completions from this directory")
	ios.Success("Restart your shell or run: source %s", completionFile)
	return nil
}

// PowerShell installs PowerShell completions.
func PowerShell(ios *iostreams.IOStreams, script []byte) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Determine PowerShell profile directory
	var profileDir string
	if runtime.GOOS == "windows" {
		profileDir = filepath.Join(home, "Documents", "WindowsPowerShell")
	} else {
		profileDir = filepath.Join(home, ".config", "powershell")
	}

	//nolint:gosec // G301: 0755 is required for directories to be traversable
	if err := os.MkdirAll(profileDir, 0o755); err != nil {
		return fmt.Errorf("failed to create profile directory: %w", err)
	}

	completionFile := filepath.Join(profileDir, "shelly.ps1")

	// Write completion script
	//nolint:gosec // G306: 0644 is required for shell to source completion files
	if err := os.WriteFile(completionFile, script, 0o644); err != nil {
		return fmt.Errorf("failed to write completion script: %w", err)
	}

	ios.Success("Completion script installed to: %s", completionFile)

	// Check profile for sourcing
	profileFile := filepath.Join(profileDir, "Microsoft.PowerShell_profile.ps1")

	if needsSourceLine(profileFile, completionFile) {
		if err := updatePowerShellProfile(ios, profileFile, completionFile); err != nil {
			ios.Warning("Could not update %s: %v", profileFile, err)
			ios.Info("Add the following to your PowerShell profile:")
			ios.Printf("\n  . %s\n\n", completionFile)
		}
	}

	ios.Success("Restart your PowerShell session to enable completions")
	return nil
}

// updatePowerShellProfile adds sourcing to PowerShell profile.
func updatePowerShellProfile(ios *iostreams.IOStreams, profileFile, completionFile string) error {
	sourceLine := fmt.Sprintf(". %s", completionFile)
	additions := []string{
		"",
		"# Shelly CLI completions",
		sourceLine,
	}

	//nolint:gosec // G302: 0644 is standard for PowerShell profile files
	f, err := os.OpenFile(profileFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer iostreams.CloseWithDebug("close powershell profile", f)

	for _, line := range additions {
		if _, err := fmt.Fprintln(f, line); err != nil {
			return err
		}
	}

	ios.Success("Updated %s with completion setup", profileFile)
	return nil
}

// needsSourceLine checks if a source line needs to be added to an rc file.
func needsSourceLine(rcFile, searchStr string) bool {
	//nolint:gosec // G304: rcFile is from known paths (home directory + constant)
	f, err := os.Open(rcFile)
	if err != nil {
		// File doesn't exist, need to add line
		return true
	}
	defer iostreams.CloseWithDebug("close rc file", f)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), searchStr) {
			return false
		}
	}
	return true
}
