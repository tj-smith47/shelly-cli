// Package current provides the theme current command.
package current

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// NewCommand creates the theme current command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "current",
		Aliases: []string{"cur", "c"},
		Short:   "Show current theme",
		Long:    `Show the currently active color theme.`,
		Example: `  # Show current theme
  shelly theme current

  # Output as JSON
  shelly theme current -o json`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return run(f)
		},
	}

	return cmd
}

func run(f *cmdutil.Factory) error {
	ios := f.IOStreams()

	current := theme.Current()
	if current == nil {
		return errors.New("no theme is currently set")
	}

	// Handle output format
	if output.WantsStructured() {
		data := map[string]any{
			"id":          current.ID,
			"name":        current.ID,
			"displayName": current.DisplayName,
		}
		return output.FormatOutput(ios.Out, data)
	}

	// Text output
	ios.Printf("Current theme: %s\n", current.ID)

	if current.DisplayName != "" && current.DisplayName != current.ID {
		ios.Printf("Display name: %s\n", current.DisplayName)
	}

	return nil
}
