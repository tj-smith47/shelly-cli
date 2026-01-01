// Package set provides the theme set command.
package set

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds the options for the set command.
type Options struct {
	Factory   *cmdutil.Factory
	Save      bool
	ThemeName string
}

// NewCommand creates the theme set command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

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
		ValidArgsFunction: completion.ThemeNames(),
		RunE: func(_ *cobra.Command, args []string) error {
			opts.ThemeName = args[0]
			return run(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Save, "save", false, "Save theme to configuration file")

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	// Check if theme exists
	if _, ok := theme.GetTheme(opts.ThemeName); !ok {
		return fmt.Errorf("theme not found: %s\nRun 'shelly theme list' to see available themes", opts.ThemeName)
	}

	// Set the theme
	if !theme.SetTheme(opts.ThemeName) {
		return fmt.Errorf("failed to set theme: %s", opts.ThemeName)
	}

	// Save to config if requested
	if opts.Save {
		if err := config.SaveTheme(opts.ThemeName); err != nil {
			return fmt.Errorf("failed to save theme to config: %w", err)
		}
		ios.Success("Theme set to '%s' and saved to config", opts.ThemeName)
	} else {
		ios.Success("Theme set to '%s'", opts.ThemeName)
	}

	return nil
}
