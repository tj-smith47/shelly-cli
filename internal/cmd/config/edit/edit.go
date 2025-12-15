// Package edit provides the config edit command.
package edit

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// NewCommand creates the config edit command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "edit",
		Aliases: []string{"e"},
		Short:   "Open CLI config in editor",
		Long: `Open the Shelly CLI configuration file in your default editor.

The editor is determined by:
  1. $EDITOR environment variable
  2. $VISUAL environment variable
  3. Falls back to 'vi' on Unix, 'notepad' on Windows

This allows you to directly edit the configuration file, including
devices, aliases, groups, scenes, and other settings.`,
		Example: `  # Open config in default editor
  shelly config edit

  # Set EDITOR and open
  EDITOR=nano shelly config edit`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f)
		},
	}

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory) error {
	ios := f.IOStreams()

	// Get config file path
	configDir, err := config.Dir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}
	configPath := configDir + "/config.yaml"

	// Check if config exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s\nRun 'shelly init' to create it", configPath)
	}

	// Find editor
	editor := getEditor()
	if editor == "" {
		return fmt.Errorf("no editor found. Set $EDITOR or $VISUAL environment variable")
	}

	ios.Info("Opening %s with %s...", configPath, editor)

	// Execute editor
	editorCmd := exec.CommandContext(ctx, editor, configPath) //nolint:gosec // User-specified editor is intentional
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	if err := editorCmd.Run(); err != nil {
		return fmt.Errorf("editor failed: %w", err)
	}

	ios.Success("Config file edited")

	// Reload config to validate
	if _, err := config.Reload(); err != nil {
		ios.Warning("Config may have errors: %v", err)
		ios.Info("Run 'shelly doctor' to check configuration")
	}

	return nil
}

func getEditor() string {
	// Check EDITOR first
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}

	// Then VISUAL
	if visual := os.Getenv("VISUAL"); visual != "" {
		return visual
	}

	// Platform defaults
	// Try common editors in order
	editors := []string{"nano", "vim", "vi"}
	for _, e := range editors {
		if path, err := exec.LookPath(e); err == nil {
			return path
		}
	}

	return ""
}
