// Package importcmd provides the theme import command.
package importcmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/term"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// NewCommand creates the theme import command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var apply bool

	cmd := &cobra.Command{
		Use:     "import <file>",
		Aliases: []string{"load"},
		Short:   "Import theme from file",
		Long: `Import a theme configuration from a file.

Supports importing theme files that reference any of the 280+ built-in themes.`,
		Example: `  # Import and apply a theme
  shelly theme import mytheme.yaml --apply

  # Just validate the file
  shelly theme import mytheme.yaml`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(f, args[0], apply)
		},
	}

	cmd.Flags().BoolVar(&apply, "apply", false, "Apply the imported theme")

	return cmd
}

func run(f *cmdutil.Factory, file string, apply bool) error {
	ios := f.IOStreams()

	// Read the file
	data, err := os.ReadFile(file) //nolint:gosec // G304: file path is user-provided, expected behavior
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse the theme
	var imported theme.Import
	if err := yaml.Unmarshal(data, &imported); err != nil {
		return fmt.Errorf("failed to parse theme file: %w", err)
	}

	// Determine theme name (new format: name, old format: id)
	themeName := imported.Name
	if themeName == "" {
		themeName = imported.ID
	}

	// Validate: need either a theme name or custom colors
	if themeName == "" && len(imported.Colors) == 0 {
		return fmt.Errorf("invalid theme file: missing 'name' or 'colors' field")
	}

	// Check if base theme exists (if specified)
	if themeName != "" {
		if _, ok := theme.GetTheme(themeName); !ok {
			ios.Warning("Theme '%s' is not a built-in theme", themeName)
			ios.Info("Run 'shelly theme list' to see available themes.")
			return fmt.Errorf("theme '%s' not found in built-in themes", themeName)
		}
	}

	if apply {
		return term.ApplyImportedTheme(ios, themeName, imported.Colors)
	}

	term.DisplayValidationResult(ios, themeName, imported.Colors)
	return nil
}
