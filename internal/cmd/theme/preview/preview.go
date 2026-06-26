// Package preview provides the theme preview command.
package preview

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/term"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds the options for the preview command.
type Options struct {
	Factory   *cmdutil.Factory
	ThemeName string
}

// NewCommand creates the theme preview command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "preview [theme]",
		Aliases: []string{"show", "demo"},
		Short:   "Preview a theme",
		Long: `Preview a theme by showing sample output.

If no theme is specified, previews the current theme.`,
		Example: `  # Preview a specific theme
  shelly theme preview nord

  # Preview current theme
  shelly theme preview`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.ThemeName = args[0]
			}
			return run(opts)
		},
	}

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	// Save current theme to restore later
	current := theme.Current()
	currentID := ""
	if current != nil {
		currentID = current.ID
	}

	// Set preview theme if specified
	if opts.ThemeName != "" {
		if _, ok := theme.GetTheme(opts.ThemeName); !ok {
			return fmt.Errorf("theme not found: %s", opts.ThemeName)
		}
		theme.SetTheme(opts.ThemeName)
		defer func() {
			if currentID != "" {
				theme.SetTheme(currentID)
			}
		}()
	}

	// Get the theme being previewed
	previewTheme := theme.Current()
	if previewTheme == nil {
		return fmt.Errorf("no theme available")
	}

	ios.Printf("%s", term.RenderThemePreview(previewTheme.ID))
	return nil
}
