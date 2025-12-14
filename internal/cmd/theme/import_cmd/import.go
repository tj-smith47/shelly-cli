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
type ThemeImport struct {
	ID          string `yaml:"id" json:"id"`
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

	if imported.ID == "" {
		return fmt.Errorf("invalid theme file: missing 'id' field")
	}

	// Check if this is a built-in theme
	if _, ok := theme.GetTheme(imported.ID); !ok {
		ios.Warning("Theme '%s' is not a built-in theme", imported.ID)
		ios.Info("Custom themes are not yet supported. Please use a built-in theme ID.")
		ios.Printf("\nRun 'shelly theme list' to see available themes.\n")
		return fmt.Errorf("theme '%s' not found in built-in themes", imported.ID)
	}

	if apply {
		if !theme.SetTheme(imported.ID) {
			return fmt.Errorf("failed to apply theme: %s", imported.ID)
		}
		ios.Success("Theme '%s' imported and applied", imported.ID)
	} else {
		ios.Success("Theme file validated: '%s' is a valid built-in theme", imported.ID)
		ios.Info("Use --apply to apply the theme")
	}

	return nil
}
