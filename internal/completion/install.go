// Package completion provides shell completion helper functions.
package completion

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// GenerateAndInstall generates and installs completions for the specified shell.
func GenerateAndInstall(ios *iostreams.IOStreams, rootCmd *cobra.Command, shell string) error {
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
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}
	if err != nil {
		return fmt.Errorf("failed to generate completions: %w", err)
	}

	switch shell {
	case ShellBash:
		return InstallBash(ios, buf.Bytes())
	case ShellZsh:
		return InstallZsh(ios, buf.Bytes())
	case ShellFish:
		return InstallFish(ios, buf.Bytes())
	case ShellPowerShell:
		return InstallPowerShell(ios, buf.Bytes())
	}
	return nil
}

// InstallBash installs bash completions.
func InstallBash(ios *iostreams.IOStreams, script []byte) error {
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
	fs := config.Fs()
	userDir := filepath.Join(home, ".local", "share", "bash-completion", "completions")

	// Try system-wide completion directory on Linux
	if runtime.GOOS == "linux" {
		if _, statErr := fs.Stat("/etc/bash_completion.d"); statErr == nil {
			completionFile = "/etc/bash_completion.d/shelly"
			if writeErr := afero.WriteFile(fs, completionFile, script, 0o644); writeErr == nil {
				return "/etc/bash_completion.d", completionFile, nil
			}
			// Fall through to user directory on write failure
		}
	}

	// Use user directory
	if err = fs.MkdirAll(userDir, 0o755); err != nil {
		return "", "", fmt.Errorf("failed to create completion directory: %w", err)
	}

	completionFile = filepath.Join(userDir, "shelly")
	if err = afero.WriteFile(fs, completionFile, script, 0o644); err != nil {
		return "", "", fmt.Errorf("failed to write completion script: %w", err)
	}

	return userDir, completionFile, nil
}

// InstallZsh installs zsh completions.
func InstallZsh(ios *iostreams.IOStreams, script []byte) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	fs := config.Fs()

	// Use user's fpath-compatible directory
	completionDir := filepath.Join(home, ".zsh", "completions")
	if err := fs.MkdirAll(completionDir, 0o755); err != nil {
		return fmt.Errorf("failed to create completion directory: %w", err)
	}

	completionFile := filepath.Join(completionDir, "_shelly")

	// Write completion script
	if err := afero.WriteFile(fs, completionFile, script, 0o644); err != nil {
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

	f, err := config.Fs().OpenFile(rcFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
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

// InstallFish installs fish completions.
func InstallFish(ios *iostreams.IOStreams, script []byte) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	fs := config.Fs()

	// Fish completions go in ~/.config/fish/completions
	completionDir := filepath.Join(home, ".config", "fish", "completions")
	if err := fs.MkdirAll(completionDir, 0o755); err != nil {
		return fmt.Errorf("failed to create completion directory: %w", err)
	}

	completionFile := filepath.Join(completionDir, "shelly.fish")

	// Write completion script
	if err := afero.WriteFile(fs, completionFile, script, 0o644); err != nil {
		return fmt.Errorf("failed to write completion script: %w", err)
	}

	ios.Success("Completion script installed to: %s", completionFile)
	ios.Info("Fish will automatically load completions from this directory")
	ios.Success("Restart your shell or run: source %s", completionFile)
	return nil
}

// InstallPowerShell installs PowerShell completions.
func InstallPowerShell(ios *iostreams.IOStreams, script []byte) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	fs := config.Fs()

	// Determine PowerShell profile directory
	var profileDir string
	if runtime.GOOS == "windows" {
		profileDir = filepath.Join(home, "Documents", "WindowsPowerShell")
	} else {
		profileDir = filepath.Join(home, ".config", "powershell")
	}

	if err := fs.MkdirAll(profileDir, 0o755); err != nil {
		return fmt.Errorf("failed to create profile directory: %w", err)
	}

	completionFile := filepath.Join(profileDir, "shelly.ps1")

	// Write completion script
	if err := afero.WriteFile(fs, completionFile, script, 0o644); err != nil {
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

	f, err := config.Fs().OpenFile(profileFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
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
	f, err := config.Fs().Open(rcFile)
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
