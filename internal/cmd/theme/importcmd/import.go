// Package importcmd provides the theme import command.
package importcmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// ThemeImport represents an imported theme configuration.
// Supports both the old format (id only) and new format (name + colors).
type ThemeImport struct {
	// New format fields
	Name   string            `yaml:"name" json:"name,omitempty"`
	Colors map[string]string `yaml:"colors" json:"colors,omitempty"`

	// Old format fields (backwards compatible)
	ID          string `yaml:"id" json:"id,omitempty"`
	DisplayName string `yaml:"display_name,omitempty" json:"display_name,omitempty"`
}

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
	var imported ThemeImport
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
		if err := theme.ApplyConfig(themeName, imported.Colors, ""); err != nil {
			return fmt.Errorf("failed to apply theme: %w", err)
		}
		if len(imported.Colors) > 0 {
			ios.Success("Theme '%s' with %d color overrides imported and applied", themeName, len(imported.Colors))
		} else {
			ios.Success("Theme '%s' imported and applied", themeName)
		}
	} else {
		ios.Success("Theme file validated successfully")
		if themeName != "" {
			ios.Info("Base theme: %s", themeName)
		}
		if len(imported.Colors) > 0 {
			ios.Info("Color overrides: %d", len(imported.Colors))
		}
		ios.Info("Use --apply to apply the theme")
	}

	return nil
}
