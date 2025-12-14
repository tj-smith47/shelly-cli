// Package set provides the theme set command.
package set

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// NewCommand creates the theme set command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var save bool

	cmd := &cobra.Command{
		Use:     "set <theme>",
		Aliases: []string{"use", "s"},
		Short:   "Set the current theme",
		Long: `Set the current CLI color theme.

Use --save to persist the theme to your configuration file. Without --save,
the theme is only applied for the current session.`,
		Example: `  # Set theme for current session
  shelly theme set dracula

  # Set and save to config
  shelly theme set nord --save`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteThemeNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(f, args[0], save)
		},
	}

	cmd.Flags().BoolVar(&save, "save", false, "Save theme to configuration file")

	return cmd
}

func run(f *cmdutil.Factory, themeName string, save bool) error {
	ios := f.IOStreams()

	// Check if theme exists
	if _, ok := theme.GetTheme(themeName); !ok {
		return fmt.Errorf("theme not found: %s\nRun 'shelly theme list' to see available themes", themeName)
	}

	// Set the theme
	if !theme.SetTheme(themeName) {
		return fmt.Errorf("failed to set theme: %s", themeName)
	}

	// Save to config if requested
	if save {
		if err := saveThemeToConfig(themeName); err != nil {
			return fmt.Errorf("failed to save theme to config: %w", err)
		}
		ios.Success("Theme set to '%s' and saved to config", themeName)
	} else {
		ios.Success("Theme set to '%s'", themeName)
	}

	return nil
}

func saveThemeToConfig(themeName string) error {
	viper.Set("theme", themeName)

	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		// Create default config path
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		configDir := filepath.Join(home, ".config", "shelly")
		if err := os.MkdirAll(configDir, 0o700); err != nil {
			return err
		}
		configFile = filepath.Join(configDir, "config.yaml")
	}

	return viper.WriteConfigAs(configFile)
}
