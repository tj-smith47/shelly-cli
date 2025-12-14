// Package list provides the theme list command.
package list

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// NewCommand creates the theme list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var filter string

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "List available themes",
		Long: `List all available color themes.

The CLI includes 280+ themes from bubbletint. Use --filter to search
for themes by name.`,
		Example: `  # List all themes
  shelly theme list

  # Filter themes by name
  shelly theme list --filter dark

  # Output as JSON
  shelly theme list -o json`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return run(f, filter)
		},
	}

	cmd.Flags().StringVar(&filter, "filter", "", "Filter themes by name pattern")

	return cmd
}

func run(f *cmdutil.Factory, filter string) error {
	ios := f.IOStreams()

	themes := theme.ListThemes()
	current := theme.Current()
	currentID := ""
	if current != nil {
		currentID = current.ID
	}

	// Filter if specified
	if filter != "" {
		filter = strings.ToLower(filter)
		var filtered []string
		for _, t := range themes {
			if strings.Contains(strings.ToLower(t), filter) {
				filtered = append(filtered, t)
			}
		}
		themes = filtered
	}

	// Handle output format
	if cmdutil.WantsStructured() {
		data := make([]map[string]any, len(themes))
		for i, t := range themes {
			data[i] = map[string]any{
				"id":      t,
				"current": t == currentID,
			}
		}
		return cmdutil.FormatOutput(ios, data)
	}

	// Text output
	if len(themes) == 0 {
		ios.Info("No themes found matching filter: %s", filter)
		return nil
	}

	ios.Printf("Available Themes (%d themes)\n\n", len(themes))

	// Display in columns for better readability
	table := output.NewTable("Theme", "Current")
	for _, t := range themes {
		isCurrent := ""
		if t == currentID {
			isCurrent = "✓"
		}
		table.AddRow(t, isCurrent)
	}
	table.Print()

	return nil
}
