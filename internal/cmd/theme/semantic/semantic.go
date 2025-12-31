// Package semantic provides the theme semantic command.
package semantic

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds the options for the semantic command.
type Options struct {
	Factory *cmdutil.Factory
}

// NewCommand creates the theme semantic command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "semantic",
		Aliases: []string{"sem", "colors"},
		Short:   "Show semantic color mappings",
		Long: `Show the current semantic color mappings for the active theme.

Semantic colors provide consistent meaning across the CLI:
- Primary/Secondary: Main UI colors
- Success/Warning/Error/Info: Feedback colors
- Online/Offline/Updating/Idle: Device state colors
- TableHeader/TableCell/TableAltCell/TableBorder: Table styling`,
		Example: `  # Show semantic colors
  shelly theme semantic

  # Output as JSON
  shelly theme semantic -o json`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return run(opts)
		},
	}

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()
	colors := theme.GetSemanticColors()

	// Handle structured output
	if output.WantsStructured() {
		data := map[string]string{
			"primary":        theme.ColorToHex(colors.Primary),
			"secondary":      theme.ColorToHex(colors.Secondary),
			"highlight":      theme.ColorToHex(colors.Highlight),
			"muted":          theme.ColorToHex(colors.Muted),
			"text":           theme.ColorToHex(colors.Text),
			"alt_text":       theme.ColorToHex(colors.AltText),
			"success":        theme.ColorToHex(colors.Success),
			"warning":        theme.ColorToHex(colors.Warning),
			"error":          theme.ColorToHex(colors.Error),
			"info":           theme.ColorToHex(colors.Info),
			"background":     theme.ColorToHex(colors.Background),
			"alt_background": theme.ColorToHex(colors.AltBackground),
			"online":         theme.ColorToHex(colors.Online),
			"offline":        theme.ColorToHex(colors.Offline),
			"updating":       theme.ColorToHex(colors.Updating),
			"idle":           theme.ColorToHex(colors.Idle),
			"table_header":   theme.ColorToHex(colors.TableHeader),
			"table_cell":     theme.ColorToHex(colors.TableCell),
			"table_alt_cell": theme.ColorToHex(colors.TableAltCell),
			"table_border":   theme.ColorToHex(colors.TableBorder),
		}
		return output.FormatOutput(ios.Out, data)
	}

	// Text output with color samples
	ios.Println("Semantic Color Mappings:")
	ios.Println("")

	// UI Colors
	ios.Println("UI Colors:")
	ios.Printf("  %s  Primary\n", theme.SemanticPrimary().Render("████"))
	ios.Printf("  %s  Secondary\n", theme.SemanticSecondary().Render("████"))
	ios.Printf("  %s  Highlight\n", theme.SemanticHighlight().Render("████"))
	ios.Printf("  %s  Muted\n", theme.SemanticMuted().Render("████"))
	ios.Println("")

	// Text Colors
	ios.Println("Text Colors:")
	ios.Printf("  %s  Text\n", theme.SemanticText().Render("████"))
	ios.Printf("  %s  AltText\n", theme.SemanticAltText().Render("████"))
	ios.Println("")

	// Feedback Colors
	ios.Println("Feedback Colors:")
	ios.Printf("  %s  Success\n", theme.SemanticSuccess().Render("████"))
	ios.Printf("  %s  Warning\n", theme.SemanticWarning().Render("████"))
	ios.Printf("  %s  Error\n", theme.SemanticError().Render("████"))
	ios.Printf("  %s  Info\n", theme.SemanticInfo().Render("████"))
	ios.Println("")

	// State Colors
	ios.Println("Device State Colors:")
	ios.Printf("  %s  Online\n", theme.SemanticOnline().Render("████"))
	ios.Printf("  %s  Offline\n", theme.SemanticOffline().Render("████"))
	ios.Printf("  %s  Updating\n", theme.SemanticUpdating().Render("████"))
	ios.Printf("  %s  Idle\n", theme.SemanticIdle().Render("████"))
	ios.Println("")

	// Table Colors
	ios.Println("Table Colors:")
	ios.Printf("  %s  TableHeader\n", theme.SemanticTableHeader().Render("████"))
	ios.Printf("  %s  TableCell\n", theme.SemanticTableCell().Render("████"))
	ios.Printf("  %s  TableAltCell\n", theme.SemanticTableAltCell().Render("████"))
	ios.Printf("  %s  TableBorder\n", theme.SemanticTableBorder().Render("████"))

	return nil
}
