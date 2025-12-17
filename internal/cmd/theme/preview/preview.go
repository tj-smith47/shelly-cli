// Package preview provides the theme preview command.
package preview

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// NewCommand creates the theme preview command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
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
			themeName := ""
			if len(args) > 0 {
				themeName = args[0]
			}
			return run(f, themeName)
		},
	}

	return cmd
}

func run(f *cmdutil.Factory, themeName string) error {
	ios := f.IOStreams()

	// Save current theme to restore later
	current := theme.Current()
	currentID := ""
	if current != nil {
		currentID = current.ID
	}

	// Set preview theme if specified
	if themeName != "" {
		if _, ok := theme.GetTheme(themeName); !ok {
			return fmt.Errorf("theme not found: %s", themeName)
		}
		theme.SetTheme(themeName)
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

	// Build preview output using the theme colors
	var b strings.Builder

	// Header
	title := lipgloss.NewStyle().Bold(true).Foreground(theme.Purple())
	b.WriteString(title.Render(fmt.Sprintf("Theme: %s", previewTheme.ID)))
	b.WriteString("\n\n")

	// Color palette
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(theme.Fg()).Render("Color Palette:"))
	b.WriteString("\n")

	colors := []struct {
		name  string
		color func() color.Color
	}{
		{"Foreground", theme.Fg},
		{"Background", theme.Bg},
		{"Red", theme.Red},
		{"Green", theme.Green},
		{"Yellow", theme.Yellow},
		{"Blue", theme.Blue},
		{"Purple", theme.Purple},
		{"Cyan", theme.Cyan},
	}

	for _, c := range colors {
		style := lipgloss.NewStyle().Foreground(c.color())
		block := lipgloss.NewStyle().Background(c.color()).Render("    ")
		b.WriteString(fmt.Sprintf("  %s %s\n", block, style.Render(c.name)))
	}

	// Status indicators
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(theme.Fg()).Render("Status Indicators:"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  %s  Success message\n", theme.StatusOK().Render("✓")))
	b.WriteString(fmt.Sprintf("  %s  Warning message\n", theme.StatusWarn().Render("!")))
	b.WriteString(fmt.Sprintf("  %s  Error message\n", theme.StatusError().Render("✗")))
	b.WriteString(fmt.Sprintf("  %s  Info message\n", theme.StatusInfo().Render("i")))

	// Device status
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(theme.Fg()).Render("Device Status:"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  %s\n", theme.DeviceOnline()))
	b.WriteString(fmt.Sprintf("  %s\n", theme.DeviceOffline()))
	b.WriteString(fmt.Sprintf("  %s\n", theme.DeviceUpdating()))

	// Switch state
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(theme.Fg()).Render("Switch State:"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  %s  %s\n", theme.SwitchOn(), theme.SwitchOff()))

	// Power/Energy
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(theme.Fg()).Render("Power/Energy:"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  %s  %s\n", theme.StyledPower(1234.5), theme.StyledEnergy(5678.9)))

	// Table sample
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(theme.Fg()).Render("Table Header:"))
	b.WriteString("\n")
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Cyan())
	dimStyle := lipgloss.NewStyle().Foreground(theme.BrightBlack())
	b.WriteString(fmt.Sprintf("  %s  %s  %s\n",
		headerStyle.Render("DEVICE"),
		headerStyle.Render("STATUS"),
		headerStyle.Render("POWER")))
	b.WriteString(fmt.Sprintf("  %s  %s  %s\n",
		lipgloss.NewStyle().Foreground(theme.Fg()).Render("kitchen"),
		theme.StatusOnline().Render("online"),
		theme.StyledPower(45.2)))
	b.WriteString(fmt.Sprintf("  %s  %s  %s\n",
		dimStyle.Render("bedroom"),
		theme.StatusOffline().Render("offline"),
		dimStyle.Render("--")))

	ios.Printf("%s", b.String())
	return nil
}
